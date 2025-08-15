import 'dart:async';
import 'dart:convert';
import 'dart:math';
import 'package:dio/dio.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:logger/logger.dart';
import 'package:url_launcher/url_launcher.dart';
import '../config/app_config.dart';
import '../storage/database_service.dart';
import '../../features/shared/domain/entities/location.dart';
import '../utils/app_exceptions.dart';
import 'location_service.dart';

class MapsService {
  static MapsService? _instance;
  final Logger _logger = Logger();
  final DatabaseService _databaseService = DatabaseService();
  final LocationService _locationService = LocationService();
  final Dio _dio = Dio();

  // Cache for routes and geocoding
  final Map<String, RouteData> _routeCache = {};
  final Map<String, LocationData> _geocodeCache = {};
  
  static const Duration _cacheExpiry = Duration(hours: 1);

  MapsService._internal() {
    _dio.options = BaseOptions(
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      headers: {
        'Content-Type': 'application/json',
      },
    );
  }

  factory MapsService() {
    _instance ??= MapsService._internal();
    return _instance!;
  }

  /// Gets directions between two locations
  Future<RouteData> getDirections({
    required LocationData origin,
    required LocationData destination,
    List<LocationData>? waypoints,
    String travelMode = 'driving',
    bool optimizeWaypoints = false,
    bool avoidTolls = false,
    bool avoidHighways = false,
  }) async {
    try {
      final String routeKey = _generateRouteKey(origin, destination, waypoints);
      
      // Check cache first
      if (_routeCache.containsKey(routeKey)) {
        final cachedRoute = _routeCache[routeKey]!;
        if (DateTime.now().difference(cachedRoute.createdAt!) < _cacheExpiry) {
          _logger.d('Returning cached route');
          return cachedRoute;
        }
      }

      final String apiKey = AppConfig.googleMapsApiKey;
      if (apiKey.isEmpty) {
        throw MapsException('Google Maps API key not configured');
      }

      final Map<String, dynamic> params = {
        'origin': '${origin.latitude},${origin.longitude}',
        'destination': '${destination.latitude},${destination.longitude}',
        'mode': travelMode,
        'key': apiKey,
        'alternatives': 'false',
        'units': 'metric',
      };

      if (waypoints != null && waypoints.isNotEmpty) {
        final waypointString = waypoints
            .map((w) => '${w.latitude},${w.longitude}')
            .join('|');
        params['waypoints'] = optimizeWaypoints 
            ? 'optimize:true|$waypointString'
            : waypointString;
      }

      if (avoidTolls || avoidHighways) {
        final List<String> avoid = [];
        if (avoidTolls) avoid.add('tolls');
        if (avoidHighways) avoid.add('highways');
        params['avoid'] = avoid.join('|');
      }

      final Response response = await _dio.get(
        'https://maps.googleapis.com/maps/api/directions/json',
        queryParameters: params,
      );

      if (response.statusCode != 200) {
        throw MapsException('Failed to get directions: ${response.statusCode}');
      }

      final Map<String, dynamic> data = response.data;
      
      if (data['status'] != 'OK') {
        throw MapsException('Directions API error: ${data['status']}');
      }

      final RouteData route = _parseDirectionsResponse(data, origin, destination);
      
      // Cache the route
      _routeCache[routeKey] = route;
      
      // Save to database for offline access
      await _saveRouteToDatabase(route);

      _logger.i('Directions fetched: ${route.distance.toStringAsFixed(1)}km, ${(route.estimatedDuration / 60).toStringAsFixed(0)}min');
      return route;
    } catch (e) {
      _logger.e('Error getting directions: $e');
      
      // Try to get cached route from database
      final cachedRoute = await _getCachedRouteFromDatabase(origin, destination);
      if (cachedRoute != null) {
        _logger.i('Using cached route from database');
        return cachedRoute;
      }
      
      if (e is MapsException) rethrow;
      throw MapsException('Failed to get directions: $e');
    }
  }

