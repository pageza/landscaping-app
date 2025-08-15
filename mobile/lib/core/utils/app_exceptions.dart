/// Base exception class for the application
class AppException implements Exception {
  final String message;
  final String? code;
  final dynamic details;

  const AppException(this.message, {this.code, this.details});

  @override
  String toString() {
    return 'AppException: $message';
  }
}

/// Network related exceptions
class NetworkException extends AppException {
  const NetworkException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'NetworkException: $message';
  }
}

/// Server related exceptions
class ServerException extends AppException {
  const ServerException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'ServerException: $message';
  }
}

/// Authentication related exceptions
class AuthException extends AppException {
  const AuthException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'AuthException: $message';
  }
}

/// Validation related exceptions
class ValidationException extends AppException {
  final Map<String, List<String>>? errors;

  const ValidationException(String message, {String? code, this.errors, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'ValidationException: $message';
  }
}

/// Cache related exceptions
class CacheException extends AppException {
  const CacheException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'CacheException: $message';
  }
}

/// Permission related exceptions
class PermissionException extends AppException {
  const PermissionException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'PermissionException: $message';
  }
}

/// Location related exceptions
class LocationException extends AppException {
  const LocationException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'LocationException: $message';
  }
}

/// Camera related exceptions
class CameraException extends AppException {
  const CameraException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'CameraException: $message';
  }
}

/// File upload related exceptions
class FileUploadException extends AppException {
  const FileUploadException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'FileUploadException: $message';
  }
}

/// Sync related exceptions
class SyncException extends AppException {
  const SyncException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'SyncException: $message';
  }
}

/// Offline related exceptions
class OfflineException extends AppException {
  const OfflineException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'OfflineException: $message';
  }
}

/// Biometric authentication exceptions
class BiometricException extends AppException {
  const BiometricException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'BiometricException: $message';
  }
}

/// Database related exceptions
class DatabaseException extends AppException {
  const DatabaseException(String message, {String? code, dynamic details})
      : super(message, code: code, details: details);

  @override
  String toString() {
    return 'DatabaseException: $message';
  }
}

/// Utility class for exception handling
class ExceptionHandler {
  /// Extracts user-friendly message from exception
  static String getErrorMessage(dynamic error) {
    if (error is AppException) {
      return error.message;
    } else if (error is Exception) {
      return error.toString();
    } else {
      return 'An unexpected error occurred';
    }
  }

  /// Determines if the error is a network error
  static bool isNetworkError(dynamic error) {
    return error is NetworkException;
  }

  /// Determines if the error is an authentication error
  static bool isAuthError(dynamic error) {
    return error is AuthException;
  }

  /// Determines if the error is a validation error
  static bool isValidationError(dynamic error) {
    return error is ValidationException;
  }

  /// Determines if the error is a server error
  static bool isServerError(dynamic error) {
    return error is ServerException;
  }

  /// Determines if the error requires user action
  static bool requiresUserAction(dynamic error) {
    return error is AuthException || 
           error is ValidationException || 
           error is PermissionException;
  }

  /// Determines if the operation can be retried
  static bool canRetry(dynamic error) {
    return error is NetworkException || 
           error is ServerException ||
           error is SyncException;
  }
}