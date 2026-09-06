import 'package:chess/widgets/custom_button.dart';
import 'package:chess/widgets/player_name.dart';
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  void _handleQuickGame(BuildContext context) {
    Navigator.of(context).pushNamed('/quick-game');
  }

  void _handlePlayFriend() {
    debugPrint('Play friend pressed');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.lg,
            vertical: AppSpacing.xl,
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              // Header Section
              Padding(
                padding: const EdgeInsets.only(top: AppSpacing.xl),
                child: Column(
                  children: const [
                    Text(
                      '♟',
                      style: TextStyle(
                        fontSize: 72,
                        color: AppColors.primary,
                        height: 1.0,
                      ),
                    ),
                    SizedBox(height: AppSpacing.sm),
                    Text(
                      'CHESS',
                      style: TextStyle(
                        color: AppColors.text,
                        fontSize: 36,
                        fontWeight: FontWeight.w800,
                        letterSpacing: 6,
                      ),
                    ),
                    SizedBox(height: AppSpacing.sm),
                    Text(
                      'Play. Think. Win.',
                      style: TextStyle(
                        color: AppColors.textSecondary,
                        fontSize: 16,
                      ),
                    ),
                  ],
                ),
              ),

              const PlayerName(),

              // Actions Section
              Padding(
                padding: const EdgeInsets.only(bottom: 1.5 * AppSpacing.xxl),
                child: Column(
                  children: [
                    // Primary Button
                    CustomButton(
                      label: 'QUICK GAME',
                      backgroundColor: AppColors.primary,
                      textColor: AppColors.background,
                      fontWeight: FontWeight.w800,
                      onPressed: () => _handleQuickGame(context),
                    ),
                    const SizedBox(height: AppSpacing.md),
                    // Secondary Button
                    CustomButton(
                      label: 'PLAY A FRIEND',
                      backgroundColor: AppColors.surface,
                      textColor: AppColors.text,
                      borderColor: AppColors.border,
                      fontWeight: FontWeight.w700,
                      onPressed: _handlePlayFriend,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}