  /// Gets multiple route options
  Future<List<RouteData>> getRouteAlternatives({
    required LocationData origin,
    required LocationData destination,
    String travelMode = 'driving',
  }) async {
    try {
      final String apiKey = AppConfig.googleMapsApiKey;
      if (apiKey.isEmpty) {
        throw MapsException('Google Maps API key not configured');
      }

      final Response response = await _dio.get(
        'https://maps.googleapis.com/maps/api/directions/json',
        queryParameters: {
          'origin': '${origin.latitude},${origin.longitude}',
          'destination': '${destination.latitude},${destination.longitude}',
          'mode': travelMode,
          'alternatives': 'true',
          'key': apiKey,
        },
      );

      if (response.statusCode != 200) {
        throw MapsException('Failed to get route alternatives: ${response.statusCode}');
      }

      final Map<String, dynamic> data = response.data;
      
      if (data['status'] != 'OK') {
        throw MapsException('Directions API error: ${data['status']}');
      }

      final List<RouteData> routes = [];
      final List<dynamic> routesData = data['routes'] as List<dynamic>;

      for (final routeData in routesData) {
        final route = _parseRouteData(routeData as Map<String, dynamic>, origin, destination);
        routes.add(route);
      }

      _logger.i('Found ${routes.length} route alternatives');
      return routes;
    } catch (e) {
      _logger.e('Error getting route alternatives: $e');
      if (e is MapsException) rethrow;
      throw MapsException('Failed to get route alternatives: $e');
    }
  }

  /// Optimizes the order of multiple stops
  Future<List<LocationData>> optimizeWaypoints({
    required LocationData origin,
    required LocationData destination,
    required List<LocationData> waypoints,
  }) async {
    try {
      if (waypoints.isEmpty) return waypoints;

      final String apiKey = AppConfig.googleMapsApiKey;
      if (apiKey.isEmpty) {
        throw MapsException('Google Maps API key not configured');
      }

      final waypointString = waypoints
          .map((w) => '${w.latitude},${w.longitude}')
          .join('|');

      final Response response = await _dio.get(
        'https://maps.googleapis.com/maps/api/directions/json',
        queryParameters: {
          'origin': '${origin.latitude},${origin.longitude}',
          'destination': '${destination.latitude},${destination.longitude}',
          'waypoints': 'optimize:true|$waypointString',
          'key': apiKey,
        },
      );

      if (response.statusCode != 200) {
        throw MapsException('Failed to optimize waypoints: ${response.statusCode}');
      }

      final Map<String, dynamic> data = response.data;
      
      if (data['status'] != 'OK') {
        throw MapsException('Waypoint optimization error: ${data['status']}');
      }

      final List<dynamic> routes = data['routes'] as List<dynamic>;
      if (routes.isEmpty) {
        return waypoints;
      }

      final List<dynamic> waypointOrder = routes.first['waypoint_order'] as List<dynamic>;
      final List<LocationData> optimizedWaypoints = [];

      for (final index in waypointOrder) {
        optimizedWaypoints.add(waypoints[index as int]);
      }

      _logger.i('Waypoints optimized: ${waypoints.length} stops');
      return optimizedWaypoints;
    } catch (e) {
      _logger.e('Error optimizing waypoints: $e');
      if (e is MapsException) rethrow;
      throw MapsException('Failed to optimize waypoints: $e');
    }
  }

  /// Gets estimated travel time and distance
  Future<TravelInfo> getTravelInfo({
    required LocationData origin,
    required LocationData destination,
    String travelMode = 'driving',
  }) async {
    try {
      final String apiKey = AppConfig.googleMapsApiKey;
      if (apiKey.isEmpty) {
        throw MapsException('Google Maps API key not configured');
      }

      final Response response = await _dio.get(
        'https://maps.googleapis.com/maps/api/distancematrix/json',
        queryParameters: {
          'origins': '${origin.latitude},${origin.longitude}',
          'destinations': '${destination.latitude},${destination.longitude}',
          'mode': travelMode,
          'units': 'metric',
          'departure_time': 'now',
          'traffic_model': 'best_guess',
          'key': apiKey,
        },
      );

      if (response.statusCode != 200) {
        throw MapsException('Failed to get travel info: ${response.statusCode}');
      }

      final Map<String, dynamic> data = response.data;
      
      if (data['status'] != 'OK') {
        throw MapsException('Distance Matrix API error: ${data['status']}');
      }

      final List<dynamic> rows = data['rows'] as List<dynamic>;
      if (rows.isEmpty) {
        throw MapsException('No travel info available');
      }

      final List<dynamic> elements = rows.first['elements'] as List<dynamic>;
      if (elements.isEmpty) {
        throw MapsException('No travel info available');
      }

      final Map<String, dynamic> element = elements.first as Map<String, dynamic>;
      
      if (element['status'] != 'OK') {
        throw MapsException('Travel info error: ${element['status']}');
      }

      final int distanceMeters = element['distance']['value'] as int;
      final int durationSeconds = element['duration']['value'] as int;
      final int? durationInTrafficSeconds = element['duration_in_traffic']?['value'] as int?;

      return TravelInfo(
        distance: distanceMeters / 1000.0, // Convert to kilometers
        duration: durationSeconds / 60.0, // Convert to minutes
        durationInTraffic: durationInTrafficSeconds != null 
            ? durationInTrafficSeconds / 60.0 
            : null,
        trafficCondition: _determineTrafficCondition(
          durationSeconds,
          durationInTrafficSeconds,
        ),
      );
    } catch (e) {
      _logger.e('Error getting travel info: $e');
      if (e is MapsException) rethrow;
      throw MapsException('Failed to get travel info: $e');
    }
  }

