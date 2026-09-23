import 'dart:async';

import 'package:chess/cubits/game_cubit.dart';
import 'package:chess/cubits/game_screen_state.dart';
import 'package:chess/models/game_session.dart';
import 'package:chess/models/game_state.dart';
import 'package:chess/services/api/game_api_service.dart';
import 'package:chess/services/game_socket_service.dart';
import 'package:chess/widgets/confirm_modal.dart';
import 'package:chess/widgets/game_screen/chess_board.dart';
import 'package:chess/widgets/game_screen/game_finished_overlay.dart';
import 'package:chess/widgets/game_screen/player_bar.dart';
import 'package:chess/widgets/game_screen/promotion_picker.dart';
import 'package:chess/widgets/loading_dots.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class GameScreen extends StatefulWidget {
  final GameState game;
  final GameSession session;
  final GameSocketService socket;

  const GameScreen({
    super.key,
    required this.game,
    required this.session,
    required this.socket,
  });

  @override
  State<GameScreen> createState() => _GameScreenState();
}

class _GameScreenState extends State<GameScreen> {
  late final GameCubit _gameCubit;

  StreamSubscription<GameSocketStatus>? _socketStatusSubscription;

  GameSocketStatus _socketStatus = GameSocketStatus.disconnected;

  bool _isLeaving = false;

  @override
  void initState() {
    super.initState();

    _gameCubit = GameCubit(
      game: widget.game,
      playerId: widget.session.playerId,
      socket: widget.socket,
    );

    _socketStatus = widget.socket.status;

    _socketStatusSubscription = widget.socket.statusStream.listen(
      _handleSocketStatus,
    );
  }

  void _handleSocketStatus(GameSocketStatus status) {
    if (!mounted) return;

    setState(() {
      _socketStatus = status;
    });
  }

