import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'shared/providers/app_providers.dart';

class AppInitializer extends ConsumerWidget {
  final Widget child;

  const AppInitializer({super.key, required this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return FutureBuilder<SharedPreferences>(
      future: SharedPreferences.getInstance(),
      builder: (context, snapshot) {
        if (snapshot.hasData) {
          return ProviderScope(
            overrides: [
              sharedPreferencesProvider.overrideWithValue(snapshot.data!),
            ],
            child: child,
          );
        }
        
        if (snapshot.hasError) {
          return MaterialApp(
            home: Scaffold(
              body: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.error, size: 64, color: Colors.red),
                    const SizedBox(height: 16),
                    Text('Failed to initialize app: ${snapshot.error}'),
                  ],
                ),
              ),
            ),
          );
        }
        
        return const MaterialApp(
          home: Scaffold(
            body: Center(
              child: CircularProgressIndicator(),
            ),
          ),
        );
      },
    );
  }
}