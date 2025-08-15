import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/status.dart' as status;
import 'package:logger/logger.dart';
import 'package:connectivity_plus/connectivity_plus.dart';
import '../config/app_config.dart';
import '../utils/app_exceptions.dart';

class WebSocketService {
  static WebSocketService? _instance;
  final Logger _logger = Logger();
  
  WebSocketChannel? _channel;
  String? _authToken;
  bool _isConnected = false;
  bool _isConnecting = false;
  bool _shouldReconnect = true;
  int _reconnectAttempts = 0;
  Timer? _reconnectTimer;
  Timer? _heartbeatTimer;
  Timer? _connectionTimeoutTimer;
  
  // Event streams
  final StreamController<WebSocketMessage> _messageController = StreamController<WebSocketMessage>.broadcast();
  final StreamController<WebSocketConnectionState> _connectionStateController = StreamController<WebSocketConnectionState>.broadcast();
  final StreamController<String> _errorController = StreamController<String>.broadcast();
  
  // Message handlers
  final Map<String, List<Function(Map<String, dynamic>)>> _messageHandlers = {};
  
  // Configuration
  static const int _maxReconnectAttempts = 5;
  static const Duration _initialReconnectDelay = Duration(seconds: 1);
  static const Duration _maxReconnectDelay = Duration(seconds: 30);
  static const Duration _heartbeatInterval = Duration(seconds: 30);
  static const Duration _connectionTimeout = Duration(seconds: 10);

  WebSocketService._internal();

  factory WebSocketService() {
    _instance ??= WebSocketService._internal();
    return _instance!;
  }

  // Streams
  Stream<WebSocketMessage> get messageStream => _messageController.stream;
  Stream<WebSocketConnectionState> get connectionStateStream => _connectionStateController.stream;
  Stream<String> get errorStream => _errorController.stream;
  
  // Getters
  bool get isConnected => _isConnected;
  bool get isConnecting => _isConnecting;
  int get reconnectAttempts => _reconnectAttempts;

  /// Connects to the WebSocket server
  Future<void> connect({String? authToken}) async {
    if (_isConnected || _isConnecting) {
      _logger.w('WebSocket already connected or connecting');
      return;
    }

    _authToken = authToken;
    _shouldReconnect = true;
    await _performConnection();
  }

  /// Disconnects from the WebSocket server
  Future<void> disconnect() async {
    _shouldReconnect = false;
    _reconnectAttempts = 0;
    
    _cancelTimers();
    
    if (_channel != null) {
      await _channel!.sink.close(status.goingAway);
      _channel = null;
    }
    
    _setConnectionState(WebSocketConnectionState.disconnected);
    _logger.i('WebSocket disconnected');
  }

  /// Sends a message through the WebSocket
  void sendMessage(WebSocketMessage message) {
    if (!_isConnected) {
      _logger.w('Cannot send message: WebSocket not connected');
      _errorController.add('Cannot send message: Not connected');
      return;
    }

    try {
      final String jsonMessage = jsonEncode(message.toJson());
      _channel!.sink.add(jsonMessage);
      _logger.d('Message sent: ${message.type}');
    } catch (e) {
      _logger.e('Failed to send message: $e');
      _errorController.add('Failed to send message: $e');
    }
  }

  /// Subscribes to specific message types
  void subscribe(String messageType, Function(Map<String, dynamic>) handler) {
    _messageHandlers.putIfAbsent(messageType, () => []).add(handler);
    _logger.d('Subscribed to message type: $messageType');
  }

  /// Unsubscribes from specific message types
  void unsubscribe(String messageType, Function(Map<String, dynamic>) handler) {
    _messageHandlers[messageType]?.remove(handler);
    if (_messageHandlers[messageType]?.isEmpty == true) {
      _messageHandlers.remove(messageType);
    }
    _logger.d('Unsubscribed from message type: $messageType');
  }

