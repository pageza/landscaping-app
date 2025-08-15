import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:logger/logger.dart';
import '../constants/app_constants.dart';
import '../utils/app_exceptions.dart';

/// Authentication Interceptor
class AuthInterceptor extends Interceptor {
  final FlutterSecureStorage _storage;

  AuthInterceptor(this._storage);

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    // Skip authentication for login/register endpoints
    if (_shouldSkipAuth(options.path)) {
      handler.next(options);
      return;
    }

    // Add authorization header
    final token = await _storage.read(key: AppConstants.keyAccessToken);
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }

    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401) {
      // Token expired, try to refresh
      final refreshed = await _refreshToken();
      if (refreshed) {
        // Retry the original request
        final retryResponse = await _retry(err.requestOptions);
        handler.resolve(retryResponse);
        return;
      } else {
        // Refresh failed, clear tokens and redirect to login
        await _clearTokens();
      }
    }

    handler.next(err);
  }

  bool _shouldSkipAuth(String path) {
    const skipAuthPaths = [
      '/auth/login',
      '/auth/register',
      '/auth/refresh',
      '/auth/forgot-password',
      '/auth/reset-password',
    ];
    return skipAuthPaths.any((skipPath) => path.contains(skipPath));
  }

  Future<bool> _refreshToken() async {
    try {
      final refreshToken = await _storage.read(key: AppConstants.keyRefreshToken);
      if (refreshToken == null) return false;

      final dio = Dio();
      final response = await dio.post(
        '${AppConstants.endpointAuth}/refresh',
        data: {'refresh_token': refreshToken},
      );

      if (response.statusCode == 200) {
        final data = response.data;
        await _storage.write(key: AppConstants.keyAccessToken, value: data['access_token']);
        await _storage.write(key: AppConstants.keyRefreshToken, value: data['refresh_token']);
        return true;
      }
    } catch (e) {
      // Refresh failed
    }
    return false;
  }

  Future<Response> _retry(RequestOptions requestOptions) async {
    final dio = Dio();
    final token = await _storage.read(key: AppConstants.keyAccessToken);
    if (token != null) {
      requestOptions.headers['Authorization'] = 'Bearer $token';
    }

    final options = Options(
      method: requestOptions.method,
      headers: requestOptions.headers,
    );

    return dio.request(
      requestOptions.path,
      data: requestOptions.data,
      queryParameters: requestOptions.queryParameters,
      options: options,
    );
  }

  Future<void> _clearTokens() async {
    await _storage.delete(key: AppConstants.keyAccessToken);
    await _storage.delete(key: AppConstants.keyRefreshToken);
    await _storage.delete(key: AppConstants.keyUserId);
    await _storage.delete(key: AppConstants.keyUserType);
  }
}

/// Error Handling Interceptor
class ErrorInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    AppException exception;

    switch (err.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        exception = NetworkException(AppConstants.errorNetwork);
        break;
      case DioExceptionType.badResponse:
        exception = _handleResponseError(err);
        break;
      case DioExceptionType.cancel:
        exception = AppException('Request was cancelled');
        break;
      case DioExceptionType.connectionError:
        exception = NetworkException(AppConstants.errorNetwork);
        break;
      case DioExceptionType.badCertificate:
        exception = NetworkException('SSL Certificate error');
        break;
      case DioExceptionType.unknown:
      default:
        exception = AppException(err.message ?? AppConstants.errorGeneral);
        break;
    }

    final newError = DioException(
      requestOptions: err.requestOptions,
      error: exception,
      type: err.type,
      response: err.response,
    );

    handler.next(newError);
  }

  AppException _handleResponseError(DioException err) {
    final statusCode = err.response?.statusCode;
    final data = err.response?.data;

    String message = AppConstants.errorGeneral;
    
    if (data is Map<String, dynamic>) {
      message = data['message'] ?? data['error'] ?? message;
    }

    switch (statusCode) {
      case 400:
        return ValidationException(message);
      case 401:
        return AuthException(AppConstants.errorAuth);
      case 403:
        return AuthException(AppConstants.errorPermission);
      case 404:
        return AppException(AppConstants.errorNotFound);
      case 422:
        return ValidationException(message);
      case 500:
      case 502:
      case 503:
      case 504:
        return ServerException(AppConstants.errorServer);
      default:
        return AppException(message);
    }
  }
}

/// Logging Interceptor
class LoggingInterceptor extends Interceptor {
  final Logger _logger;

  LoggingInterceptor(this._logger);

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    _logger.d('''
🚀 REQUEST
${options.method} ${options.uri}
Headers: ${options.headers}
Data: ${options.data}
''');
    handler.next(options);
  }

  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    _logger.d('''
✅ RESPONSE
${response.statusCode} ${response.requestOptions.uri}
Data: ${response.data}
''');
    handler.next(response);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    _logger.e('''
❌ ERROR
${err.response?.statusCode} ${err.requestOptions.uri}
Message: ${err.message}
Data: ${err.response?.data}
''');
    handler.next(err);
  }
}

/// Retry Interceptor
class RetryInterceptor extends Interceptor {
  final int maxRetries;
  final Duration retryDelay;
  final List<int> retryStatusCodes;

  RetryInterceptor({
    this.maxRetries = AppConstants.maxRetryAttempts,
    this.retryDelay = AppConstants.retryDelay,
    this.retryStatusCodes = const [502, 503, 504],
  });

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    final shouldRetry = _shouldRetry(err);
    
    if (shouldRetry && _getRetryCount(err.requestOptions) < maxRetries) {
      await Future.delayed(retryDelay);
      _incrementRetryCount(err.requestOptions);
      
      try {
        final response = await _retry(err.requestOptions);
        handler.resolve(response);
        return;
      } catch (e) {
        // Retry failed, continue with error
      }
    }

    handler.next(err);
  }

  bool _shouldRetry(DioException err) {
    if (err.type == DioExceptionType.connectionTimeout ||
        err.type == DioExceptionType.receiveTimeout ||
        err.type == DioExceptionType.connectionError) {
      return true;
    }

    if (err.response?.statusCode != null) {
      return retryStatusCodes.contains(err.response!.statusCode);
    }

    return false;
  }

  int _getRetryCount(RequestOptions options) {
    return options.extra['retry_count'] ?? 0;
  }

  void _incrementRetryCount(RequestOptions options) {
    final currentCount = _getRetryCount(options);
    options.extra['retry_count'] = currentCount + 1;
  }

  Future<Response> _retry(RequestOptions requestOptions) async {
    final dio = Dio();
    
    final options = Options(
      method: requestOptions.method,
      headers: requestOptions.headers,
      extra: requestOptions.extra,
    );

    return dio.request(
      requestOptions.path,
      data: requestOptions.data,
      queryParameters: requestOptions.queryParameters,
      options: options,
    );
  }
}