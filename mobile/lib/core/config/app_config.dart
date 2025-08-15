import 'package:flutter_dotenv/flutter_dotenv.dart';

class AppConfig {
  // Environment
  static String get environment => dotenv.env['ENVIRONMENT'] ?? 'development';
  static bool get isDevelopment => environment == 'development';
  static bool get isProduction => environment == 'production';

  // API Configuration
  static String get apiBaseUrl => dotenv.env['API_BASE_URL'] ?? 'http://localhost:8080/api/v1';
  static String get websocketUrl => dotenv.env['WEBSOCKET_URL'] ?? 'ws://localhost:8080/ws';

  // Database Configuration
  static String get dbName => dotenv.env['DB_NAME'] ?? 'landscaping_app.db';
  static int get dbVersion => int.parse(dotenv.env['DB_VERSION'] ?? '1');

  // Security Configuration
  static String get jwtSecretKey => dotenv.env['JWT_SECRET_KEY'] ?? '';
  static String get encryptionKey => dotenv.env['ENCRYPTION_KEY'] ?? '';

  // Firebase Configuration
  static String get firebaseProjectId => dotenv.env['FIREBASE_PROJECT_ID'] ?? '';
  static String get firebaseMessagingSenderId => dotenv.env['FIREBASE_MESSAGING_SENDER_ID'] ?? '';
  static String get firebaseAppId => dotenv.env['FIREBASE_APP_ID'] ?? '';
  static String get firebaseApiKey => dotenv.env['FIREBASE_API_KEY'] ?? '';

  // Maps Configuration
  static String get googleMapsApiKey => dotenv.env['GOOGLE_MAPS_API_KEY'] ?? '';

  // Feature Flags
  static bool get enableBiometricAuth => dotenv.env['ENABLE_BIOMETRIC_AUTH']?.toLowerCase() == 'true';
  static bool get enableOfflineMode => dotenv.env['ENABLE_OFFLINE_MODE']?.toLowerCase() == 'true';
  static bool get enableDebugLogging => dotenv.env['ENABLE_DEBUG_LOGGING']?.toLowerCase() == 'true';

  // App Configuration
  static String get appName => dotenv.env['APP_NAME'] ?? 'Landscaping Pro';
  static String get companyName => dotenv.env['COMPANY_NAME'] ?? 'Your Company Name';
  static String get supportEmail => dotenv.env['SUPPORT_EMAIL'] ?? 'support@yourcompany.com';
  static String get privacyPolicyUrl => dotenv.env['PRIVACY_POLICY_URL'] ?? 'https://yourcompany.com/privacy';
  static String get termsOfServiceUrl => dotenv.env['TERMS_OF_SERVICE_URL'] ?? 'https://yourcompany.com/terms';

  // Timeout configurations
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);
  static const Duration sendTimeout = Duration(seconds: 30);

  // Pagination
  static const int defaultPageSize = 20;
  static const int maxPageSize = 100;

  // Cache configurations
  static const Duration cacheMaxAge = Duration(hours: 1);
  static const Duration imagesCacheMaxAge = Duration(days: 7);

  // File upload configurations
  static const int maxFileSize = 10 * 1024 * 1024; // 10MB
  static const List<String> allowedImageTypes = ['jpg', 'jpeg', 'png', 'gif'];
  static const List<String> allowedDocumentTypes = ['pdf', 'doc', 'docx'];
}