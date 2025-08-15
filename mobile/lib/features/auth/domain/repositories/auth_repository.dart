import '../../../../shared/models/user.dart';

abstract class AuthRepository {
  Future<AuthResponse> login(LoginRequest request);
  Future<AuthResponse> register(RegisterRequest request);
  Future<AuthResponse> refreshToken();
  Future<void> logout();
  
  Future<User> getProfile();
  Future<User> updateProfile(UpdateProfileRequest request);
  
  Future<void> changePassword(ChangePasswordRequest request);
  Future<void> forgotPassword(ForgotPasswordRequest request);
  Future<void> resetPassword(ResetPasswordRequest request);
  
  Future<bool> isLoggedIn();
  Future<String?> getAccessToken();
  Future<User?> getCurrentUser();
  
  Future<void> saveBiometricEnabled(bool enabled);
  Future<bool> getBiometricEnabled();
  
  Future<void> saveOnboardingCompleted(bool completed);
  Future<bool> getOnboardingCompleted();
  
  Future<void> verifyEmail(String token);
  Future<void> resendVerification();
  
  Future<void> deleteAccount(String password);
}