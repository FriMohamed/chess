import 'package:flutter/material.dart';

class ErrorOverlay extends StatelessWidget {
  final String title;
  final String message;

  final String primaryButtonText;
  final String secondaryButtonText;

  final VoidCallback onPrimaryPressed;
  final VoidCallback onSecondaryPressed;

  final Widget? icon;

  const ErrorOverlay({
    super.key,
    required this.title,
    required this.message,
    required this.primaryButtonText,
    required this.secondaryButtonText,
    required this.onPrimaryPressed,
    required this.onSecondaryPressed,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Positioned.fill(
      child: Container(
        color: colors.scrim.withValues(alpha: 0.6),
        alignment: Alignment.center,
        child: Container(
          margin: const EdgeInsets.symmetric(horizontal: 32),
          padding: const EdgeInsets.all(28),
          decoration: BoxDecoration(
            color: colors.surface,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: colors.outline),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              icon ??
                  Icon(
                    Icons.warning_amber_rounded,
                    size: 42,
                    color: colors.primary,
                  ),

              const SizedBox(height: 20),

              Text(
                title,
                textAlign: TextAlign.center,
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                  letterSpacing: 1,
                ),
              ),

              const SizedBox(height: 12),

              Text(
                message,
                textAlign: TextAlign.center,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: colors.onSurface.withValues(alpha: 0.65),
                ),
              ),

              const SizedBox(height: 28),

              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: onPrimaryPressed,
                  child: Text(primaryButtonText),
                ),
              ),

              const SizedBox(height: 8),

              SizedBox(
                width: double.infinity,
                child: TextButton(
                  onPressed: onSecondaryPressed,
                  child: Text(
                    secondaryButtonText,
                    style: TextStyle(
                      color: colors.onSurface.withValues(alpha: 0.7),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
