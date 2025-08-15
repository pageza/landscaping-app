import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../core/network/api_client.dart';
import '../../core/network/connectivity_service.dart';
import '../../features/auth/data/datasources/auth_remote_datasource.dart';
import '../../features/auth/data/datasources/auth_local_datasource.dart';
import '../../features/auth/data/repositories/auth_repository_impl.dart';
import '../../features/auth/domain/repositories/auth_repository.dart';
import '../../features/auth/domain/usecases/login_usecase.dart';
import '../../features/auth/presentation/providers/auth_provider.dart';
import '../../features/auth/presentation/pages/login_page.dart';
import '../../features/auth/presentation/pages/register_page.dart';
import '../../features/auth/presentation/pages/splash_page.dart';
import '../../features/customer/presentation/pages/customer_home_page.dart';
import '../../features/crew/presentation/pages/crew_home_page.dart';

// Core Providers
final sharedPreferencesProvider = Provider<SharedPreferences>((ref) {
  throw UnimplementedError('SharedPreferences must be overridden');
});

final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage(
    aOptions: AndroidOptions(
      encryptedSharedPreferences: true,
    ),
    iOptions: IOSOptions(
      accessibility: KeychainAccessibility.first_unlock_this_device,
    ),
  );
});

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient();
});

final connectivityServiceProvider = Provider<ConnectivityService>((ref) {
  return ConnectivityService();
});

// Data Source Providers
final authRemoteDataSourceProvider = Provider<AuthRemoteDataSource>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return AuthRemoteDataSource(apiClient.dio);
});

final authLocalDataSourceProvider = Provider<AuthLocalDataSource>((ref) {
  final secureStorage = ref.watch(secureStorageProvider);
  final prefs = ref.watch(sharedPreferencesProvider);
  return AuthLocalDataSourceImpl(
    secureStorage: secureStorage,
    prefs: prefs,
  );
});

// Repository Providers
final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final remoteDataSource = ref.watch(authRemoteDataSourceProvider);
  final localDataSource = ref.watch(authLocalDataSourceProvider);
  final connectivityService = ref.watch(connectivityServiceProvider);
  
  return AuthRepositoryImpl(
    remoteDataSource: remoteDataSource,
    localDataSource: localDataSource,
    connectivityService: connectivityService,
  );
});

// Use Case Providers
final loginUseCaseProvider = Provider<LoginUseCase>((ref) {
  final repository = ref.watch(authRepositoryProvider);
  return LoginUseCase(repository);
});

final registerUseCaseProvider = Provider<RegisterUseCase>((ref) {
  final repository = ref.watch(authRepositoryProvider);
  return RegisterUseCase(repository);
});

final logoutUseCaseProvider = Provider<LogoutUseCase>((ref) {
  final repository = ref.watch(authRepositoryProvider);
  return LogoutUseCase(repository);
});

final getProfileUseCaseProvider = Provider<GetProfileUseCase>((ref) {
  final repository = ref.watch(authRepositoryProvider);
  return GetProfileUseCase(repository);
});

final isLoggedInUseCaseProvider = Provider<IsLoggedInUseCase>((ref) {
  final repository = ref.watch(authRepositoryProvider);
  return IsLoggedInUseCase(repository);
});

// Router Provider
final routerProvider = Provider<GoRouter>((ref) {
  // final authNotifier = ref.watch(authProvider.notifier);
  
  return GoRouter(
    initialLocation: '/splash',
    redirect: (context, state) {
      final isLoggedIn = ref.read(authProvider).maybeWhen(
        authenticated: (_) => true,
        orElse: () => false,
      );
      
      final isOnAuthPage = state.fullPath == '/login' || 
                          state.fullPath == '/register' ||
                          state.fullPath == '/splash';
      
      // If not logged in and not on auth page, redirect to login
      if (!isLoggedIn && !isOnAuthPage) {
        return '/login';
      }
      
      // If logged in and on auth page, redirect to appropriate home
      if (isLoggedIn && isOnAuthPage && state.fullPath != '/splash') {
        final user = ref.read(authProvider).maybeWhen(
          authenticated: (user) => user,
          orElse: () => null,
        );
        
        if (user?.userType == 'customer') {
          return '/customer';
        } else if (user?.userType == 'crew') {
          return '/crew';
        }
      }
      
      return null;
    },
    routes: [
      GoRoute(
        path: '/splash',
        name: 'splash',
        builder: (context, state) => const SplashPage(),
      ),
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (context, state) => const LoginPage(),
      ),
      GoRoute(
        path: '/register',
        name: 'register',
        builder: (context, state) => const RegisterPage(),
      ),
      GoRoute(
        path: '/customer',
        name: 'customer_home',
        builder: (context, state) => const CustomerHomePage(),
        routes: [
          // Customer routes will be added here
        ],
      ),
      GoRoute(
        path: '/crew',
        name: 'crew_home',
        builder: (context, state) => const CrewHomePage(),
        routes: [
          // Crew routes will be added here
        ],
      ),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.error_outline,
              size: 64,
              color: Colors.red,
            ),
            const SizedBox(height: 16),
            Text(
              'Page not found',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              state.fullPath ?? 'Unknown location',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () => context.go('/login'),
              child: const Text('Go to Login'),
            ),
          ],
        ),
      ),
    ),
  );
});

// Theme Provider
final themeProvider = StateNotifierProvider<ThemeNotifier, ThemeMode>((ref) {
  return ThemeNotifier();
});

class ThemeNotifier extends StateNotifier<ThemeMode> {
  ThemeNotifier() : super(ThemeMode.system);

  void setTheme(ThemeMode theme) {
    state = theme;
  }

  void toggleTheme() {
    state = state == ThemeMode.light ? ThemeMode.dark : ThemeMode.light;
  }
}