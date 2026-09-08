import 'package:flutter/material.dart';

class _AppColors {
  static const Color background = Color(0xFF111315);

  static const Color surface = Color(0xFF191C1F);

  // static const Color surfaceElevated = Color(0xFF202428);

  static const Color primary = Color(0xFFC9A96E);

  // static const Color primaryPressed = Color(0xFFA98B5B);

  static const Color text = Color(0xFFE8E5DE);

  static const Color textSecondary = Color(0xFF9B9A95);

  static const Color error = Color(0xFFD96B6B);

  static const Color border = Color(0xFF2B2F32);
}

class AppTheme {
  static ThemeData get dark {
    return ThemeData(
      brightness: Brightness.dark,

      scaffoldBackgroundColor: _AppColors.background,

      colorScheme: const ColorScheme.dark(
        primary: _AppColors.primary,
        surface: _AppColors.surface,
        error: _AppColors.error,
        onPrimary: _AppColors.background,
        onSurface: _AppColors.text,
      ),

      textTheme: const TextTheme(
        bodyLarge: TextStyle(
          color: _AppColors.text,
        ),
        bodyMedium: TextStyle(
          color: _AppColors.text,
        ),
        bodySmall: TextStyle(
          color: _AppColors.textSecondary,
        ),
      ),

      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: _AppColors.primary,
          foregroundColor: _AppColors.background,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.all(
              Radius.circular(14),
            ),
          ),
        ),
      ),

      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: _AppColors.surface,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.all(
            Radius.circular(12),
          ),
          borderSide: BorderSide(
            color: _AppColors.border,
          ),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.all(
            Radius.circular(12),
          ),
          borderSide: BorderSide(
            color: _AppColors.border,
          ),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.all(
            Radius.circular(12),
          ),
          borderSide: BorderSide(
            color: _AppColors.primary,
          ),
        ),
      ),
    );
  }
}