  /// Opens navigation in external maps app
  Future<void> openNavigation({
    required LocationData destination,
    LocationData? origin,
    String? destinationLabel,
  }) async {
    try {
      final String destinationQuery = destinationLabel != null
          ? Uri.encodeComponent(destinationLabel)
          : '${destination.latitude},${destination.longitude}';

      String mapsUrl;
      
      if (origin != null) {
        // Navigation from specific origin
        mapsUrl = 'https://www.google.com/maps/dir/${origin.latitude},${origin.longitude}/$destinationQuery';
      } else {
        // Navigation from current location
        mapsUrl = 'https://www.google.com/maps/search/?api=1&query=$destinationQuery';
      }

      final Uri uri = Uri.parse(mapsUrl);
      
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
        _logger.i('Navigation opened for destination: $destinationLabel');
      } else {
        throw MapsException('Could not open navigation');
      }
    } catch (e) {
      _logger.e('Error opening navigation: $e');
      throw MapsException('Failed to open navigation: $e');
    }
  }

  /// Creates markers for job locations
  Set<Marker> createJobMarkers(List<JobLocation> jobLocations) {
    return jobLocations.where((job) => job.latitude != null && job.longitude != null)
        .map((job) => Marker(
          markerId: MarkerId(job.address),
          position: LatLng(job.latitude!, job.longitude!),
          infoWindow: InfoWindow(
            title: job.address,
            snippet: job.specialInstructions ?? job.city,
          ),
          icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueBlue),
        ))
        .toSet();
  }

  /// Creates route polyline
  Polyline createRoutePolyline(RouteData route, {Color? color}) {
    final List<LatLng> points = route.waypoints
        .map((point) => LatLng(point.latitude, point.longitude))
        .toList();

    return Polyline(
      polylineId: PolylineId(route.id),
      points: points,
      color: color ?? const Color(0xFF2196F3),
      width: 4,
      patterns: [],
      startCap: Cap.roundCap,
      endCap: Cap.roundCap,
    );
  }

  /// Gets nearby places (gas stations, restaurants, etc.)
  Future<List<NearbyPlace>> getNearbyPlaces({
    required LocationData location,
    required String placeType,
    int radius = 5000,
  }) async {
    try {
      final String apiKey = AppConfig.googleMapsApiKey;
      if (apiKey.isEmpty) {
        throw MapsException('Google Maps API key not configured');
      }

      final Response response = await _dio.get(
        'https://maps.googleapis.com/maps/api/place/nearbysearch/json',
        queryParameters: {
          'location': '${location.latitude},${location.longitude}',
          'radius': radius.toString(),
          'type': placeType,
          'key': apiKey,
        },
      );

      if (response.statusCode != 200) {
        throw MapsException('Failed to get nearby places: ${response.statusCode}');
      }

      final Map<String, dynamic> data = response.data;
      
      if (data['status'] != 'OK' && data['status'] != 'ZERO_RESULTS') {
        throw MapsException('Places API error: ${data['status']}');
      }

      final List<dynamic> results = data['results'] as List<dynamic>;
      final List<NearbyPlace> places = [];

      for (final result in results) {
        final place = _parseNearbyPlace(result as Map<String, dynamic>);
        places.add(place);
      }

      _logger.i('Found ${places.length} nearby ${placeType}s');
      return places;
    } catch (e) {
      _logger.e('Error getting nearby places: $e');
      if (e is MapsException) rethrow;
      throw MapsException('Failed to get nearby places: $e');
    }
  }

  /// Parses directions API response
  RouteData _parseDirectionsResponse(
    Map<String, dynamic> data,
    LocationData origin,
    LocationData destination,
  ) {
    final List<dynamic> routes = data['routes'] as List<dynamic>;
    if (routes.isEmpty) {
      throw MapsException('No routes found');
    }

    return _parseRouteData(routes.first as Map<String, dynamic>, origin, destination);
  }

  /// Parses individual route data
  RouteData _parseRouteData(
    Map<String, dynamic> routeData,
    LocationData origin,
    LocationData destination,
  ) {
    final Map<String, dynamic> leg = (routeData['legs'] as List<dynamic>).first as Map<String, dynamic>;
    final String encodedPolyline = routeData['overview_polyline']['points'] as String;
    
    final double distance = (leg['distance']['value'] as int) / 1000.0; // Convert to km
    final double duration = (leg['duration']['value'] as int) / 60.0; // Convert to minutes

    final List<LocationData> waypoints = _decodePolyline(encodedPolyline);
    final List<RouteStep> steps = _parseRouteSteps(leg['steps'] as List<dynamic>);

    return RouteData(
      id: _generateRouteId(origin, destination),
      startAddress: leg['start_address'] as String,
      endAddress: leg['end_address'] as String,
      startLocation: origin,
      endLocation: destination,
      distance: distance,
      estimatedDuration: duration,
      waypoints: waypoints,
      encodedPolyline: encodedPolyline,
      steps: steps,
      createdAt: DateTime.now(),
    );
  }

  /// Parses route steps
  List<RouteStep> _parseRouteSteps(List<dynamic> stepsData) {
    return stepsData.map((stepData) {
      final Map<String, dynamic> step = stepData as Map<String, dynamic>;
      
      return RouteStep(
        instruction: _stripHtmlTags(step['html_instructions'] as String),
        distance: (step['distance']['value'] as int) / 1000.0,
        duration: (step['duration']['value'] as int) / 60.0,
        startLocation: LocationData(
          latitude: step['start_location']['lat'] as double,
          longitude: step['start_location']['lng'] as double,
        ),
        endLocation: LocationData(
          latitude: step['end_location']['lat'] as double,
          longitude: step['end_location']['lng'] as double,
        ),
        maneuver: step['maneuver'] as String?,
      );
    }).toList();
  }

  /// Decodes Google Maps polyline
  List<LocationData> _decodePolyline(String encoded) {
    final List<LocationData> points = [];
    int index = 0;
    int lat = 0;
    int lng = 0;

    while (index < encoded.length) {
      int shift = 0;
      int result = 0;
      int byte;
      do {
        byte = encoded.codeUnitAt(index++) - 63;
        result |= (byte & 0x1f) << shift;
        shift += 5;
      } while (byte >= 0x20);
      int dlat = ((result & 1) != 0 ? ~(result >> 1) : (result >> 1));
      lat += dlat;

      shift = 0;
      result = 0;
      do {
        byte = encoded.codeUnitAt(index++) - 63;
        result |= (byte & 0x1f) << shift;
        shift += 5;
      } while (byte >= 0x20);
      int dlng = ((result & 1) != 0 ? ~(result >> 1) : (result >> 1));
      lng += dlng;

      points.add(LocationData(
        latitude: lat / 1e5,
        longitude: lng / 1e5,
      ));
    }

    return points;
  }

  /// Parses nearby place data
  NearbyPlace _parseNearbyPlace(Map<String, dynamic> placeData) {
    final Map<String, dynamic> geometry = placeData['geometry'] as Map<String, dynamic>;
    final Map<String, dynamic> location = geometry['location'] as Map<String, dynamic>;

    return NearbyPlace(
      id: placeData['place_id'] as String,
      name: placeData['name'] as String,
      address: placeData['vicinity'] as String? ?? '',
      location: LocationData(
        latitude: location['lat'] as double,
        longitude: location['lng'] as double,
      ),
      rating: (placeData['rating'] as num?)?.toDouble(),
      priceLevel: placeData['price_level'] as int?,
      types: (placeData['types'] as List<dynamic>).cast<String>(),
      isOpen: placeData['opening_hours']?['open_now'] as bool?,
    );
  }

  /// Generates route cache key
  String _generateRouteKey(LocationData origin, LocationData destination, List<LocationData>? waypoints) {
    final StringBuffer buffer = StringBuffer();
    buffer.write('${origin.latitude.toStringAsFixed(6)},${origin.longitude.toStringAsFixed(6)}');
    buffer.write('-${destination.latitude.toStringAsFixed(6)},${destination.longitude.toStringAsFixed(6)}');
    
    if (waypoints != null) {
      for (final waypoint in waypoints) {
        buffer.write('|${waypoint.latitude.toStringAsFixed(6)},${waypoint.longitude.toStringAsFixed(6)}');
      }
    }
    
    return buffer.toString();
  }

  /// Generates route ID
  String _generateRouteId(LocationData origin, LocationData destination) {
    return '${origin.latitude.toStringAsFixed(6)}_${origin.longitude.toStringAsFixed(6)}_'
           '${destination.latitude.toStringAsFixed(6)}_${destination.longitude.toStringAsFixed(6)}_'
           '${DateTime.now().millisecondsSinceEpoch}';
  }

  /// Saves route to database
  Future<void> _saveRouteToDatabase(RouteData route) async {
    try {
      await _databaseService.insert('routes', {
        'id': route.id,
        'start_address': route.startAddress,
        'end_address': route.endAddress,
        'start_location': jsonEncode(route.startLocation.toJson()),
        'end_location': jsonEncode(route.endLocation.toJson()),
        'distance': route.distance,
        'estimated_duration': route.estimatedDuration,
        'waypoints': jsonEncode(route.waypoints.map((w) => w.toJson()).toList()),
        'encoded_polyline': route.encodedPolyline,
        'steps': jsonEncode(route.steps.map((s) => s.toJson()).toList()),
        'created_at': route.createdAt?.toIso8601String(),
      });
    } catch (e) {
      _logger.e('Error saving route to database: $e');
    }
  }

  /// Gets cached route from database
  Future<RouteData?> _getCachedRouteFromDatabase(LocationData origin, LocationData destination) async {
    try {
      final List<Map<String, dynamic>> results = await _databaseService.query(
        'routes',
        orderBy: 'created_at DESC',
        limit: 1,
      );

      if (results.isNotEmpty) {
        final Map<String, dynamic> data = results.first;
        return RouteData(
          id: data['id'] as String,
          startAddress: data['start_address'] as String,
          endAddress: data['end_address'] as String,
          startLocation: LocationData.fromJson(jsonDecode(data['start_location'] as String)),
          endLocation: LocationData.fromJson(jsonDecode(data['end_location'] as String)),
          distance: data['distance'] as double,
          estimatedDuration: data['estimated_duration'] as double,
          waypoints: (jsonDecode(data['waypoints'] as String) as List<dynamic>)
              .map((w) => LocationData.fromJson(w as Map<String, dynamic>))
              .toList(),
          encodedPolyline: data['encoded_polyline'] as String?,
          steps: (jsonDecode(data['steps'] as String) as List<dynamic>)
              .map((s) => RouteStep.fromJson(s as Map<String, dynamic>))
              .toList(),
          createdAt: DateTime.parse(data['created_at'] as String),
        );
      }
    } catch (e) {
      _logger.e('Error getting cached route: $e');
    }
    return null;
  }

  /// Determines traffic condition based on duration comparison
  TrafficCondition _determineTrafficCondition(int normalDuration, int? trafficDuration) {
    if (trafficDuration == null) return TrafficCondition.unknown;
    
    final double ratio = trafficDuration / normalDuration;
    
    if (ratio < 1.1) return TrafficCondition.clear;
    if (ratio < 1.3) return TrafficCondition.light;
    if (ratio < 1.6) return TrafficCondition.moderate;
    if (ratio < 2.0) return TrafficCondition.heavy;
    return TrafficCondition.severe;
  }

  /// Strips HTML tags from text
  String _stripHtmlTags(String htmlText) {
    return htmlText.replaceAll(RegExp(r'<[^>]*>'), '');
  }

  /// Clears route cache
  void clearCache() {
    _routeCache.clear();
    _geocodeCache.clear();
    _logger.i('Maps cache cleared');
  }
}

class TravelInfo {
  final double distance; // in kilometers
  final double duration; // in minutes
  final double? durationInTraffic; // in minutes
  final TrafficCondition trafficCondition;

  const TravelInfo({
    required this.distance,
    required this.duration,
    this.durationInTraffic,
    required this.trafficCondition,
  });

  String get distanceFormatted => '${distance.toStringAsFixed(1)} km';
  String get durationFormatted => '${duration.toStringAsFixed(0)} min';
  String get durationInTrafficFormatted => durationInTraffic != null 
      ? '${durationInTraffic!.toStringAsFixed(0)} min'
      : durationFormatted;
}

class NearbyPlace {
  final String id;
  final String name;
  final String address;
  final LocationData location;
  final double? rating;
  final int? priceLevel;
  final List<String> types;
  final bool? isOpen;

  const NearbyPlace({
    required this.id,
    required this.name,
    required this.address,
    required this.location,
    this.rating,
    this.priceLevel,
    required this.types,
    this.isOpen,
  });
}

class MapsException extends AppException {
  MapsException(String message) : super(message);
}