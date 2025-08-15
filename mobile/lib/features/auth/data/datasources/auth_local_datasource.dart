import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../../../../shared/models/user.dart';
import '../../../../core/constants/app_constants.dart';
import '../../../../core/utils/app_exceptions.dart';
import 'dart:convert';

abstract class AuthLocalDataSource {
  Future<void> saveTokens(String accessToken, String refreshToken);
  Future<String?> getAccessToken();
  Future<String?> getRefreshToken();
  Future<void> clearTokens();
  Future<void> saveUser(User user);
  Future<User?> getUser();
  Future<void> clearUser();
  Future<void> saveBiometricEnabled(bool enabled);
  Future<bool> getBiometricEnabled();
  Future<void> saveOnboardingCompleted(bool completed);
  Future<bool> getOnboardingCompleted();
  Future<void> clearAll();
}

class AuthLocalDataSourceImpl implements AuthLocalDataSource {
  final FlutterSecureStorage _secureStorage;
  final SharedPreferences _prefs;

  AuthLocalDataSourceImpl({
    required FlutterSecureStorage secureStorage,
    required SharedPreferences prefs,
  }) : _secureStorage = secureStorage, _prefs = prefs;

  @override
  Future<void> saveTokens(String accessToken, String refreshToken) async {
    try {
      await Future.wait([
        _secureStorage.write(key: AppConstants.keyAccessToken, value: accessToken),
        _secureStorage.write(key: AppConstants.keyRefreshToken, value: refreshToken),
      ]);
    } catch (e) {
      throw CacheException('Failed to save tokens: ${e.toString()}');
    }
  }

  @override
  Future<String?> getAccessToken() async {
    try {
      return await _secureStorage.read(key: AppConstants.keyAccessToken);
    } catch (e) {
      throw CacheException('Failed to get access token: ${e.toString()}');
    }
  }

  @override
  Future<String?> getRefreshToken() async {
    try {
      return await _secureStorage.read(key: AppConstants.keyRefreshToken);
    } catch (e) {
      throw CacheException('Failed to get refresh token: ${e.toString()}');
    }
  }

  @override
  Future<void> clearTokens() async {
    try {
      await Future.wait([
        _secureStorage.delete(key: AppConstants.keyAccessToken),
        _secureStorage.delete(key: AppConstants.keyRefreshToken),
      ]);
    } catch (e) {
      throw CacheException('Failed to clear tokens: ${e.toString()}');
    }
  }

  @override
  Future<void> saveUser(User user) async {
    try {
      final userJson = jsonEncode(user.toJson());
      await Future.wait([
        _secureStorage.write(key: AppConstants.keyUserId, value: user.id),
        _prefs.setString('user_data', userJson),
        _prefs.setString(AppConstants.keyUserType, user.userType),
      ]);
    } catch (e) {
      throw CacheException('Failed to save user: ${e.toString()}');
    }
  }

  @override
  Future<User?> getUser() async {
    try {
      final userJson = _prefs.getString('user_data');
      if (userJson == null) return null;
      
      final userMap = jsonDecode(userJson) as Map<String, dynamic>;
      return User.fromJson(userMap);
    } catch (e) {
      throw CacheException('Failed to get user: ${e.toString()}');
    }
  }

  @override
  Future<void> clearUser() async {
    try {
      await Future.wait([
        _secureStorage.delete(key: AppConstants.keyUserId),
        _prefs.remove('user_data'),
        _prefs.remove(AppConstants.keyUserType),
      ]);
    } catch (e) {
      throw CacheException('Failed to clear user: ${e.toString()}');
    }
  }

  @override
  Future<void> saveBiometricEnabled(bool enabled) async {
    try {
      await _prefs.setBool(AppConstants.keyBiometricEnabled, enabled);
    } catch (e) {
      throw CacheException('Failed to save biometric setting: ${e.toString()}');
    }
  }

  @override
  Future<bool> getBiometricEnabled() async {
    try {
      return _prefs.getBool(AppConstants.keyBiometricEnabled) ?? false;
    } catch (e) {
      throw CacheException('Failed to get biometric setting: ${e.toString()}');
    }
  }

  @override
  Future<void> saveOnboardingCompleted(bool completed) async {
    try {
      await _prefs.setBool(AppConstants.keyOnboardingCompleted, completed);
    } catch (e) {
      throw CacheException('Failed to save onboarding status: ${e.toString()}');
    }
  }

  @override
  Future<bool> getOnboardingCompleted() async {
    try {
      return _prefs.getBool(AppConstants.keyOnboardingCompleted) ?? false;
    } catch (e) {
      throw CacheException('Failed to get onboarding status: ${e.toString()}');
    }
  }

  @override
  Future<void> clearAll() async {
    try {
      await Future.wait([
        _secureStorage.deleteAll(),
        _prefs.clear(),
      ]);
    } catch (e) {
      throw CacheException('Failed to clear all data: ${e.toString()}');
    }
  }
}