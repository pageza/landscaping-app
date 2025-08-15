import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:freezed_annotation/freezed_annotation.dart';
import '../../../../shared/models/user.dart';
import '../../domain/usecases/login_usecase.dart';
import '../../../../core/utils/app_exceptions.dart';
import '../../../../shared/providers/app_providers.dart';

part 'auth_provider.freezed.dart';

@freezed
class AuthState with _$AuthState {
  const factory AuthState.initial() = _Initial;
  const factory AuthState.loading() = _Loading;
  const factory AuthState.authenticated(User user) = _Authenticated;
  const factory AuthState.unauthenticated() = _Unauthenticated;
  const factory AuthState.error(String message) = _Error;
}

class AuthNotifier extends StateNotifier<AuthState> {
  final LoginUseCase _loginUseCase;
  final RegisterUseCase _registerUseCase;
  final LogoutUseCase _logoutUseCase;
  final GetProfileUseCase _getProfileUseCase;
  final IsLoggedInUseCase _isLoggedInUseCase;

  AuthNotifier({
    required LoginUseCase loginUseCase,
    required RegisterUseCase registerUseCase,
    required LogoutUseCase logoutUseCase,
    required GetProfileUseCase getProfileUseCase,
    required IsLoggedInUseCase isLoggedInUseCase,
  })  : _loginUseCase = loginUseCase,
        _registerUseCase = registerUseCase,
        _logoutUseCase = logoutUseCase,
        _getProfileUseCase = getProfileUseCase,
        _isLoggedInUseCase = isLoggedInUseCase,
        super(const AuthState.initial());

  Future<void> checkAuthStatus() async {
    try {
      state = const AuthState.loading();
      
      final isLoggedIn = await _isLoggedInUseCase();
      if (isLoggedIn) {
        final user = await _getProfileUseCase();
        state = AuthState.authenticated(user);
      } else {
        state = const AuthState.unauthenticated();
      }
    } catch (e) {
      state = AuthState.error(ExceptionHandler.getErrorMessage(e));
    }
  }

  Future<bool> login(String email, String password) async {
    try {
      state = const AuthState.loading();
      
      final request = LoginRequest(
        email: email,
        password: password,
      );
      
      final response = await _loginUseCase(request);
      state = AuthState.authenticated(response.user);
      
      return true;
    } catch (e) {
      state = AuthState.error(ExceptionHandler.getErrorMessage(e));
      return false;
    }
  }

  Future<bool> register({
    required String email,
    required String password,
    required String firstName,
    required String lastName,
    required String userType,
    String? phone,
    String? companyName,
  }) async {
    try {
      state = const AuthState.loading();
      
      final request = RegisterRequest(
        email: email,
        password: password,
        firstName: firstName,
        lastName: lastName,
        userType: userType,
        phone: phone,
        companyName: companyName,
      );
      
      final response = await _registerUseCase(request);
      state = AuthState.authenticated(response.user);
      
      return true;
    } catch (e) {
      state = AuthState.error(ExceptionHandler.getErrorMessage(e));
      return false;
    }
  }

  Future<void> logout() async {
    try {
      await _logoutUseCase();
      state = const AuthState.unauthenticated();
    } catch (e) {
      // Even if logout fails on server, clear local state
      state = const AuthState.unauthenticated();
    }
  }

  Future<void> refreshProfile() async {
    try {
      final user = await _getProfileUseCase();
      state = AuthState.authenticated(user);
    } catch (e) {
      state = AuthState.error(ExceptionHandler.getErrorMessage(e));
    }
  }

  void clearError() {
    state.maybeWhen(
      error: (_) => state = const AuthState.unauthenticated(),
      orElse: () {},
    );
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(
    loginUseCase: ref.watch(loginUseCaseProvider),
    registerUseCase: ref.watch(registerUseCaseProvider),
    logoutUseCase: ref.watch(logoutUseCaseProvider),
    getProfileUseCase: ref.watch(getProfileUseCaseProvider),
    isLoggedInUseCase: ref.watch(isLoggedInUseCaseProvider),
  );
});

// Helper providers for specific auth data
final currentUserProvider = Provider<User?>((ref) {
  return ref.watch(authProvider).maybeWhen(
    authenticated: (user) => user,
    orElse: () => null,
  );
});

final isAuthenticatedProvider = Provider<bool>((ref) {
  return ref.watch(authProvider).maybeWhen(
    authenticated: (_) => true,
    orElse: () => false,
  );
});

final isLoadingProvider = Provider<bool>((ref) {
  return ref.watch(authProvider).maybeWhen(
    loading: () => true,
    orElse: () => false,
  );
});

final authErrorProvider = Provider<String?>((ref) {
  return ref.watch(authProvider).maybeWhen(
    error: (message) => message,
    orElse: () => null,
  );
});