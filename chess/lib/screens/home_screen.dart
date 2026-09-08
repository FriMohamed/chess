import 'package:chess/widgets/custom_button.dart';
import 'package:chess/widgets/player_name.dart';
import 'package:flutter/material.dart';

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
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
          child: Column(
            children: [
              const SizedBox(height: 32),
              Text(
                '♔',
                style: TextStyle(
                  fontSize: 72,
                  color: colors.primary,
                  height: 1.0,
                ),
              ),

              const SizedBox(height: 8),

              Text(
                'CHESS',
                style: theme.textTheme.headlineLarge?.copyWith(
                  fontWeight: FontWeight.w800,
                  letterSpacing: 6,
                ),
              ),

              const SizedBox(height: 8),

              Text('Play. Think. Win.', style: theme.textTheme.bodyMedium),

              const SizedBox(height: 48),

              const PlayerName(),

              const SizedBox(height: 140),

              Column(
                children: [
                  CustomButton(
                    label: 'QUICK GAME',
                    backgroundColor: colors.primary,
                    textColor: colors.onPrimary,
                    fontWeight: FontWeight.w800,
                    onPressed: () => _handleQuickGame(context),
                  ),

                  const SizedBox(height: 16),

                  CustomButton(
                    label: 'PRIVATE',
                    backgroundColor: colors.surface,
                    textColor: colors.onSurface,
                    borderColor: colors.outline,
                    fontWeight: FontWeight.w700,
                    onPressed: _handlePlayFriend,
                  ),
                ],
              ),

              const SizedBox(height: 48),
            ],
          ),
        ),
      ),
    );
  }
}
