import 'package:chess_app/screens/game_waiting_screen.dart';
import 'package:chess_app/widgets/custom_button.dart';
import 'package:chess_app/widgets/player_name.dart';
import 'package:flutter/material.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  void _handleQuickGame(BuildContext context) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => const GameWaitingScreen(mode: GameMode.public),
      ),
    );
  }

  void _handlePlayFriend(BuildContext context) {
    // No API call here.
    //
    // Home only lets the player choose between creating a private game
    // or joining an existing one.
    _showPrivateOptions(context);
  }

  void _showPrivateOptions(BuildContext homeContext) {
    final colors = Theme.of(homeContext).colorScheme;

    showModalBottomSheet(
      context: homeContext,
      backgroundColor: colors.surface,
      showDragHandle: true,
      builder: (sheetContext) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(24, 8, 24, 24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // ...
                CustomButton(
                  label: 'CREATE GAME',
                  backgroundColor: colors.primary,
                  textColor: colors.onPrimary,
                  fontWeight: FontWeight.w800,
                  onPressed: () {
                    Navigator.of(sheetContext).pop();

                    Navigator.of(homeContext).push(
                      MaterialPageRoute(
                        builder: (_) => const GameWaitingScreen(
                          mode: GameMode.privateCreate,
                        ),
                      ),
                    );
                  },
                ),

                const SizedBox(height: 12),

                CustomButton(
                  label: 'JOIN GAME',
                  backgroundColor: colors.surface,
                  textColor: colors.onSurface,
                  borderColor: colors.outline,
                  fontWeight: FontWeight.w700,
                  onPressed: () {
                    Navigator.of(sheetContext).pop();

                    _showJoinGameDialog(homeContext);
                  },
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  void _showJoinGameDialog(BuildContext homeContext) {
    final controller = TextEditingController();

    showDialog(
      context: homeContext,
      builder: (dialogContext) {
        return AlertDialog(
          title: const Text('JOIN GAME'),
          content: TextField(
            controller: controller,
            autofocus: true,
            textCapitalization: TextCapitalization.characters,
            decoration: const InputDecoration(
              labelText: 'Game code',
              hintText: 'Enter code',
            ),
          ),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.of(dialogContext).pop();
              },
              child: const Text('CANCEL'),
            ),

            FilledButton(
              onPressed: () {
                final code = controller.text.trim();

                if (code.isEmpty) return;

                Navigator.of(dialogContext).pop();

                Navigator.of(homeContext).push(
                  MaterialPageRoute(
                    builder: (_) => GameWaitingScreen(
                      mode: GameMode.privateJoin,
                      code: code,
                    ),
                  ),
                );
              },
              child: const Text('JOIN'),
            ),
          ],
        );
      },
    );
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
                    onPressed: () => _handlePlayFriend(context),
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
