import 'dart:async';

import 'package:chess/models/game_state.dart';
import 'package:chess/services/api/game_api_service.dart';
import 'package:chess/services/game_socket_service.dart';
import 'package:chess/services/player_name_service.dart';
import 'package:chess/widgets/confirm_modal.dart';
import 'package:chess/widgets/error_overlay.dart';
import 'package:chess/widgets/loading_dots.dart';
import 'package:flutter/material.dart';

enum QuickGameState { searching, opponentFound, error }

class QuickGameScreen extends StatefulWidget {
  const QuickGameScreen({super.key});

  @override
  State<QuickGameScreen> createState() => _QuickGameScreenState();
}

class _QuickGameScreenState extends State<QuickGameScreen> {
  QuickGameState _state = QuickGameState.searching;

  GameState? _game;
  String _errorMessage = '';

  String? _gameId;
  String? _playerId;

  bool _showCancelModal = false;
  bool _allowNavigation = false;

  final GameSocketService _socketService = GameSocketService();

  StreamSubscription<GameSocketMessage>? _socketSubscription;
  StreamSubscription<Object>? _socketErrorSubscription;

  @override
  void initState() {
    super.initState();

    _listenToSocket();
    _startMatchmaking();
  }

  void _listenToSocket() {
    _socketSubscription = _socketService.messages.listen(_handleSocketMessage);

    _socketErrorSubscription = _socketService.errors.listen(_handleSocketError);
  }

  void _handleSocketMessage(GameSocketMessage message) {
    if (!mounted) return;

    if (message is GameStartedMessage) {
      setState(() {
        _game = message.game;
        _state = QuickGameState.opponentFound;
      });

      // Later:
      // Navigator.of(context).pushReplacementNamed(
      //   '/game',
      //   arguments: ...
      // );
    }

    if (message is InvalidMessage) {
      _setError('The server received an invalid message.');
    }

    if (message is UnknownMessage) {
      _setError('The server sent an unknown message.');
    }
  }

  void _handleSocketError(Object error) {
    if (!mounted) return;

    _setError('Couldn’t connect to the game server.');
  }

  Future<void> _startMatchmaking() async {
    if (!mounted) return;

    setState(() {
      _state = QuickGameState.searching;
      _errorMessage = '';
      _game = null;
    });

    try {
      final nickname = await PlayerService.getPlayerName();

      if (!mounted) return;

      final response = await GameApiService.quickGame(nickname);

      _playerId = response.playerId;
      _gameId = response.gameId;

      if (!mounted) return;

      _socketService.connect(
        gameId: response.gameId,
        playerId: response.playerId,
      );
    } catch (error) {
      if (!mounted) return;

      _setError('Couldn’t connect to the game server.');
    }
  }

  void _setError(String message) {
    if (!mounted) return;

    setState(() {
      _state = QuickGameState.error;
      _errorMessage = message;
    });
  }

  Future<void> _handleRetry() async {
    _socketService.close();

    await _startMatchmaking();
  }

  void _handleBack() {
    if (_state == QuickGameState.error || _allowNavigation) {
      Navigator.of(context).pop();
      return;
    }

    setState(() {
      _showCancelModal = true;
    });
  }

  void _keepSearching() {
    setState(() {
      _showCancelModal = false;
    });
  }

  Future<void> _confirmCancel() async {
    if (_gameId == null || _playerId == null) {
      if (mounted) Navigator.of(context).pop();
      return;
    }

    await GameApiService.quitGame(gameId: _gameId!, playerId: _playerId!);
  }

  @override
  void dispose() {
    _socketSubscription?.cancel();
    _socketErrorSubscription?.cancel();

    _socketService.dispose();

    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

    return PopScope(
      canPop: _state == QuickGameState.error || _allowNavigation,
      onPopInvokedWithResult: (didPop, result) {
        if (didPop) return;

        _handleBack();
      },
      child: Scaffold(
        body: SafeArea(
          child: Stack(
            children: [
              _buildContent(),

              if (_state == QuickGameState.error)
                ErrorOverlay(
                  title: 'CONNECTION ERROR',
                  message: _errorMessage.isEmpty
                      ? 'Couldn’t connect to the game server.'
                      : _errorMessage,
                  primaryButtonText: 'TRY AGAIN',
                  secondaryButtonText: 'GO BACK',
                  onPrimaryPressed: _handleRetry,
                  onSecondaryPressed: _goBack,
                ),

              if (_showCancelModal)
                Positioned.fill(
                  child: Container(
                    color: colors.scrim.withValues(alpha: 0.5),
                    alignment: Alignment.center,
                    child: ConfirmModal(
                      title: 'CANCEL MATCHMAKING?',
                      message: 'Are you sure you want to quit?',
                      confirmText: 'CANCEL GAME',
                      cancelText: 'KEEP SEARCHING',
                      onConfirm: _confirmCancel,
                      onCancel: _keepSearching,
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildContent() {
    switch (_state) {
      case QuickGameState.searching:
        return _buildSearching();

      case QuickGameState.opponentFound:
        return _buildOpponentFound();

      case QuickGameState.error:
        return _buildSearching();
    }
  }

  Widget _buildSearching() {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(
            '♟',
            style: TextStyle(fontSize: 64, color: colors.primary, height: 1),
          ),

          const SizedBox(height: 24),

          Text(
            'FINDING OPPONENT',
            style: theme.textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.bold,
              letterSpacing: 2,
            ),
          ),

          const SizedBox(height: 12),

          Text(
            'Waiting for a player to join...',
            style: theme.textTheme.bodyMedium?.copyWith(
              color: colors.onSurface.withValues(alpha: 0.65),
            ),
          ),

          const SizedBox(height: 24),

          const LoadingDots(),
        ],
      ),
    );
  }

  Widget _buildOpponentFound() {
    final theme = Theme.of(context);

    final white = _game?.white;
    final black = _game?.black;

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(
            'OPPONENT FOUND',
            style: theme.textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.bold,
              letterSpacing: 2,
            ),
          ),

          const SizedBox(height: 40),

          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              _buildPlayer(nickname: white?.nickname ?? 'White', isWhite: true),

              const SizedBox(width: 28),

              Text(
                'VS',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),

              const SizedBox(width: 28),

              _buildPlayer(
                nickname: black?.nickname ?? 'Black',
                isWhite: false,
              ),
            ],
          ),

          const SizedBox(height: 40),

          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text('Starting', style: theme.textTheme.bodyLarge),

              const SizedBox(width: 12),

              const LoadingDots(),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildPlayer({required String nickname, required bool isWhite}) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Column(
      children: [
        Container(
          width: 64,
          height: 64,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: colors.surface,
            border: Border.all(width: 1),
          ),
          alignment: Alignment.center,
          child: Text(
            '♔',
            style: TextStyle(fontSize: 36, color: colors.primary, height: 1),
          ),
        ),

        const SizedBox(height: 10),

        SizedBox(
          width: 100,
          child: Text(
            nickname,
            textAlign: TextAlign.center,
            overflow: TextOverflow.ellipsis,
            style: theme.textTheme.bodyMedium?.copyWith(
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
      ],
    );
  }

  void _goBack() {
    setState(() {
      _allowNavigation = true;
    });

    Navigator.of(context).pop();
  }
}
