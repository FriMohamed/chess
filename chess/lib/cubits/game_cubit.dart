import 'dart:async';

import 'package:chess/chess/chess_rules.dart';
import 'package:chess/cubits/game_screen_state.dart';
import 'package:chess/models/game_state.dart';
import 'package:chess/services/game_socket_service.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class GameCubit extends Cubit<GameScreenState> {
  final ChessRules _chessRules = ChessRules();
  final GameSocketService _socket;
  final String _playerId;
  Timer? _clockTimer;

  late final StreamSubscription<GameSocketMessage> _messageSubscription;

  GameCubit({
    required GameState game,
    required String playerId,
    required GameSocketService socket,
  }) : _socket = socket,
       _playerId = playerId,
       super(GameScreenState(game: game)) {
    _messageSubscription = _socket.messages.listen(_handleMessage);
    _startClock();
  }

  void _startClock() {
    _clockTimer = Timer.periodic(
      const Duration(seconds: 1),
      (_) => _tickClock(),
    );
  }

  void _tickClock() {
    if (state.game.status != GameStatus.playing) {
      return;
    }

    if (state.game.active == ActiveColor.white) {
      if (state.whiteTime <= 0) return;

      emit(
        GameScreenState(
          game: state.game,
          whiteTime: state.whiteTime - 1000,
          blackTime: state.blackTime,
          selectedSquare: state.selectedSquare,
          legalMoves: state.legalMoves,
          promotionFrom: state.promotionFrom,
          promotionTo: state.promotionTo,
        ),
      );
    } else {
      if (state.blackTime <= 0) return;

      emit(
        GameScreenState(
          game: state.game,
          whiteTime: state.whiteTime,
          blackTime: state.blackTime - 1000,
          selectedSquare: state.selectedSquare,
          legalMoves: state.legalMoves,
          promotionFrom: state.promotionFrom,
          promotionTo: state.promotionTo,
        ),
      );
    }
  }

  void selectSquare(String square) {
    final selected = state.selectedSquare;

    if (selected == null) {
      _selectPiece(square);
      return;
    }

    if (selected == square) {
      _clearSelection();
      return;
    }

    if (state.legalMoves.contains(square)) {
      _makeMove(selected, square);
      return;
    }

    _selectPiece(square);
  }

  void _selectPiece(String square) {
    final legalMoves = _chessRules.legalMoves(state.game.fen, square);

    emit(
      GameScreenState(
        game: state.game,
        selectedSquare: square,
        legalMoves: legalMoves,
      ),
    );
  }

  void _clearSelection() {
    emit(GameScreenState(game: state.game));
  }

  void _makeMove(String from, String to) {
    if (!_isMyTurn) return;

    if (_chessRules.isPromotionMove(state.game.fen, from)) {
      emit(
        GameScreenState(game: state.game, promotionFrom: from, promotionTo: to),
      );

      return;
    }
    print("from: "+ from + " to: " +to );
    _socket.sendMove(from: from, to: to);

    _clearSelection();
  }

  void promote(String promotion) {
    final from = state.promotionFrom;
    final to = state.promotionTo;

    if (from == null || to == null) {
      return;
    }

    print("from: "+ from + " to: " +to + " promote: " + promotion);
    _socket.sendMove(from: from, to: to, promotion: promotion);

    _clearSelection();
  }

  void _handleMessage(GameSocketMessage message) {
    switch (message) {
      case GameStartedMessage():
        updateGame(message.game);

      case GameStateMessage():
        updateGame(message.game);

      case ErrorMessage():
        // We will handle UI errors later.
        break;

      case UnknownMessage():
        break;
    }
  }

  void updateGame(GameState game) {
    emit(GameScreenState(game: game));
  }

  @override
  Future<void> close() async {
    await _messageSubscription.cancel();
    await _socket.dispose();

    return super.close();
  }

  bool get _isMyTurn {
    final game = state.game;

    if (game.white?.id == _playerId) {
      return game.active == ActiveColor.white;
    }

    if (game.black?.id == _playerId) {
      return game.active == ActiveColor.black;
    }

    return false;
  }
}
