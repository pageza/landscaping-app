import 'dart:io';

import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:mobile/main.dart' as app;
import 'package:mobile/features/auth/presentation/pages/login_page.dart';
import 'package:mobile/features/auth/presentation/pages/register_page.dart';
import 'package:mobile/features/customer/presentation/pages/customer_home_page.dart';
import 'package:mobile/features/crew/presentation/pages/crew_home_page.dart';
import 'package:mobile/shared/widgets/custom_button.dart';
import 'package:mobile/shared/widgets/custom_text_field.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('Authentication Flow Integration Tests', () {
    testWidgets('complete login flow for customer', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Should start on splash/login screen
      expect(find.byType(LoginPage), findsOneWidget);

      // Enter email
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'customer@example.com',
      );

      // Enter password
      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      // Tap login button
      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // Should navigate to customer home page
      expect(find.byType(CustomerHomePage), findsOneWidget);
      
      // Verify UI elements are present
      expect(find.text('Welcome'), findsOneWidget);
      expect(find.byKey(const Key('customer_dashboard')), findsOneWidget);
    });

    testWidgets('complete login flow for crew member', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Login as crew member
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'crew@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // Should navigate to crew home page
      expect(find.byType(CrewHomePage), findsOneWidget);
      
      // Verify crew-specific UI
      expect(find.text('Today\'s Jobs'), findsOneWidget);
      expect(find.byKey(const Key('jobs_list')), findsOneWidget);
    });

    testWidgets('handles login with invalid credentials', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Enter invalid credentials
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'invalid@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'wrongpassword',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Should stay on login page and show error
      expect(find.byType(LoginPage), findsOneWidget);
      expect(find.text('Invalid email or password'), findsOneWidget);
    });

    testWidgets('registration flow', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Tap register link
      await tester.tap(find.byKey(const Key('register_link')));
      await tester.pumpAndSettle();

      expect(find.byType(RegisterPage), findsOneWidget);

      // Fill registration form
      await tester.enterText(
        find.byKey(const Key('first_name_field')),
        'John',
      );

      await tester.enterText(
        find.byKey(const Key('last_name_field')),
        'Doe',
      );

      await tester.enterText(
        find.byKey(const Key('email_field')),
        'john.doe@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'SecurePass123!',
      );

      await tester.enterText(
        find.byKey(const Key('confirm_password_field')),
        'SecurePass123!',
      );

      await tester.enterText(
        find.byKey(const Key('company_field')),
        'Doe Landscaping',
      );

      // Scroll to register button if needed
      await tester.scrollUntilVisible(
        find.byKey(const Key('register_button')),
        500.0,
      );

      await tester.tap(find.byKey(const Key('register_button')));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // Should show success message or navigate to verification
      expect(
        find.text('Registration successful') || 
        find.text('Please check your email'),
        findsOneWidget,
      );
    });

    testWidgets('password validation during registration', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      await tester.tap(find.byKey(const Key('register_link')));
      await tester.pumpAndSettle();

      // Fill form with weak password
      await tester.enterText(
        find.byKey(const Key('first_name_field')),
        'Test',
      );

      await tester.enterText(
        find.byKey(const Key('email_field')),
        'test@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        '123', // Weak password
      );

      await tester.enterText(
        find.byKey(const Key('confirm_password_field')),
        '456', // Non-matching password
      );

      await tester.tap(find.byKey(const Key('register_button')));
      await tester.pumpAndSettle();

      // Should show validation errors
      expect(find.text('Password must be at least 8 characters'), findsOneWidget);
      expect(find.text('Passwords do not match'), findsOneWidget);
    });

    testWidgets('forgot password flow', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Tap forgot password link
      await tester.tap(find.byKey(const Key('forgot_password_link')));
      await tester.pumpAndSettle();

      // Should navigate to forgot password page
      expect(find.text('Reset Password'), findsOneWidget);

      // Enter email
      await tester.enterText(
        find.byKey(const Key('reset_email_field')),
        'user@example.com',
      );

      await tester.tap(find.byKey(const Key('send_reset_button')));
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Should show success message
      expect(find.text('Reset link sent to your email'), findsOneWidget);
    });

    testWidgets('logout flow', (WidgetTester tester) async {
      // First login
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      await tester.enterText(
        find.byKey(const Key('email_field')),
        'customer@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // Navigate to profile/settings
      await tester.tap(find.byKey(const Key('profile_menu')));
      await tester.pumpAndSettle();

      // Tap logout
      await tester.tap(find.byKey(const Key('logout_button')));
      await tester.pumpAndSettle();

      // Should show confirmation dialog
      expect(find.text('Are you sure you want to logout?'), findsOneWidget);

      await tester.tap(find.text('Yes'));
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Should return to login page
      expect(find.byType(LoginPage), findsOneWidget);
    });
  });

  group('Network Connectivity Tests', () {
    testWidgets('handles offline login attempt', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Simulate network disconnection
      await tester.binding.defaultBinaryMessenger.handlePlatformMessage(
        'flutter/connectivity',
        const StandardMethodCodec().encodeMethodCall(
          const MethodCall('networkStateChanged', 'none'),
        ),
        (data) {},
      );

      await tester.enterText(
        find.byKey(const Key('email_field')),
        'user@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Should show offline message
      expect(find.text('No internet connection'), findsOneWidget);
    });

    testWidgets('handles slow network during login', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      await tester.enterText(
        find.byKey(const Key('email_field')),
        'user@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      
      // Should show loading indicator immediately
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      
      await tester.pumpAndSettle(const Duration(seconds: 10));
      
      // Should eventually complete or show timeout
      expect(
        find.byType(CustomerHomePage) || find.text('Request timeout'),
        findsOneWidget,
      );
    });
  });

  group('Authentication State Persistence', () {
    testWidgets('remembers login state after app restart', (WidgetTester tester) async {
      // First launch - login
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      await tester.enterText(
        find.byKey(const Key('email_field')),
        'customer@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      expect(find.byType(CustomerHomePage), findsOneWidget);

      // Simulate app restart
      await tester.binding.reassembleApplication();
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Should go directly to home page (logged in)
      expect(find.byType(CustomerHomePage), findsOneWidget);
    });

    testWidgets('handles expired token gracefully', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Login first
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'customer@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // Simulate token expiration by making an API call
      await tester.tap(find.byKey(const Key('refresh_button')));
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Should redirect to login if token is expired
      expect(
        find.byType(LoginPage) || find.byType(CustomerHomePage),
        findsOneWidget,
      );
    });
  });

  group('Accessibility Tests', () {
    testWidgets('login form is accessible', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Check for semantic labels
      expect(find.bySemanticsLabel('Email address'), findsOneWidget);
      expect(find.bySemanticsLabel('Password'), findsOneWidget);
      expect(find.bySemanticsLabel('Login'), findsOneWidget);
      
      // Check for proper tab order
      await tester.sendKeyEvent(LogicalKeyboardKey.tab);
      await tester.pumpAndSettle();
      
      // Should focus on email field first
      final emailField = tester.widget<TextField>(
        find.byKey(const Key('email_field')),
      );
      expect(emailField.focusNode?.hasFocus, isTrue);
    });

    testWidgets('supports screen reader navigation', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Simulate screen reader navigation
      final SemanticsHandle handle = tester.ensureSemantics();
      
      // Should have proper semantic tree
      expect(find.byType(Semantics), findsWidgets);
      
      handle.dispose();
    });
  });

  group('Biometric Authentication Tests', () {
    testWidgets('enables biometric login after first login', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Login normally first
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'customer@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // Should prompt to enable biometric login
      if (Platform.isAndroid || Platform.isIOS) {
        expect(find.text('Enable biometric login?'), findsOneWidget);
        
        await tester.tap(find.text('Yes'));
        await tester.pumpAndSettle();
      }
    });

    testWidgets('biometric login flow', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      if (Platform.isAndroid || Platform.isIOS) {
        // Should show biometric login option
        expect(find.byKey(const Key('biometric_login_button')), findsOneWidget);

        await tester.tap(find.byKey(const Key('biometric_login_button')));
        await tester.pumpAndSettle();

        // Would need to mock biometric authentication
        // For now, just verify the UI elements
        expect(find.text('Use your fingerprint to login'), findsOneWidget);
      }
    });
  });

  group('Multi-language Support', () {
    testWidgets('switches language successfully', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Tap language selector
      await tester.tap(find.byKey(const Key('language_selector')));
      await tester.pumpAndSettle();

      // Select Spanish
      await tester.tap(find.text('Español'));
      await tester.pumpAndSettle();

      // UI should be in Spanish
      expect(find.text('Iniciar sesión'), findsOneWidget);
      expect(find.text('Correo electrónico'), findsOneWidget);
      expect(find.text('Contraseña'), findsOneWidget);
    });
  });

  group('Error Scenarios', () {
    testWidgets('handles server errors gracefully', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // This would require mocking server responses
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'server-error@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 5));

      // Should show appropriate error message
      expect(find.text('Server is temporarily unavailable'), findsOneWidget);
    });

    testWidgets('validates form fields', (WidgetTester tester) async {
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      // Try to login with empty fields
      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle();

      // Should show validation errors
      expect(find.text('Email is required'), findsOneWidget);
      expect(find.text('Password is required'), findsOneWidget);

      // Invalid email format
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'invalid-email',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle();

      expect(find.text('Enter a valid email address'), findsOneWidget);
    });
  });

  group('Performance Tests', () {
    testWidgets('login performance is acceptable', (WidgetTester tester) async {
      final Stopwatch stopwatch = Stopwatch()..start();
      
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 3));

      await tester.enterText(
        find.byKey(const Key('email_field')),
        'customer@example.com',
      );

      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle(const Duration(seconds: 10));

      stopwatch.stop();

      // Login should complete within reasonable time
      expect(stopwatch.elapsedMilliseconds, lessThan(10000));
    });

    testWidgets('app startup time is acceptable', (WidgetTester tester) async {
      final Stopwatch stopwatch = Stopwatch()..start();
      
      app.main();
      await tester.pumpAndSettle(const Duration(seconds: 5));
      
      stopwatch.stop();

      // App should start within 5 seconds
      expect(stopwatch.elapsedMilliseconds, lessThan(5000));
      expect(find.byType(LoginPage), findsOneWidget);
    });
  });
}