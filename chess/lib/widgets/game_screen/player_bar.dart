import 'package:chess/cubits/game_screen_state.dart';
import 'package:flutter/material.dart';

class PlayerBar extends StatelessWidget {
  final String name;
  final int time;
  final bool isActive;

  final bool isCurrentPlayer;
  final DrawOfferState drawState;

  final VoidCallback? onDrawPressed;
  final VoidCallback? onResignPressed;

  const PlayerBar({
    super.key,
    required this.name,
    required this.time,
    required this.isActive,
    required this.isCurrentPlayer,
    this.drawState = DrawOfferState.none,
    this.onDrawPressed,
    this.onResignPressed,
  });

  static const Color _gold = Color(0xFFD6A84F);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 12),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          border: Border.all(
            color: isActive ? _gold : Colors.transparent,
            width: 1.5,
          ),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    name,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  const SizedBox(height: 20),
                ],
              ),
            ),

            if (isCurrentPlayer) ...[
              _DrawAction(
                state: drawState,
                enabled: isActive,
                onPressed: onDrawPressed,
              ),

              const SizedBox(width: 4),

              IconButton(
                onPressed: isActive ? onResignPressed : null,
                tooltip: 'Resign',
                icon: const Icon(Icons.flag_outlined),
              ),

              const SizedBox(width: 8),
            ],

            Container(
              padding: const EdgeInsets.symmetric(
                horizontal: 12,
                vertical: 6,
              ),
              decoration: BoxDecoration(
                color: isActive ? _gold : colors.surfaceContainer,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                _formatTime(time),
                style: theme.textTheme.titleLarge?.copyWith(
                  color: isActive ? Colors.black : null,
                  fontWeight: FontWeight.bold,
                  fontFeatures: const [
                    FontFeature.tabularFigures(),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatTime(int milliseconds) {
    final totalSeconds = milliseconds ~/ 1000;
    final minutes = totalSeconds ~/ 60;
    final seconds = totalSeconds % 60;

    return '$minutes:${seconds.toString().padLeft(2, '0')}';
  }
}

class _DrawAction extends StatelessWidget {
  final DrawOfferState state;
  final bool enabled;
  final VoidCallback? onPressed;

  const _DrawAction({
    required this.state,
    required this.enabled,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    IconData icon;
    String tooltip;

    switch (state) {
      case DrawOfferState.none:
        icon = Icons.handshake_outlined;
        tooltip = 'Offer draw';

      case DrawOfferState.sent:
        icon = Icons.hourglass_top_rounded;
        tooltip = 'Draw offer pending';

      case DrawOfferState.received:
        icon = Icons.handshake;
        tooltip = 'Draw offer received (Tap to respond)';

      case DrawOfferState.declined:
        icon = Icons.handshake_outlined;
        tooltip = 'Offer draw';
    }

    final canPress =
        state == DrawOfferState.received ||
        (enabled && state == DrawOfferState.none);

    return Stack(
      clipBehavior: Clip.none,
      children: [
        IconButton(
          onPressed: canPress ? onPressed : null,
          tooltip: tooltip,
          icon: Icon(icon),
        ),

        if (state == DrawOfferState.received)
          Positioned(
            top: 7,
            right: 7,
            child: Container(
              width: 9,
              height: 9,
              decoration: BoxDecoration(
                color: theme.colorScheme.error,
                shape: BoxShape.circle,
                border: Border.all(
                  color: theme.scaffoldBackgroundColor,
                  width: 2,
                ),
              ),
            ),
          ),
      ],
    );
  }
}