  /// Performs the actual WebSocket connection
  Future<void> _performConnection() async {
    if (_isConnecting) return;
    
    _isConnecting = true;
    _setConnectionState(WebSocketConnectionState.connecting);
    
    try {
      // Check network connectivity
      final connectivityResult = await Connectivity().checkConnectivity();
      if (connectivityResult == ConnectivityResult.none) {
        throw WebSocketException('No network connection');
      }

      final Uri wsUri = _buildWebSocketUri();
      _logger.i('Connecting to WebSocket: ${wsUri.toString()}');

      // Set connection timeout
      _connectionTimeoutTimer = Timer(_connectionTimeout, () {
        if (_isConnecting) {
          _logger.w('WebSocket connection timeout');
          _handleConnectionError('Connection timeout');
        }
      });

      _channel = WebSocketChannel.connect(wsUri);
      
      // Listen to the WebSocket stream
      _channel!.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDisconnected,
        cancelOnError: false,
      );

      // Send authentication if token is available
      if (_authToken != null) {
        _sendAuthMessage();
      }

      _isConnecting = false;
      _isConnected = true;
      _reconnectAttempts = 0;
      _connectionTimeoutTimer?.cancel();
      
      _setConnectionState(WebSocketConnectionState.connected);
      _startHeartbeat();
      
      _logger.i('WebSocket connected successfully');
    } catch (e) {
      _isConnecting = false;
      _connectionTimeoutTimer?.cancel();
      _handleConnectionError(e.toString());
    }
  }

  /// Builds the WebSocket URI with authentication
  Uri _buildWebSocketUri() {
    final String baseUrl = AppConfig.websocketUrl;
    final Uri uri = Uri.parse(baseUrl);
    
    final Map<String, String> queryParams = {};
    
    if (_authToken != null) {
      queryParams['token'] = _authToken!;
    }
    
    return uri.replace(queryParameters: queryParams);
  }

  /// Sends authentication message
  void _sendAuthMessage() {
    final authMessage = WebSocketMessage(
      type: 'auth',
      data: {'token': _authToken},
      timestamp: DateTime.now(),
    );
    sendMessage(authMessage);
  }

  /// Handles incoming messages
  void _onMessage(dynamic message) {
    try {
      final Map<String, dynamic> messageData = jsonDecode(message as String);
      final WebSocketMessage wsMessage = WebSocketMessage.fromJson(messageData);
      
      _logger.d('Message received: ${wsMessage.type}');
      
      // Handle system messages
      if (wsMessage.type == 'pong') {
        _logger.d('Pong received');
        return;
      }
      
      if (wsMessage.type == 'auth_success') {
        _logger.i('Authentication successful');
        return;
      }
      
      if (wsMessage.type == 'auth_failed') {
        _logger.e('Authentication failed');
        _errorController.add('Authentication failed');
        return;
      }
      
      // Broadcast message to subscribers
      _messageController.add(wsMessage);
      
      // Call specific handlers
      final handlers = _messageHandlers[wsMessage.type];
      if (handlers != null) {
        for (final handler in handlers) {
          try {
            handler(wsMessage.data);
          } catch (e) {
            _logger.e('Error in message handler: $e');
          }
        }
      }
    } catch (e) {
      _logger.e('Error processing message: $e');
      _errorController.add('Error processing message: $e');
    }
  }

  /// Handles WebSocket errors
  void _onError(dynamic error) {
    _logger.e('WebSocket error: $error');
    _errorController.add('WebSocket error: $error');
    _handleConnectionError(error.toString());
  }

  /// Handles WebSocket disconnection
  void _onDisconnected() {
    _logger.w('WebSocket disconnected');
    _isConnected = false;
    _isConnecting = false;
    _cancelTimers();
    
    _setConnectionState(WebSocketConnectionState.disconnected);
    
    if (_shouldReconnect) {
      _scheduleReconnect();
    }
  }

  /// Handles connection errors and triggers reconnection
  void _handleConnectionError(String error) {
    _logger.e('Connection error: $error');
    _isConnected = false;
    _isConnecting = false;
    
    _setConnectionState(WebSocketConnectionState.error);
    _errorController.add(error);
    
    if (_shouldReconnect) {
      _scheduleReconnect();
    }
  }

  /// Schedules a reconnection attempt
  void _scheduleReconnect() {
    if (_reconnectAttempts >= _maxReconnectAttempts) {
      _logger.e('Max reconnection attempts reached');
      _setConnectionState(WebSocketConnectionState.failed);
      _shouldReconnect = false;
      return;
    }

    _reconnectAttempts++;
    
    // Calculate delay with exponential backoff
    final int delaySeconds = (_initialReconnectDelay.inSeconds * 
        (1 << (_reconnectAttempts - 1))).clamp(
        _initialReconnectDelay.inSeconds, 
        _maxReconnectDelay.inSeconds
    );
    
    final Duration delay = Duration(seconds: delaySeconds);
    
    _logger.i('Scheduling reconnection attempt $_reconnectAttempts in ${delay.inSeconds}s');
    
    _reconnectTimer = Timer(delay, () {
      if (_shouldReconnect) {
        _performConnection();
      }
    });
  }

  /// Starts the heartbeat mechanism
  void _startHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = Timer.periodic(_heartbeatInterval, (timer) {
      if (_isConnected) {
        _sendPing();
      } else {
        timer.cancel();
      }
    });
  }

  /// Sends a ping message
  void _sendPing() {
    final pingMessage = WebSocketMessage(
      type: 'ping',
      data: {'timestamp': DateTime.now().millisecondsSinceEpoch},
      timestamp: DateTime.now(),
    );
    sendMessage(pingMessage);
  }

  /// Sets the connection state and notifies listeners
  void _setConnectionState(WebSocketConnectionState state) {
    _connectionStateController.add(state);
  }

  /// Cancels all timers
  void _cancelTimers() {
    _reconnectTimer?.cancel();
    _heartbeatTimer?.cancel();
    _connectionTimeoutTimer?.cancel();
  }

  /// Resets the connection (useful for token refresh)
  Future<void> resetConnection({String? newAuthToken}) async {
    await disconnect();
    await Future.delayed(const Duration(milliseconds: 500));
    await connect(authToken: newAuthToken);
  }

  /// Disposes of the service
  void dispose() {
    disconnect();
    _messageController.close();
    _connectionStateController.close();
    _errorController.close();
    _messageHandlers.clear();
  }
}

