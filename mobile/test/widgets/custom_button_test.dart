import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/shared/widgets/custom_button.dart';

void main() {
  group('CustomButton Widget Tests', () {
    testWidgets('renders with default properties', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Test Button',
              onPressed: () {},
            ),
          ),
        ),
      );

      // Find the button
      expect(find.byType(CustomButton), findsOneWidget);
      expect(find.text('Test Button'), findsOneWidget);

      // Check default styling
      final ElevatedButton button = tester.widget(find.byType(ElevatedButton));
      expect(button.onPressed, isNotNull);
    });

    testWidgets('handles tap events', (WidgetTester tester) async {
      bool wasTapped = false;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Tap Me',
              onPressed: () {
                wasTapped = true;
              },
            ),
          ),
        ),
      );

      // Tap the button
      await tester.tap(find.byType(CustomButton));
      await tester.pumpAndSettle();

      expect(wasTapped, isTrue);
    });

    testWidgets('shows loading state', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Loading Button',
              onPressed: () {},
              isLoading: true,
            ),
          ),
        ),
      );

      // Should show loading indicator instead of text
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Loading Button'), findsNothing);

      // Button should be disabled when loading
      final ElevatedButton button = tester.widget(find.byType(ElevatedButton));
      expect(button.onPressed, isNull);
    });

    testWidgets('disabled state', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Disabled Button',
              onPressed: null, // Disabled button
            ),
          ),
        ),
      );

      final ElevatedButton button = tester.widget(find.byType(ElevatedButton));
      expect(button.onPressed, isNull);

      // Verify disabled styling
      expect(find.text('Disabled Button'), findsOneWidget);
    });

    testWidgets('custom styling applied', (WidgetTester tester) async {
      const customColor = Colors.red;
      
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Custom Button',
              onPressed: () {},
              backgroundColor: customColor,
              textColor: Colors.white,
              borderRadius: 12.0,
            ),
          ),
        ),
      );

      // Find the button and check its properties
      final CustomButton customButton = tester.widget(find.byType(CustomButton));
      expect(customButton.backgroundColor, equals(customColor));
      expect(customButton.textColor, equals(Colors.white));
      expect(customButton.borderRadius, equals(12.0));
    });

    testWidgets('different button variants', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Column(
              children: [
                CustomButton(
                  text: 'Primary',
                  onPressed: () {},
                  variant: ButtonVariant.primary,
                ),
                CustomButton(
                  text: 'Secondary',
                  onPressed: () {},
                  variant: ButtonVariant.secondary,
                ),
                CustomButton(
                  text: 'Outline',
                  onPressed: () {},
                  variant: ButtonVariant.outline,
                ),
              ],
            ),
          ),
        ),
      );

      expect(find.text('Primary'), findsOneWidget);
      expect(find.text('Secondary'), findsOneWidget);
      expect(find.text('Outline'), findsOneWidget);
    });

    testWidgets('icon button variant', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'With Icon',
              onPressed: () {},
              icon: Icons.add,
            ),
          ),
        ),
      );

      expect(find.byIcon(Icons.add), findsOneWidget);
      expect(find.text('With Icon'), findsOneWidget);
    });

    testWidgets('full width button', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SizedBox(
              width: 300,
              child: CustomButton(
                text: 'Full Width',
                onPressed: () {},
                fullWidth: true,
              ),
            ),
          ),
        ),
      );

      // Button should take full available width
      final Size buttonSize = tester.getSize(find.byType(CustomButton));
      expect(buttonSize.width, equals(300.0));
    });

    testWidgets('accessibility properties', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Accessible Button',
              onPressed: () {},
              semanticLabel: 'Custom accessibility label',
              tooltip: 'Button tooltip',
            ),
          ),
        ),
      );

      // Check semantic properties
      final Semantics semantics = tester.widget(
        find.descendant(
          of: find.byType(CustomButton),
          matching: find.byType(Semantics),
        ),
      );
      
      expect(semantics.properties.label, contains('Custom accessibility label'));
    });

    testWidgets('button sizing variants', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Column(
              children: [
                CustomButton(
                  text: 'Small',
                  onPressed: () {},
                  size: ButtonSize.small,
                ),
                CustomButton(
                  text: 'Medium',
                  onPressed: () {},
                  size: ButtonSize.medium,
                ),
                CustomButton(
                  text: 'Large',
                  onPressed: () {},
                  size: ButtonSize.large,
                ),
              ],
            ),
          ),
        ),
      );

      final Size smallButton = tester.getSize(find.text('Small'));
      final Size mediumButton = tester.getSize(find.text('Medium'));
      final Size largeButton = tester.getSize(find.text('Large'));

      expect(smallButton.height < mediumButton.height, isTrue);
      expect(mediumButton.height < largeButton.height, isTrue);
    });

    testWidgets('button animation on press', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: CustomButton(
              text: 'Animated Button',
              onPressed: () {},
              animateOnPress: true,
            ),
          ),
        ),
      );

      // Get initial transform
      final Matrix4 initialTransform = tester.widget<Transform>(
        find.descendant(
          of: find.byType(CustomButton),
          matching: find.byType(Transform),
        ),
      ).transform;

      // Press and hold the button
      final TestGesture gesture = await tester.startGesture(
        tester.getCenter(find.byType(CustomButton)),
      );

      await tester.pump(const Duration(milliseconds: 100));

      // Transform should have changed (scale down)
      final Matrix4 pressedTransform = tester.widget<Transform>(
        find.descendant(
          of: find.byType(CustomButton),
          matching: find.byType(Transform),
        ),
      ).transform;

      expect(pressedTransform != initialTransform, isTrue);

      // Release the button
      await gesture.up();
      await tester.pumpAndSettle();
    });

    group('Edge Cases', () {
      testWidgets('handles null onPressed gracefully', (WidgetTester tester) async {
        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: CustomButton(
                text: 'Null Handler',
                onPressed: null,
              ),
            ),
          ),
        );

        // Should render without error
        expect(find.byType(CustomButton), findsOneWidget);
        expect(find.text('Null Handler'), findsOneWidget);
      });

      testWidgets('handles empty text', (WidgetTester tester) async {
        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: CustomButton(
                text: '',
                onPressed: () {},
              ),
            ),
          ),
        );

        expect(find.byType(CustomButton), findsOneWidget);
      });

      testWidgets('handles very long text', (WidgetTester tester) async {
        const longText = 'This is a very long button text that should handle overflow properly';
        
        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: SizedBox(
                width: 200,
                child: CustomButton(
                  text: longText,
                  onPressed: () {},
                ),
              ),
            ),
          ),
        );

        expect(find.byType(CustomButton), findsOneWidget);
        expect(find.text(longText), findsOneWidget);
      });
    });

    group('Interaction Tests', () {
      testWidgets('debounces rapid taps', (WidgetTester tester) async {
        int tapCount = 0;

        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: CustomButton(
                text: 'Debounced Button',
                onPressed: () {
                  tapCount++;
                },
                debounceMs: 500,
              ),
            ),
          ),
        );

        // Tap rapidly multiple times
        await tester.tap(find.byType(CustomButton));
        await tester.pump(const Duration(milliseconds: 100));
        await tester.tap(find.byType(CustomButton));
        await tester.pump(const Duration(milliseconds: 100));
        await tester.tap(find.byType(CustomButton));
        
        // Wait for debounce to complete
        await tester.pump(const Duration(milliseconds: 600));

        // Should only register one tap due to debouncing
        expect(tapCount, equals(1));
      });

      testWidgets('handles gesture conflicts', (WidgetTester tester) async {
        bool wasPressed = false;

        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: GestureDetector(
                onTap: () {
                  // Outer gesture detector
                },
                child: CustomButton(
                  text: 'Nested Button',
                  onPressed: () {
                    wasPressed = true;
                  },
                ),
              ),
            ),
          ),
        );

        await tester.tap(find.byType(CustomButton));
        await tester.pumpAndSettle();

        // Button's onPressed should still work despite nested GestureDetector
        expect(wasPressed, isTrue);
      });
    });

    group('Theme Integration', () {
      testWidgets('respects theme colors', (WidgetTester tester) async {
        const primaryColor = Colors.purple;
        
        await tester.pumpWidget(
          MaterialApp(
            theme: ThemeData(
              primarySwatch: Colors.purple,
              elevatedButtonTheme: ElevatedButtonThemeData(
                style: ElevatedButton.styleFrom(
                  backgroundColor: primaryColor,
                ),
              ),
            ),
            home: Scaffold(
              body: CustomButton(
                text: 'Themed Button',
                onPressed: () {},
              ),
            ),
          ),
        );

        // Button should inherit theme colors when no custom color is specified
        expect(find.byType(CustomButton), findsOneWidget);
      });

      testWidgets('dark theme support', (WidgetTester tester) async {
        await tester.pumpWidget(
          MaterialApp(
            theme: ThemeData.dark(),
            home: Scaffold(
              body: CustomButton(
                text: 'Dark Theme Button',
                onPressed: () {},
              ),
            ),
          ),
        );

        expect(find.byType(CustomButton), findsOneWidget);
        expect(find.text('Dark Theme Button'), findsOneWidget);
      });
    });
  });
}