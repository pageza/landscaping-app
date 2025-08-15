import 'dart:async';
import 'dart:convert';
import 'package:geolocator/geolocator.dart';
import 'package:geocoding/geocoding.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:logger/logger.dart';
import '../storage/database_service.dart';
import '../../features/shared/domain/entities/location.dart';
import '../utils/app_exceptions.dart';
import 'websocket_service.dart';

class LocationService {
  static LocationService? _instance;
  final Logger _logger = Logger();
  final DatabaseService _databaseService = DatabaseService();
  final WebSocketService _webSocketService = WebSocketService();

  StreamSubscription<Position>? _positionSubscription;
  Timer? _trackingTimer;
  bool _isTracking = false;
  bool _isBackgroundTracking = false;
  String? _currentJobId;
  String? _currentUserId;
  
  // Location settings
  static const LocationSettings _locationSettings = LocationSettings(
    accuracy: LocationAccuracy.high,
    distanceFilter: 10, // Update every 10 meters
  );

  static const LocationSettings _backgroundLocationSettings = LocationSettings(
    accuracy: LocationAccuracy.medium,
    distanceFilter: 50, // Update every 50 meters in background
  );

  // Geofence management
  final List<GeofenceArea> _activeGeofences = [];
  final Map<String, Timer> _geofenceTimers = {};

  LocationService._internal();

  factory LocationService() {
    _instance ??= LocationService._internal();
    return _instance!;
  }

  bool get isTracking => _isTracking;
  bool get isBackgroundTracking => _isBackgroundTracking;
  String? get currentJobId => _currentJobId;

