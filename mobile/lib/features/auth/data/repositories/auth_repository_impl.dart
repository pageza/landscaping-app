import 'package:dio/dio.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/auth_remote_datasource.dart';
import '../datasources/auth_local_datasource.dart';
import '../../../../shared/models/user.dart';
import '../../../../core/utils/app_exceptions.dart';
import '../../../../core/network/connectivity_service.dart';

class AuthRepositoryImpl implements AuthRepository {
  final AuthRemoteDataSource _remoteDataSource;
  final AuthLocalDataSource _localDataSource;
  final ConnectivityService _connectivityService;

  AuthRepositoryImpl({
    required AuthRemoteDataSource remoteDataSource,
    required AuthLocalDataSource localDataSource,
    required ConnectivityService connectivityService,
  }) : _remoteDataSource = remoteDataSource,
       _localDataSource = localDataSource,
       _connectivityService = connectivityService;

  @override
  Future<AuthResponse> login(LoginRequest request) async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      final response = await _remoteDataSource.login(request);
      
      // Save tokens and user data locally
      await _localDataSource.saveTokens(response.accessToken, response.refreshToken);
      await _localDataSource.saveUser(response.user);
      
      return response;
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Login failed: ${e.toString()}');
    }
  }

  @override
  Future<AuthResponse> register(RegisterRequest request) async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      final response = await _remoteDataSource.register(request);
      
      // Save tokens and user data locally
      await _localDataSource.saveTokens(response.accessToken, response.refreshToken);
      await _localDataSource.saveUser(response.user);
      
      return response;
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Registration failed: ${e.toString()}');
    }
  }

  @override
  Future<AuthResponse> refreshToken() async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      final refreshToken = await _localDataSource.getRefreshToken();
      if (refreshToken == null) {
        throw AuthException('No refresh token available');
      }

      final request = RefreshTokenRequest(refreshToken: refreshToken);
      final response = await _remoteDataSource.refreshToken(request);
      
      // Update tokens locally
      await _localDataSource.saveTokens(response.accessToken, response.refreshToken);
      await _localDataSource.saveUser(response.user);
      
      return response;
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Token refresh failed: ${e.toString()}');
    }
  }

  @override
  Future<void> logout() async {
    try {
      if (_connectivityService.isConnected) {
        // Try to logout on server
        try {
          await _remoteDataSource.logout();
        } catch (e) {
          // Continue with local logout even if server logout fails
        }
      }
      
      // Clear local data
      await _localDataSource.clearTokens();
      await _localDataSource.clearUser();
    } catch (e) {
      throw AppException('Logout failed: ${e.toString()}');
    }
  }

  @override
  Future<User> getProfile() async {
    try {
      if (_connectivityService.isConnected) {
        // Try to get from server first
        try {
          final user = await _remoteDataSource.getProfile();
          await _localDataSource.saveUser(user);
          return user;
        } catch (e) {
          // Fall back to local data if server fails
        }
      }
      
      // Get from local storage
      final user = await _localDataSource.getUser();
      if (user == null) {
        throw AuthException('No user data available');
      }
      
      return user;
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Failed to get profile: ${e.toString()}');
    }
  }

  @override
  Future<User> updateProfile(UpdateProfileRequest request) async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      final user = await _remoteDataSource.updateProfile(request);
      await _localDataSource.saveUser(user);
      
      return user;
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Profile update failed: ${e.toString()}');
    }
  }

  @override
  Future<void> changePassword(ChangePasswordRequest request) async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      await _remoteDataSource.changePassword(request);
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Password change failed: ${e.toString()}');
    }
  }

  @override
  Future<void> forgotPassword(ForgotPasswordRequest request) async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      await _remoteDataSource.forgotPassword(request);
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Forgot password request failed: ${e.toString()}');
    }
  }

  @override
  Future<void> resetPassword(ResetPasswordRequest request) async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      await _remoteDataSource.resetPassword(request);
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Password reset failed: ${e.toString()}');
    }
  }

  @override
  Future<bool> isLoggedIn() async {
    try {
      final accessToken = await _localDataSource.getAccessToken();
      final user = await _localDataSource.getUser();
      return accessToken != null && user != null;
    } catch (e) {
      return false;
    }
  }

  @override
  Future<String?> getAccessToken() async {
    return await _localDataSource.getAccessToken();
  }

  @override
  Future<User?> getCurrentUser() async {
    return await _localDataSource.getUser();
  }

  @override
  Future<void> saveBiometricEnabled(bool enabled) async {
    await _localDataSource.saveBiometricEnabled(enabled);
  }

  @override
  Future<bool> getBiometricEnabled() async {
    return await _localDataSource.getBiometricEnabled();
  }

  @override
  Future<void> saveOnboardingCompleted(bool completed) async {
    await _localDataSource.saveOnboardingCompleted(completed);
  }

  @override
  Future<bool> getOnboardingCompleted() async {
    return await _localDataSource.getOnboardingCompleted();
  }

  @override
  Future<void> verifyEmail(String token) async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      await _remoteDataSource.verifyEmail({'token': token});
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Email verification failed: ${e.toString()}');
    }
  }

  @override
  Future<void> resendVerification() async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      await _remoteDataSource.resendVerification();
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Resend verification failed: ${e.toString()}');
    }
  }

  @override
  Future<void> deleteAccount(String password) async {
    try {
      if (!_connectivityService.isConnected) {
        throw NetworkException('No internet connection');
      }

      await _remoteDataSource.deleteAccount({'password': password});
      
      // Clear local data
      await _localDataSource.clearAll();
    } on DioException catch (e) {
      throw _handleDioException(e);
    } catch (e) {
      throw AppException('Account deletion failed: ${e.toString()}');
    }
  }

  AppException _handleDioException(DioException e) {
    if (e.error is AppException) {
      return e.error as AppException;
    }
    
    switch (e.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return NetworkException('Connection timeout');
      case DioExceptionType.connectionError:
        return NetworkException('Connection error');
      case DioExceptionType.badResponse:
        final statusCode = e.response?.statusCode;
        final message = e.response?.data?['message'] ?? 'Request failed';
        
        switch (statusCode) {
          case 401:
            return AuthException(message);
          case 422:
            return ValidationException(message);
          case 500:
            return ServerException(message);
          default:
            return AppException(message);
        }
      default:
        return AppException('Request failed: ${e.message}');
    }
  }
}