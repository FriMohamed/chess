import 'package:chess_app/models/game_state.dart';
import 'package:flutter/material.dart';

class GameFinishedOverlay extends StatelessWidget {
  final GameState game;
  final String playerId;
  final VoidCallback onHomePressed;

  const GameFinishedOverlay({
    super.key,
    required this.game,
    required this.playerId,
    required this.onHomePressed,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Positioned.fill(
      child: Container(
        color: colors.scrim.withValues(alpha: 0.65),
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
              Icon(_icon, size: 48, color: colors.primary),
              const SizedBox(height: 20),
              Text(
                _title,
                textAlign: TextAlign.center,
                style: theme.textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 10),
              Text(
                _message,
                textAlign: TextAlign.center,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: colors.onSurface.withValues(alpha: 0.65),
                ),
              ),
              const SizedBox(height: 28),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: onHomePressed,
                  child: const Text('Back to Home'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String get _title {
    switch (game.endReason) {
      case EndReason.checkmate:
      case EndReason.resignation:
      case EndReason.timeout:
      case EndReason.disconnect:
      case EndReason.quit:
        return _didIWin ? 'You Win' : 'You Lose';

      case EndReason.stalemate:
      case EndReason.threefoldRepetition:
      case EndReason.fivefoldRepetition:
      case EndReason.fiftyMoveRule:
      case EndReason.seventyFiveMoveRule:
      case EndReason.insufficientMaterial:
      case EndReason.drawAgreement:
        return 'Draw';

      case EndReason.none:
        return 'Game Over';
    }
  }

  bool get _didIWin {
    if (game.result == GameResult.whiteWins) {
      return game.white?.id == playerId;
    }

    if (game.result == GameResult.blackWins) {
      return game.black?.id == playerId;
    }

    return false;
  }

  String get _message {
    switch (game.endReason) {
      case EndReason.checkmate:
        return 'Checkmate';

      case EndReason.stalemate:
        return 'Stalemate';

      case EndReason.threefoldRepetition:
        return 'Draw by threefold repetition';

      case EndReason.fivefoldRepetition:
        return 'Draw by fivefold repetition';

      case EndReason.fiftyMoveRule:
        return 'Draw by the fifty-move rule';

      case EndReason.seventyFiveMoveRule:
        return 'Draw by the seventy-five-move rule';

      case EndReason.insufficientMaterial:
        return 'Draw by insufficient material';

      case EndReason.drawAgreement:
        return 'Draw by agreement';

      case EndReason.resignation:
        return _didIWin ? 'Your opponent resigned' : 'You resigned';

      case EndReason.timeout:
        return _didIWin
            ? 'Your opponent ran out of time'
            : 'You ran out of time';

      case EndReason.quit:
        return 'Your opponent left the game';

      case EndReason.disconnect:
        return 'Your opponent disconnected';

      case EndReason.none:
        return '';
    }
  }

  IconData get _icon {
    if (game.result == GameResult.draw) {
      return Icons.handshake_outlined;
    }

    switch (game.endReason) {
      case EndReason.resignation:
        return Icons.flag_outlined;

      case EndReason.timeout:
        return Icons.timer_off_outlined;

      case EndReason.disconnect:
        return Icons.wifi_off_outlined;

      case EndReason.quit:
        return Icons.exit_to_app_outlined;

      case EndReason.checkmate:
        return Icons.emoji_events_outlined;

      case EndReason.stalemate:
      case EndReason.threefoldRepetition:
      case EndReason.fivefoldRepetition:
      case EndReason.fiftyMoveRule:
      case EndReason.seventyFiveMoveRule:
      case EndReason.insufficientMaterial:
      case EndReason.drawAgreement:
        return Icons.handshake_outlined;

      case EndReason.none:
        return Icons.flag_outlined;
    }
  }
}