  /// Checks and requests location permissions
  Future<bool> checkAndRequestPermissions() async {
    try {
      // Check if location services are enabled
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        _logger.w('Location services are disabled');
        throw LocationException('Location services are disabled. Please enable them in settings.');
      }

      // Check location permission
      LocationPermission permission = await Geolocator.checkPermission();
      
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) {
          _logger.w('Location permission denied');
          throw LocationException('Location permission denied');
        }
      }

      if (permission == LocationPermission.deniedForever) {
        _logger.e('Location permission permanently denied');
        throw LocationException('Location permission permanently denied. Please enable in app settings.');
      }

      // For background tracking, check background location permission on Android
      if (permission == LocationPermission.whileInUse) {
        final backgroundPermission = await Permission.locationAlways.status;
        if (backgroundPermission.isDenied) {
          _logger.i('Background location permission not granted');
        }
      }

      _logger.i('Location permissions granted: $permission');
      return true;
    } catch (e) {
      _logger.e('Error checking location permissions: $e');
      if (e is LocationException) rethrow;
      throw LocationException('Failed to check location permissions: $e');
    }
  }

  /// Gets the current location
  Future<LocationData> getCurrentLocation() async {
    try {
      await checkAndRequestPermissions();

      final Position position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
        timeLimit: const Duration(seconds: 10),
      );

      return await _positionToLocationData(position);
    } catch (e) {
      _logger.e('Error getting current location: $e');
      if (e is LocationException) rethrow;
      throw LocationException('Failed to get current location: $e');
    }
  }

  /// Starts location tracking for a job
  Future<void> startTracking({
    required String userId,
    String? jobId,
    bool backgroundTracking = false,
  }) async {
    try {
      if (_isTracking) {
        _logger.w('Location tracking already active');
        return;
      }

      await checkAndRequestPermissions();

      _currentUserId = userId;
      _currentJobId = jobId;
      _isTracking = true;
      _isBackgroundTracking = backgroundTracking;

      final LocationSettings settings = backgroundTracking 
          ? _backgroundLocationSettings 
          : _locationSettings;

      _positionSubscription = Geolocator.getPositionStream(
        locationSettings: settings,
      ).listen(
        (Position position) => _handleLocationUpdate(position),
        onError: (error) => _handleLocationError(error),
      );

      // Also start periodic tracking for more reliable updates
      _trackingTimer = Timer.periodic(
        Duration(minutes: backgroundTracking ? 5 : 2),
        (_) => _performPeriodicLocationUpdate(),
      );

      _logger.i('Location tracking started (background: $backgroundTracking)');
    } catch (e) {
      _logger.e('Error starting location tracking: $e');
      throw LocationException('Failed to start location tracking: $e');
    }
  }

  /// Stops location tracking
  Future<void> stopTracking() async {
    try {
      await _positionSubscription?.cancel();
      _positionSubscription = null;

      _trackingTimer?.cancel();
      _trackingTimer = null;

      _isTracking = false;
      _isBackgroundTracking = false;
      _currentJobId = null;
      _currentUserId = null;

      _logger.i('Location tracking stopped');
    } catch (e) {
      _logger.e('Error stopping location tracking: $e');
    }
  }

  /// Records a manual location check-in
  Future<LocationTracking> checkIn({
    required String userId,
    required String jobId,
    String? notes,
    Map<String, dynamic>? metadata,
  }) async {
    try {
      final LocationData location = await getCurrentLocation();
      
      final LocationTracking tracking = LocationTracking(
        id: _generateId(),
        userId: userId,
        jobId: jobId,
        location: location,
        type: LocationType.checkIn,
        notes: notes,
        metadata: metadata,
        createdAt: DateTime.now(),
      );

      await _saveLocationTracking(tracking);
      await _sendLocationUpdate(tracking);

      _logger.i('Check-in recorded for job $jobId');
      return tracking;
    } catch (e) {
      _logger.e('Error recording check-in: $e');
      throw LocationException('Failed to record check-in: $e');
    }
  }

  /// Records a manual location check-out
  Future<LocationTracking> checkOut({
    required String userId,
    required String jobId,
    String? notes,
    Map<String, dynamic>? metadata,
  }) async {
    try {
      final LocationData location = await getCurrentLocation();
      
      final LocationTracking tracking = LocationTracking(
        id: _generateId(),
        userId: userId,
        jobId: jobId,
        location: location,
        type: LocationType.checkOut,
        notes: notes,
        metadata: metadata,
        createdAt: DateTime.now(),
      );

      await _saveLocationTracking(tracking);
      await _sendLocationUpdate(tracking);

      _logger.i('Check-out recorded for job $jobId');
      return tracking;
    } catch (e) {
      _logger.e('Error recording check-out: $e');
      throw LocationException('Failed to record check-out: $e');
    }
  }

  /// Gets address from coordinates
  Future<String> getAddressFromCoordinates(double latitude, double longitude) async {
    try {
      final List<Placemark> placemarks = await placemarkFromCoordinates(
        latitude,
        longitude,
      );

      if (placemarks.isNotEmpty) {
        final Placemark place = placemarks.first;
        final List<String> addressParts = [
          if (place.street?.isNotEmpty == true) place.street!,
          if (place.locality?.isNotEmpty == true) place.locality!,
          if (place.administrativeArea?.isNotEmpty == true) place.administrativeArea!,
          if (place.postalCode?.isNotEmpty == true) place.postalCode!,
        ];
        return addressParts.join(', ');
      }

      return 'Unknown location';
    } catch (e) {
      _logger.e('Error getting address from coordinates: $e');
      return 'Address lookup failed';
    }
  }

  /// Gets coordinates from address
  Future<LocationData?> getCoordinatesFromAddress(String address) async {
    try {
      final List<Location> locations = await locationFromAddress(address);
      
      if (locations.isNotEmpty) {
        final Location location = locations.first;
        return LocationData(
          latitude: location.latitude,
          longitude: location.longitude,
          timestamp: DateTime.now(),
          address: address,
        );
      }

      return null;
    } catch (e) {
      _logger.e('Error getting coordinates from address: $e');
      return null;
    }
  }

  /// Calculates distance between two locations
  double calculateDistance(
    double startLatitude,
    double startLongitude,
    double endLatitude,
    double endLongitude,
  ) {
    return Geolocator.distanceBetween(
      startLatitude,
      startLongitude,
      endLatitude,
      endLongitude,
    );
  }

  /// Calculates bearing between two locations
  double calculateBearing(
    double startLatitude,
    double startLongitude,
    double endLatitude,
    double endLongitude,
  ) {
    return Geolocator.bearingBetween(
      startLatitude,
      startLongitude,
      endLatitude,
      endLongitude,
    );
  }

  /// Adds a geofence area
  Future<void> addGeofence(GeofenceArea geofence) async {
    try {
      _activeGeofences.add(geofence);
      
      await _databaseService.insert('geofence_areas', {
        'id': geofence.id,
        'name': geofence.name,
        'center_location': jsonEncode(geofence.center.toJson()),
        'radius': geofence.radius,
        'description': geofence.description,
        'is_active': geofence.isActive ? 1 : 0,
        'trigger_events': jsonEncode(geofence.triggerEvents),
        'metadata': jsonEncode(geofence.metadata),
        'company_id': geofence.companyId,
        'created_at': geofence.createdAt?.toIso8601String(),
        'updated_at': geofence.updatedAt?.toIso8601String(),
      });

      _logger.i('Geofence added: ${geofence.name}');
    } catch (e) {
      _logger.e('Error adding geofence: $e');
      throw LocationException('Failed to add geofence: $e');
    }
  }

  /// Removes a geofence area
  Future<void> removeGeofence(String geofenceId) async {
    try {
      _activeGeofences.removeWhere((g) => g.id == geofenceId);
      _geofenceTimers[geofenceId]?.cancel();
      _geofenceTimers.remove(geofenceId);

      await _databaseService.delete(
        'geofence_areas',
        where: 'id = ?',
        whereArgs: [geofenceId],
      );

      _logger.i('Geofence removed: $geofenceId');
    } catch (e) {
      _logger.e('Error removing geofence: $e');
    }
  }

  /// Gets location tracking history
  Future<List<LocationTracking>> getLocationHistory({
    String? userId,
    String? jobId,
    DateTime? startDate,
    DateTime? endDate,
    int? limit,
  }) async {
    try {
      String where = '1=1';
      List<dynamic> whereArgs = [];

      if (userId != null) {
        where += ' AND user_id = ?';
        whereArgs.add(userId);
      }

      if (jobId != null) {
        where += ' AND job_id = ?';
        whereArgs.add(jobId);
      }

      if (startDate != null) {
        where += ' AND created_at >= ?';
        whereArgs.add(startDate.toIso8601String());
      }

      if (endDate != null) {
        where += ' AND created_at <= ?';
        whereArgs.add(endDate.toIso8601String());
      }

      final List<Map<String, dynamic>> results = await _databaseService.query(
        'location_tracking',
        where: where,
        whereArgs: whereArgs,
        orderBy: 'created_at DESC',
        limit: limit,
      );

      return results.map((data) => _locationTrackingFromMap(data)).toList();
    } catch (e) {
      _logger.e('Error getting location history: $e');
      return [];
    }
  }

  /// Handles real-time location updates
  Future<void> _handleLocationUpdate(Position position) async {
    try {
      if (!_isTracking || _currentUserId == null) return;

      final LocationData location = await _positionToLocationData(position);
      
      final LocationTracking tracking = LocationTracking(
        id: _generateId(),
        userId: _currentUserId!,
        jobId: _currentJobId ?? '',
        location: location,
        type: LocationType.tracking,
        createdAt: DateTime.now(),
      );

      await _saveLocationTracking(tracking);
      await _sendLocationUpdate(tracking);
      await _checkGeofences(location);
    } catch (e) {
      _logger.e('Error handling location update: $e');
    }
  }

  /// Handles location errors
  void _handleLocationError(dynamic error) {
    _logger.e('Location stream error: $error');
    // Attempt to restart tracking after a delay
    Timer(const Duration(seconds: 30), () {
      if (_isTracking && _currentUserId != null) {
        _performPeriodicLocationUpdate();
      }
    });
  }

  /// Performs periodic location updates as backup
  Future<void> _performPeriodicLocationUpdate() async {
    try {
      if (!_isTracking || _currentUserId == null) return;

      final Position position = await Geolocator.getCurrentPosition();
      await _handleLocationUpdate(position);
    } catch (e) {
      _logger.e('Error in periodic location update: $e');
    }
  }

  /// Converts Position to LocationData
  Future<LocationData> _positionToLocationData(Position position) async {
    String? address;
    try {
      address = await getAddressFromCoordinates(
        position.latitude,
        position.longitude,
      );
    } catch (e) {
      _logger.w('Failed to get address for location: $e');
    }

    return LocationData(
      latitude: position.latitude,
      longitude: position.longitude,
      accuracy: position.accuracy,
      altitude: position.altitude,
      speed: position.speed,
      heading: position.heading,
      timestamp: position.timestamp,
      address: address,
    );
  }

  /// Saves location tracking to database
  Future<void> _saveLocationTracking(LocationTracking tracking) async {
    try {
      await _databaseService.insert('location_tracking', {
        'id': tracking.id,
        'user_id': tracking.userId,
        'job_id': tracking.jobId,
        'location_data': jsonEncode(tracking.location.toJson()),
        'type': tracking.type.name,
        'notes': tracking.notes,
        'metadata': jsonEncode(tracking.metadata),
        'created_at': tracking.createdAt?.toIso8601String(),
      });
    } catch (e) {
      _logger.e('Error saving location tracking: $e');
    }
  }

  /// Sends location update via WebSocket
  Future<void> _sendLocationUpdate(LocationTracking tracking) async {
    try {
      if (_webSocketService.isConnected) {
        _webSocketService.sendLocationUpdate(
          userId: tracking.userId,
          latitude: tracking.location.latitude,
          longitude: tracking.location.longitude,
          jobId: tracking.jobId.isNotEmpty ? tracking.jobId : null,
        );
      }
    } catch (e) {
      _logger.e('Error sending location update: $e');
    }
  }

  /// Checks if current location triggers any geofences
  Future<void> _checkGeofences(LocationData location) async {
    for (final geofence in _activeGeofences) {
      if (!geofence.isActive) continue;

      final double distance = location.distanceTo(geofence.center);
      final bool isInside = distance <= geofence.radius;
      final String geofenceKey = geofence.id;

      if (isInside && !_geofenceTimers.containsKey(geofenceKey)) {
        // Entered geofence
        await _triggerGeofenceEvent(geofence, GeofenceEventType.enter, location);
        
        // Start dwell timer if configured
        if (geofence.triggerEvents.contains('dwell')) {
          _geofenceTimers[geofenceKey] = Timer(
            const Duration(minutes: 5),
            () => _triggerGeofenceEvent(geofence, GeofenceEventType.dwell, location),
          );
        }
      } else if (!isInside && _geofenceTimers.containsKey(geofenceKey)) {
        // Exited geofence
        _geofenceTimers[geofenceKey]?.cancel();
        _geofenceTimers.remove(geofenceKey);
        await _triggerGeofenceEvent(geofence, GeofenceEventType.exit, location);
      }
    }
  }

  /// Triggers a geofence event
  Future<void> _triggerGeofenceEvent(
    GeofenceArea geofence,
    GeofenceEventType eventType,
    LocationData location,
  ) async {
    try {
      final GeofenceEvent event = GeofenceEvent(
        id: _generateId(),
        geofenceId: geofence.id,
        userId: _currentUserId!,
        eventType: eventType,
        location: location,
        timestamp: DateTime.now(),
        jobId: _currentJobId,
        createdAt: DateTime.now(),
      );

      await _databaseService.insert('geofence_events', {
        'id': event.id,
        'geofence_id': event.geofenceId,
        'user_id': event.userId,
        'event_type': event.eventType.name,
        'location_data': jsonEncode(event.location.toJson()),
        'timestamp': event.timestamp?.toIso8601String(),
        'job_id': event.jobId,
        'metadata': jsonEncode(event.metadata),
        'created_at': event.createdAt?.toIso8601String(),
      });

      _logger.i('Geofence event triggered: ${geofence.name} - ${eventType.name}');
    } catch (e) {
      _logger.e('Error triggering geofence event: $e');
    }
  }

  /// Converts database map to LocationTracking
  LocationTracking _locationTrackingFromMap(Map<String, dynamic> map) {
    return LocationTracking(
      id: map['id'] as String,
      userId: map['user_id'] as String,
      jobId: map['job_id'] as String,
      location: LocationData.fromJson(jsonDecode(map['location_data'] as String)),
      type: LocationType.values.firstWhere(
        (e) => e.name == map['type'],
        orElse: () => LocationType.tracking,
      ),
      notes: map['notes'] as String?,
      metadata: map['metadata'] != null 
          ? jsonDecode(map['metadata'] as String) as Map<String, dynamic>?
          : null,
      createdAt: map['created_at'] != null 
          ? DateTime.parse(map['created_at'] as String)
          : null,
    );
  }

  /// Generates a unique ID
  String _generateId() {
    return DateTime.now().millisecondsSinceEpoch.toString();
  }

  /// Cleans up old location data
  Future<void> cleanupOldLocationData({int maxAgeInDays = 30}) async {
    try {
      final cutoffDate = DateTime.now().subtract(Duration(days: maxAgeInDays));
      
      await _databaseService.delete(
        'location_tracking',
        where: 'created_at < ?',
        whereArgs: [cutoffDate.toIso8601String()],
      );

      await _databaseService.delete(
        'geofence_events',
        where: 'created_at < ?',
        whereArgs: [cutoffDate.toIso8601String()],
      );

      _logger.i('Old location data cleaned up');
    } catch (e) {
      _logger.e('Error cleaning up location data: $e');
    }
  }

  /// Disposes of the service
  void dispose() {
    stopTracking();
    for (final timer in _geofenceTimers.values) {
      timer.cancel();
    }
    _geofenceTimers.clear();
    _activeGeofences.clear();
  }
}

class LocationException extends AppException {
  LocationException(String message) : super(message);
}