class WebSocketMessage {
  final String type;
  final Map<String, dynamic> data;
  final DateTime timestamp;
  final String? id;

  const WebSocketMessage({
    required this.type,
    required this.data,
    required this.timestamp,
    this.id,
  });

  Map<String, dynamic> toJson() => {
    'type': type,
    'data': data,
    'timestamp': timestamp.toIso8601String(),
    if (id != null) 'id': id,
  };

  factory WebSocketMessage.fromJson(Map<String, dynamic> json) => WebSocketMessage(
    type: json['type'] as String,
    data: json['data'] as Map<String, dynamic>,
    timestamp: DateTime.parse(json['timestamp'] as String),
    id: json['id'] as String?,
  );
}

enum WebSocketConnectionState {
  disconnected,
  connecting,
  connected,
  error,
  failed,
}

class WebSocketException extends AppException {
  WebSocketException(String message) : super(message);
}

// Helper methods for common message types
extension WebSocketServiceHelpers on WebSocketService {
  /// Subscribes to job updates
  void subscribeToJobUpdates(Function(Map<String, dynamic>) onJobUpdate) {
    subscribe('job_updated', onJobUpdate);
    subscribe('job_created', onJobUpdate);
    subscribe('job_deleted', onJobUpdate);
  }

  /// Subscribes to location updates
  void subscribeToLocationUpdates(Function(Map<String, dynamic>) onLocationUpdate) {
    subscribe('location_updated', onLocationUpdate);
  }

  /// Subscribes to chat messages
  void subscribeToChatMessages(Function(Map<String, dynamic>) onChatMessage) {
    subscribe('chat_message', onChatMessage);
  }

  /// Sends a location update
  void sendLocationUpdate({
    required String userId,
    required double latitude,
    required double longitude,
    String? jobId,
  }) {
    final message = WebSocketMessage(
      type: 'location_update',
      data: {
        'userId': userId,
        'latitude': latitude,
        'longitude': longitude,
        if (jobId != null) 'jobId': jobId,
      },
      timestamp: DateTime.now(),
    );
    sendMessage(message);
  }

  /// Sends a job status update
  void sendJobStatusUpdate({
    required String jobId,
    required String status,
    String? notes,
  }) {
    final message = WebSocketMessage(
      type: 'job_status_update',
      data: {
        'jobId': jobId,
        'status': status,
        if (notes != null) 'notes': notes,
      },
      timestamp: DateTime.now(),
    );
    sendMessage(message);
  }

  /// Sends a chat message
  void sendChatMessage({
    required String message,
    required String channelId,
    String? replyToId,
  }) {
    final wsMessage = WebSocketMessage(
      type: 'chat_message',
      data: {
        'message': message,
        'channelId': channelId,
        if (replyToId != null) 'replyToId': replyToId,
      },
      timestamp: DateTime.now(),
    );
    sendMessage(wsMessage);
  }
}