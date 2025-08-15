import 'dart:async';
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:dio/dio.dart';
import '../config/app_config.dart';

class ConnectivityService {
  static final ConnectivityService _instance = ConnectivityService._internal();
  factory ConnectivityService() => _instance;
  ConnectivityService._internal();

  final Connectivity _connectivity = Connectivity();
  final Dio _dio = Dio();
  
  StreamController<bool>? _connectivityController;
  Stream<bool>? _connectivityStream;
  
  bool _isConnected = true;
  Timer? _connectivityTimer;

  /// Initialize the connectivity service
  void initialize() {
    _connectivityController = StreamController<bool>.broadcast();
    _connectivityStream = _connectivityController!.stream;
    
    // Listen to connectivity changes
    _connectivity.onConnectivityChanged.listen(_onConnectivityChanged);
    
    // Initial connectivity check
    _checkConnectivity();
    
    // Periodic connectivity check
    _startPeriodicCheck();
  }

  /// Get current connectivity status
  bool get isConnected => _isConnected;

  /// Get connectivity stream
  Stream<bool>? get connectivityStream => _connectivityStream;

  /// Check if device is connected to internet
  Future<bool> get hasInternetConnection async {
    try {
      final connectivityResults = await _connectivity.checkConnectivity();
      
      if (connectivityResults.contains(ConnectivityResult.none) || connectivityResults.isEmpty) {
        return false;
      }
      
      // Additional check by making a lightweight network request
      return await _pingServer();
    } catch (e) {
      return false;
    }
  }

  /// Ping server to verify internet connectivity
  Future<bool> _pingServer() async {
    try {
      final response = await _dio.get(
        '${AppConfig.apiBaseUrl}/ping',
        options: Options(
          sendTimeout: const Duration(seconds: 5),
          receiveTimeout: const Duration(seconds: 5),
        ),
      );
      return response.statusCode == 200;
    } catch (e) {
      // If ping endpoint doesn't exist, try a simple HEAD request
      try {
        final response = await _dio.head(
          AppConfig.apiBaseUrl,
          options: Options(
            sendTimeout: const Duration(seconds: 5),
            receiveTimeout: const Duration(seconds: 5),
          ),
        );
        return response.statusCode != null;
      } catch (e) {
        return false;
      }
    }
  }

  /// Handle connectivity changes
  void _onConnectivityChanged(List<ConnectivityResult> results) {
    _checkConnectivity();
  }

  /// Check connectivity and update status
  Future<void> _checkConnectivity() async {
    final wasConnected = _isConnected;
    _isConnected = await hasInternetConnection;
    
    if (wasConnected != _isConnected) {
      _connectivityController?.add(_isConnected);
    }
  }

  /// Start periodic connectivity check
  void _startPeriodicCheck() {
    _connectivityTimer = Timer.periodic(
      const Duration(seconds: 30),
      (_) => _checkConnectivity(),
    );
  }

  /// Wait for internet connection
  Future<bool> waitForConnection({Duration timeout = const Duration(seconds: 30)}) async {
    if (_isConnected) return true;
    
    final completer = Completer<bool>();
    late StreamSubscription subscription;
    
    subscription = connectivityStream!.listen((isConnected) {
      if (isConnected) {
        subscription.cancel();
        if (!completer.isCompleted) {
          completer.complete(true);
        }
      }
    });
    
    // Set timeout
    Timer(timeout, () {
      subscription.cancel();
      if (!completer.isCompleted) {
        completer.complete(false);
      }
    });
    
    return completer.future;
  }

  /// Force connectivity check
  Future<bool> forceCheck() async {
    await _checkConnectivity();
    return _isConnected;
  }

  /// Get connection type
  Future<List<ConnectivityResult>> getConnectionType() async {
    return await _connectivity.checkConnectivity();
  }

  /// Check if connection is metered (mobile data)
  Future<bool> isMeteredConnection() async {
    final results = await _connectivity.checkConnectivity();
    return results.contains(ConnectivityResult.mobile);
  }

  /// Check if connection is WiFi
  Future<bool> isWiFiConnection() async {
    final results = await _connectivity.checkConnectivity();
    return results.contains(ConnectivityResult.wifi);
  }

  /// Dispose the service
  void dispose() {
    _connectivityController?.close();
    _connectivityTimer?.cancel();
  }
}

/// Connectivity states
enum ConnectionState {
  connected,
  disconnected,
  connecting,
  unknown,
}

/// Connectivity info
class ConnectivityInfo {
  final bool isConnected;
  final ConnectivityResult connectionType;
  final bool isMetered;
  final DateTime timestamp;

  ConnectivityInfo({
    required this.isConnected,
    required this.connectionType,
    required this.isMetered,
    required this.timestamp,
  });

  @override
  String toString() {
    return 'ConnectivityInfo(isConnected: $isConnected, '
           'connectionType: $connectionType, '
           'isMetered: $isMetered, '
           'timestamp: $timestamp)';
  }
}