import 'package:freezed_annotation/freezed_annotation.dart';

part 'location.freezed.dart';
part 'location.g.dart';

@freezed
class LocationData with _$LocationData {
  const factory LocationData({
    required double latitude,
    required double longitude,
    double? accuracy,
    double? altitude,
    double? speed,
    double? heading,
    DateTime? timestamp,
    String? address,
    String? city,
    String? state,
    String? country,
    String? zipCode,
  }) = _LocationData;

  factory LocationData.fromJson(Map<String, dynamic> json) => _$LocationDataFromJson(json);
}

@freezed
class LocationTracking with _$LocationTracking {
  const factory LocationTracking({
    required String id,
    required String userId,
    required String jobId,
    required LocationData location,
    required LocationType type,
    String? notes,
    Map<String, dynamic>? metadata,
    DateTime? createdAt,
  }) = _LocationTracking;

  factory LocationTracking.fromJson(Map<String, dynamic> json) => _$LocationTrackingFromJson(json);
}

@freezed
class RouteData with _$RouteData {
  const factory RouteData({
    required String id,
    required String startAddress,
    required String endAddress,
    required LocationData startLocation,
    required LocationData endLocation,
    required double distance,
    required double estimatedDuration,
    @Default([]) List<LocationData> waypoints,
    String? encodedPolyline,
    @Default([]) List<RouteStep> steps,
    TrafficCondition? trafficCondition,
    DateTime? createdAt,
  }) = _RouteData;

  factory RouteData.fromJson(Map<String, dynamic> json) => _$RouteDataFromJson(json);
}

@freezed
class RouteStep with _$RouteStep {
  const factory RouteStep({
    required String instruction,
    required double distance,
    required double duration,
    required LocationData startLocation,
    required LocationData endLocation,
    String? maneuver,
  }) = _RouteStep;

  factory RouteStep.fromJson(Map<String, dynamic> json) => _$RouteStepFromJson(json);
}

@freezed
class GeofenceArea with _$GeofenceArea {
  const factory GeofenceArea({
    required String id,
    required String name,
    required LocationData center,
    required double radius,
    String? description,
    @Default(true) bool isActive,
    @Default([]) List<String> triggerEvents,
    Map<String, dynamic>? metadata,
    String? companyId,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) = _GeofenceArea;

  factory GeofenceArea.fromJson(Map<String, dynamic> json) => _$GeofenceAreaFromJson(json);
}

@freezed
class GeofenceEvent with _$GeofenceEvent {
  const factory GeofenceEvent({
    required String id,
    required String geofenceId,
    required String userId,
    required GeofenceEventType eventType,
    required LocationData location,
    DateTime? timestamp,
    String? jobId,
    Map<String, dynamic>? metadata,
    DateTime? createdAt,
  }) = _GeofenceEvent;

  factory GeofenceEvent.fromJson(Map<String, dynamic> json) => _$GeofenceEventFromJson(json);
}

enum LocationType {
  @JsonValue('check_in')
  checkIn,
  @JsonValue('check_out')
  checkOut,
  @JsonValue('arrival')
  arrival,
  @JsonValue('departure')
  departure,
  @JsonValue('tracking')
  tracking,
  @JsonValue('manual')
  manual,
}

enum TrafficCondition {
  @JsonValue('unknown')
  unknown,
  @JsonValue('clear')
  clear,
  @JsonValue('light')
  light,
  @JsonValue('moderate')
  moderate,
  @JsonValue('heavy')
  heavy,
  @JsonValue('severe')
  severe,
}

enum GeofenceEventType {
  @JsonValue('enter')
  enter,
  @JsonValue('exit')
  exit,
  @JsonValue('dwell')
  dwell,
}

// Extension methods
extension LocationTypeExtension on LocationType {
  String get displayName {
    switch (this) {
      case LocationType.checkIn:
        return 'Check In';
      case LocationType.checkOut:
        return 'Check Out';
      case LocationType.arrival:
        return 'Arrival';
      case LocationType.departure:
        return 'Departure';
      case LocationType.tracking:
        return 'Tracking';
      case LocationType.manual:
        return 'Manual';
    }
  }
}

extension TrafficConditionExtension on TrafficCondition {
  String get displayName {
    switch (this) {
      case TrafficCondition.unknown:
        return 'Unknown';
      case TrafficCondition.clear:
        return 'Clear';
      case TrafficCondition.light:
        return 'Light';
      case TrafficCondition.moderate:
        return 'Moderate';
      case TrafficCondition.heavy:
        return 'Heavy';
      case TrafficCondition.severe:
        return 'Severe';
    }
  }
}

extension GeofenceEventTypeExtension on GeofenceEventType {
  String get displayName {
    switch (this) {
      case GeofenceEventType.enter:
        return 'Enter';
      case GeofenceEventType.exit:
        return 'Exit';
      case GeofenceEventType.dwell:
        return 'Dwell';
    }
  }
}

// Utility functions
extension LocationDataExtension on LocationData {
  /// Calculate distance between two locations in kilometers
  double distanceTo(LocationData other) {
    const double earthRadius = 6371; // Earth's radius in kilometers
    
    final double lat1Rad = latitude * (3.14159265359 / 180);
    final double lat2Rad = other.latitude * (3.14159265359 / 180);
    final double deltaLatRad = (other.latitude - latitude) * (3.14159265359 / 180);
    final double deltaLngRad = (other.longitude - longitude) * (3.14159265359 / 180);
    
    final double a = (deltaLatRad / 2).sin() * (deltaLatRad / 2).sin() +
        lat1Rad.cos() * lat2Rad.cos() *
        (deltaLngRad / 2).sin() * (deltaLngRad / 2).sin();
    final double c = 2 * a.sqrt().asin();
    
    return earthRadius * c;
  }
  
  /// Check if this location is within a certain distance of another location
  bool isWithinDistance(LocationData other, double maxDistanceKm) {
    return distanceTo(other) <= maxDistanceKm;
  }
  
  /// Format coordinates as a string
  String get coordinateString => '${latitude.toStringAsFixed(6)}, ${longitude.toStringAsFixed(6)}';
  
  /// Get full address if available, otherwise coordinates
  String get displayAddress => address ?? coordinateString;
}