import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../constants/app_colors.dart';

class AppTheme {
  static ThemeData get lightTheme {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.light,
      colorScheme: const ColorScheme.light(
        primary: AppColors.primary,
        primaryContainer: AppColors.primaryLight,
        secondary: AppColors.secondary,
        secondaryContainer: AppColors.secondaryLight,
        tertiary: AppColors.accent,
        surface: AppColors.surface,
        background: AppColors.background,
        error: AppColors.error,
        onPrimary: AppColors.textOnPrimary,
        onSecondary: AppColors.textOnSecondary,
        onSurface: AppColors.textOnSurface,
        onBackground: AppColors.textPrimary,
        onError: AppColors.textOnPrimary,
        outline: AppColors.border,
        shadow: AppColors.shadow,
      ),
      textTheme: _buildTextTheme(AppColors.textPrimary),
      appBarTheme: _buildAppBarTheme(true),
      cardTheme: _buildCardTheme(true),
      elevatedButtonTheme: _buildElevatedButtonTheme(true),
      outlinedButtonTheme: _buildOutlinedButtonTheme(true),
      textButtonTheme: _buildTextButtonTheme(true),
      inputDecorationTheme: _buildInputDecorationTheme(true),
      chipTheme: _buildChipTheme(true),
      bottomNavigationBarTheme: _buildBottomNavigationBarTheme(true),
      tabBarTheme: _buildTabBarTheme(true),
      floatingActionButtonTheme: _buildFloatingActionButtonTheme(true),
      dividerTheme: const DividerThemeData(
        color: AppColors.divider,
        thickness: 1,
      ),
      scaffoldBackgroundColor: AppColors.background,
    );
  }

  static ThemeData get darkTheme {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      colorScheme: const ColorScheme.dark(
        primary: AppColors.primaryLight,
        primaryContainer: AppColors.primaryDark,
        secondary: AppColors.secondaryLight,
        secondaryContainer: AppColors.secondaryDark,
        tertiary: AppColors.accentLight,
        surface: AppColors.surfaceDark,
        background: AppColors.backgroundDark,
        error: AppColors.error,
        onPrimary: AppColors.textPrimary,
        onSecondary: AppColors.textPrimary,
        onSurface: AppColors.textPrimaryDark,
        onBackground: AppColors.textPrimaryDark,
        onError: AppColors.textOnPrimary,
        outline: AppColors.borderDark,
        shadow: AppColors.shadowDark,
      ),
      textTheme: _buildTextTheme(AppColors.textPrimaryDark),
      appBarTheme: _buildAppBarTheme(false),
      cardTheme: _buildCardTheme(false),
      elevatedButtonTheme: _buildElevatedButtonTheme(false),
      outlinedButtonTheme: _buildOutlinedButtonTheme(false),
      textButtonTheme: _buildTextButtonTheme(false),
      inputDecorationTheme: _buildInputDecorationTheme(false),
      chipTheme: _buildChipTheme(false),
      bottomNavigationBarTheme: _buildBottomNavigationBarTheme(false),
      tabBarTheme: _buildTabBarTheme(false),
      floatingActionButtonTheme: _buildFloatingActionButtonTheme(false),
      dividerTheme: const DividerThemeData(
        color: AppColors.borderDark,
        thickness: 1,
      ),
      scaffoldBackgroundColor: AppColors.backgroundDark,
    );
  }

  static TextTheme _buildTextTheme(Color textColor) {
    return GoogleFonts.robotoTextTheme().copyWith(
      displayLarge: GoogleFonts.roboto(
        fontSize: 57,
        fontWeight: FontWeight.w400,
        color: textColor,
      ),
      displayMedium: GoogleFonts.roboto(
        fontSize: 45,
        fontWeight: FontWeight.w400,
        color: textColor,
      ),
      displaySmall: GoogleFonts.roboto(
        fontSize: 36,
        fontWeight: FontWeight.w400,
        color: textColor,
      ),
      headlineLarge: GoogleFonts.roboto(
        fontSize: 32,
        fontWeight: FontWeight.w600,
        color: textColor,
      ),
      headlineMedium: GoogleFonts.roboto(
        fontSize: 28,
        fontWeight: FontWeight.w600,
        color: textColor,
      ),
      headlineSmall: GoogleFonts.roboto(
        fontSize: 24,
        fontWeight: FontWeight.w600,
        color: textColor,
      ),
      titleLarge: GoogleFonts.roboto(
        fontSize: 22,
        fontWeight: FontWeight.w500,
        color: textColor,
      ),
      titleMedium: GoogleFonts.roboto(
        fontSize: 16,
        fontWeight: FontWeight.w500,
        color: textColor,
      ),
      titleSmall: GoogleFonts.roboto(
        fontSize: 14,
        fontWeight: FontWeight.w500,
        color: textColor,
      ),
      bodyLarge: GoogleFonts.roboto(
        fontSize: 16,
        fontWeight: FontWeight.w400,
        color: textColor,
      ),
      bodyMedium: GoogleFonts.roboto(
        fontSize: 14,
        fontWeight: FontWeight.w400,
        color: textColor,
      ),
      bodySmall: GoogleFonts.roboto(
        fontSize: 12,
        fontWeight: FontWeight.w400,
        color: textColor,
      ),
      labelLarge: GoogleFonts.roboto(
        fontSize: 14,
        fontWeight: FontWeight.w500,
        color: textColor,
      ),
      labelMedium: GoogleFonts.roboto(
        fontSize: 12,
        fontWeight: FontWeight.w500,
        color: textColor,
      ),
      labelSmall: GoogleFonts.roboto(
        fontSize: 11,
        fontWeight: FontWeight.w500,
        color: textColor,
      ),
    );
  }

  static AppBarTheme _buildAppBarTheme(bool isLight) {
    return AppBarTheme(
      elevation: 0,
      backgroundColor: isLight ? AppColors.surface : AppColors.surfaceDark,
      foregroundColor: isLight ? AppColors.textPrimary : AppColors.textPrimaryDark,
      titleTextStyle: GoogleFonts.roboto(
        fontSize: 20,
        fontWeight: FontWeight.w600,
        color: isLight ? AppColors.textPrimary : AppColors.textPrimaryDark,
      ),
      iconTheme: IconThemeData(
        color: isLight ? AppColors.textPrimary : AppColors.textPrimaryDark,
      ),
      actionsIconTheme: IconThemeData(
        color: isLight ? AppColors.textPrimary : AppColors.textPrimaryDark,
      ),
    );
  }

  static CardTheme _buildCardTheme(bool isLight) {
    return CardTheme(
      elevation: 2,
      color: isLight ? AppColors.surface : AppColors.surfaceDark,
      shadowColor: AppColors.shadow,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
      ),
    );
  }

  static ElevatedButtonThemeData _buildElevatedButtonTheme(bool isLight) {
    return ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: AppColors.primary,
        foregroundColor: AppColors.textOnPrimary,
        elevation: 2,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        textStyle: GoogleFonts.roboto(
          fontSize: 16,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }

  static OutlinedButtonThemeData _buildOutlinedButtonTheme(bool isLight) {
    return OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: AppColors.primary,
        side: const BorderSide(color: AppColors.primary),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        textStyle: GoogleFonts.roboto(
          fontSize: 16,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }

  static TextButtonThemeData _buildTextButtonTheme(bool isLight) {
    return TextButtonThemeData(
      style: TextButton.styleFrom(
        foregroundColor: AppColors.primary,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        textStyle: GoogleFonts.roboto(
          fontSize: 16,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }

  static InputDecorationTheme _buildInputDecorationTheme(bool isLight) {
    return InputDecorationTheme(
      filled: true,
      fillColor: isLight ? AppColors.surfaceVariant : AppColors.surfaceVariantDark,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(8),
        borderSide: const BorderSide(color: AppColors.border),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(8),
        borderSide: const BorderSide(color: AppColors.border),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(8),
        borderSide: const BorderSide(color: AppColors.primary, width: 2),
      ),
      errorBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(8),
        borderSide: const BorderSide(color: AppColors.error),
      ),
      focusedErrorBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(8),
        borderSide: const BorderSide(color: AppColors.error, width: 2),
      ),
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      labelStyle: GoogleFonts.roboto(
        fontSize: 16,
        color: isLight ? AppColors.textSecondary : AppColors.textSecondaryDark,
      ),
      hintStyle: GoogleFonts.roboto(
        fontSize: 16,
        color: isLight ? AppColors.textTertiary : AppColors.textTertiaryDark,
      ),
    );
  }

  static ChipThemeData _buildChipTheme(bool isLight) {
    return ChipThemeData(
      backgroundColor: isLight ? AppColors.surfaceVariant : AppColors.surfaceVariantDark,
      labelStyle: GoogleFonts.roboto(
        fontSize: 14,
        color: isLight ? AppColors.textPrimary : AppColors.textPrimaryDark,
      ),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
    );
  }

  static BottomNavigationBarThemeData _buildBottomNavigationBarTheme(bool isLight) {
    return BottomNavigationBarThemeData(
      backgroundColor: isLight ? AppColors.surface : AppColors.surfaceDark,
      selectedItemColor: AppColors.primary,
      unselectedItemColor: isLight ? AppColors.textSecondary : AppColors.textSecondaryDark,
      type: BottomNavigationBarType.fixed,
      elevation: 8,
    );
  }

  static TabBarTheme _buildTabBarTheme(bool isLight) {
    return TabBarTheme(
      labelColor: AppColors.primary,
      unselectedLabelColor: isLight ? AppColors.textSecondary : AppColors.textSecondaryDark,
      indicator: const UnderlineTabIndicator(
        borderSide: BorderSide(color: AppColors.primary, width: 2),
      ),
      labelStyle: GoogleFonts.roboto(
        fontSize: 14,
        fontWeight: FontWeight.w600,
      ),
      unselectedLabelStyle: GoogleFonts.roboto(
        fontSize: 14,
        fontWeight: FontWeight.w400,
      ),
    );
  }

  static FloatingActionButtonThemeData _buildFloatingActionButtonTheme(bool isLight) {
    return const FloatingActionButtonThemeData(
      backgroundColor: AppColors.primary,
      foregroundColor: AppColors.textOnPrimary,
      elevation: 4,
      shape: CircleBorder(),
    );
  }
}