  @override
  void dispose() {
    _socketStatusSubscription?.cancel();

    _gameCubit.close();
    widget.socket.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final playerId = widget.session.playerId;

    return BlocProvider.value(
      value: _gameCubit,
      child: PopScope(
        canPop: false,
        onPopInvokedWithResult: (didPop, result) {
          if (didPop) return;

          final state = _gameCubit.state;
          if (state.game.status == GameStatus.finished) {
            _goHome(context);
            return;
          }

          if (_isLeaving) return;
          _showQuitDialog(context);
        },
        child: Scaffold(
          body: SafeArea(
            child: BlocBuilder<GameCubit, GameScreenState>(
              builder: (context, state) {
                final isWhite = playerId == state.game.white?.id;
                final myPlayer = isWhite ? state.game.white : state.game.black;
                final opponent = isWhite ? state.game.black : state.game.white;
                final myTime = isWhite ? state.whiteTime : state.blackTime;
                final opponentTime = isWhite
                    ? state.blackTime
                    : state.whiteTime;
                final myTurn = isWhite
                    ? state.game.active == ActiveColor.white
                    : state.game.active == ActiveColor.black;

                return Stack(
                  children: [
                    // Center the entire UI vertically and constrain its maximum width
                    Center(
                      child: ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 600),
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            PlayerBar(
                              name: opponent?.nickname ?? 'Opponent',
                              time: opponentTime,
                              isActive: !myTurn,
                              isCurrentPlayer: false,
                              // Pass draw state to opponent bar so they can see draw offers!
                              drawState: state.drawOfferState,
                              onDrawPressed: () {
                                if (state.drawOfferState ==
                                    DrawOfferState.received) {
                                  _showDrawOfferDialog(context);
                                }
                              },
                            ),

                            Stack(
                              alignment: Alignment.center,
                              children: [
                                ChessBoard(
                                  game: state.game,
                                  isWhite: isWhite,
                                  selectedSquare: state.selectedSquare,
                                  legalMoves: state.legalMoves,
                                  onSquareTap: (square) {
                                    context.read<GameCubit>().selectSquare(
                                      square,
                                    );
                                  },
                                ),

                                if (state.isPromotionPending)
                                  PromotionPicker(
                                    isWhite: isWhite,
                                    onSelected: (promotion) {
                                      context.read<GameCubit>().promote(
                                        promotion,
                                      );
                                    },
                                  ),

                                if (state.game.status == GameStatus.finished)
                                  GameFinishedOverlay(
                                    game: state.game,
                                    playerId: playerId,
                                    onHomePressed: () {
                                      _goHome(context);
                                    },
                                  ),
                              ],
                            ),

                            PlayerBar(
                              name: myPlayer?.nickname ?? 'You',
                              time: myTime,
                              isActive: myTurn,
                              isCurrentPlayer: true,
                              drawState: state.drawOfferState,
                              onResignPressed: () {
                                _showResignDialog(context);
                              },
                              onDrawPressed: () {
                                if (state.drawOfferState ==
                                    DrawOfferState.none) {
                                  context.read<GameCubit>().offerDraw();
                                  return;
                                }

                                if (state.drawOfferState ==
                                    DrawOfferState.received) {
                                  _showDrawOfferDialog(context);
                                }
                              },
                            ),
                          ],
                        ),
                      ),
                    ),

                    if (_socketStatus == GameSocketStatus.reconnecting)
                      _buildReconnectingOverlay(context),

                    if (_socketStatus == GameSocketStatus.disconnected)
                      _buildDisconnectedOverlay(context),
                  ],
                );
              },
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildReconnectingOverlay(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Positioned.fill(
      child: Container(
        color: colors.scrim.withValues(alpha: 0.55),
        alignment: Alignment.center,
        child: Container(
          padding: const EdgeInsets.all(28),
          margin: const EdgeInsets.symmetric(horizontal: 32),
          decoration: BoxDecoration(
            color: colors.surface,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: colors.outline.withValues(alpha: 0.25)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const LoadingDots(),

              const SizedBox(height: 20),

              Text(
                'CONNECTION LOST',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  letterSpacing: 1.5,
                ),
              ),

              const SizedBox(height: 8),

              Text(
                'Trying to reconnect...',
                textAlign: TextAlign.center,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: colors.onSurface.withValues(alpha: 0.65),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildDisconnectedOverlay(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Positioned.fill(
      child: Container(
        color: colors.scrim.withValues(alpha: 0.65),
        alignment: Alignment.center,
        child: Container(
          padding: const EdgeInsets.all(28),
          margin: const EdgeInsets.symmetric(horizontal: 32),
          decoration: BoxDecoration(
            color: colors.surface,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: colors.outline.withValues(alpha: 0.25)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.wifi_off_rounded,
                size: 42,
                color: colors.onSurface.withValues(alpha: 0.7),
              ),

              const SizedBox(height: 18),

              Text(
                'CONNECTION LOST',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  letterSpacing: 1.5,
                ),
              ),

              const SizedBox(height: 8),

              Text(
                'We couldn’t reconnect to the game.',
                textAlign: TextAlign.center,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: colors.onSurface.withValues(alpha: 0.65),
                ),
              ),

              const SizedBox(height: 24),

              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  onPressed: _retryConnection,
                  child: const Text('TRY AGAIN'),
                ),
              ),

              const SizedBox(height: 8),

              SizedBox(
                width: double.infinity,
                child: TextButton(
                  onPressed: () {
                    _goHome(context);
                  },
                  child: const Text('GO HOME'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _retryConnection() {
    if (_isLeaving) {
      return;
    }

    widget.socket.retry();
  }

  void _showDrawOfferDialog(BuildContext context) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) {
        return ConfirmModal(
          title: 'Draw Offer',
          message: 'Your opponent has offered a draw.',
          confirmText: 'Accept',
          cancelText: 'Decline',
          onConfirm: () {
            Navigator.of(dialogContext).pop();

            _gameCubit.respondToDraw(accepted: true);
          },
          onCancel: () {
            Navigator.of(dialogContext).pop();

            _gameCubit.respondToDraw(accepted: false);
          },
        );
      },
    );
  }

  void _showResignDialog(BuildContext context) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) {
        return ConfirmModal(
          title: 'Resign Game',
          message: 'Are you sure you want to resign?',
          confirmText: 'Resign',
          cancelText: 'Cancel',
          onConfirm: () {
            Navigator.of(dialogContext).pop();

            _gameCubit.resign();
          },
          onCancel: () {
            Navigator.of(dialogContext).pop();
          },
        );
      },
    );
  }

  void _showQuitDialog(BuildContext context) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) {
        return ConfirmModal(
          title: 'Leave Game',
          message: 'Are you sure you want to leave the game?',
          confirmText: 'Leave',
          cancelText: 'Stay',
          onConfirm: () async {
            Navigator.of(dialogContext).pop();

            await _leaveGame();
          },
          onCancel: () {
            Navigator.of(dialogContext).pop();
          },
        );
      },
    );
  }

  Future<void> _leaveGame() async {
    if (_isLeaving) {
      return;
    }

    setState(() {
      _isLeaving = true;
    });

    try {
      await GameApiService.quitGame(sessionId: widget.session.sessionId);
    } catch (_) {
      // Leaving is best-effort.
    }

    if (!mounted) {
      return;
    }

    widget.socket.close();

    _goHome(context);
  }

  void _goHome(BuildContext context) {
    if (!mounted) {
      return;
    }

    Navigator.of(context).pushNamedAndRemoveUntil('/', (route) => false);
  }
}
