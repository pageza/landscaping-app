import 'package:retrofit/retrofit.dart';
import 'package:dio/dio.dart';
import '../../../../shared/models/user.dart';
import '../../../../core/constants/app_constants.dart';

part 'auth_remote_datasource.g.dart';

@RestApi()
abstract class AuthRemoteDataSource {
  factory AuthRemoteDataSource(Dio dio) = _AuthRemoteDataSource;

  @POST(AppConstants.endpointLogin)
  Future<AuthResponse> login(@Body() LoginRequest request);

  @POST(AppConstants.endpointRegister)
  Future<AuthResponse> register(@Body() RegisterRequest request);

  @POST(AppConstants.endpointRefreshToken)
  Future<AuthResponse> refreshToken(@Body() RefreshTokenRequest request);

  @POST('${AppConstants.endpointAuth}/logout')
  Future<void> logout();

  @GET(AppConstants.endpointProfile)
  Future<User> getProfile();

  @PUT(AppConstants.endpointProfile)
  Future<User> updateProfile(@Body() UpdateProfileRequest request);

  @POST('${AppConstants.endpointAuth}/change-password')
  Future<void> changePassword(@Body() ChangePasswordRequest request);

  @POST('${AppConstants.endpointAuth}/forgot-password')
  Future<void> forgotPassword(@Body() ForgotPasswordRequest request);

  @POST('${AppConstants.endpointAuth}/reset-password')
  Future<void> resetPassword(@Body() ResetPasswordRequest request);

  @POST('${AppConstants.endpointAuth}/verify-email')
  Future<void> verifyEmail(@Body() Map<String, String> request);

  @POST('${AppConstants.endpointAuth}/resend-verification')
  Future<void> resendVerification();

  @DELETE('${AppConstants.endpointAuth}/delete-account')
  Future<void> deleteAccount(@Body() Map<String, String> request);
}