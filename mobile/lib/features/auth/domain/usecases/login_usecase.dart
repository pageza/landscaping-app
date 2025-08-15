import '../../../../shared/models/user.dart';
import '../repositories/auth_repository.dart';

class LoginUseCase {
  final AuthRepository _repository;

  LoginUseCase(this._repository);

  Future<AuthResponse> call(LoginRequest request) async {
    return await _repository.login(request);
  }
}

class RegisterUseCase {
  final AuthRepository _repository;

  RegisterUseCase(this._repository);

  Future<AuthResponse> call(RegisterRequest request) async {
    return await _repository.register(request);
  }
}

class LogoutUseCase {
  final AuthRepository _repository;

  LogoutUseCase(this._repository);

  Future<void> call() async {
    return await _repository.logout();
  }
}

class GetProfileUseCase {
  final AuthRepository _repository;

  GetProfileUseCase(this._repository);

  Future<User> call() async {
    return await _repository.getProfile();
  }
}

class UpdateProfileUseCase {
  final AuthRepository _repository;

  UpdateProfileUseCase(this._repository);

  Future<User> call(UpdateProfileRequest request) async {
    return await _repository.updateProfile(request);
  }
}

class ChangePasswordUseCase {
  final AuthRepository _repository;

  ChangePasswordUseCase(this._repository);

  Future<void> call(ChangePasswordRequest request) async {
    return await _repository.changePassword(request);
  }
}

class ForgotPasswordUseCase {
  final AuthRepository _repository;

  ForgotPasswordUseCase(this._repository);

  Future<void> call(ForgotPasswordRequest request) async {
    return await _repository.forgotPassword(request);
  }
}

class ResetPasswordUseCase {
  final AuthRepository _repository;

  ResetPasswordUseCase(this._repository);

  Future<void> call(ResetPasswordRequest request) async {
    return await _repository.resetPassword(request);
  }
}

class IsLoggedInUseCase {
  final AuthRepository _repository;

  IsLoggedInUseCase(this._repository);

  Future<bool> call() async {
    return await _repository.isLoggedIn();
  }
}

class GetCurrentUserUseCase {
  final AuthRepository _repository;

  GetCurrentUserUseCase(this._repository);

  Future<User?> call() async {
    return await _repository.getCurrentUser();
  }
}

class RefreshTokenUseCase {
  final AuthRepository _repository;

  RefreshTokenUseCase(this._repository);

  Future<AuthResponse> call() async {
    return await _repository.refreshToken();
  }
}

class VerifyEmailUseCase {
  final AuthRepository _repository;

  VerifyEmailUseCase(this._repository);

  Future<void> call(String token) async {
    return await _repository.verifyEmail(token);
  }
}

class ResendVerificationUseCase {
  final AuthRepository _repository;

  ResendVerificationUseCase(this._repository);

  Future<void> call() async {
    return await _repository.resendVerification();
  }
}

class DeleteAccountUseCase {
  final AuthRepository _repository;

  DeleteAccountUseCase(this._repository);

  Future<void> call(String password) async {
    return await _repository.deleteAccount(password);
  }
}

class BiometricUseCase {
  final AuthRepository _repository;

  BiometricUseCase(this._repository);

  Future<void> saveBiometricEnabled(bool enabled) async {
    return await _repository.saveBiometricEnabled(enabled);
  }

  Future<bool> getBiometricEnabled() async {
    return await _repository.getBiometricEnabled();
  }
}

class OnboardingUseCase {
  final AuthRepository _repository;

  OnboardingUseCase(this._repository);

  Future<void> saveOnboardingCompleted(bool completed) async {
    return await _repository.saveOnboardingCompleted(completed);
  }

  Future<bool> getOnboardingCompleted() async {
    return await _repository.getOnboardingCompleted();
  }
}