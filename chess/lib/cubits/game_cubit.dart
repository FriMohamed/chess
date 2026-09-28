import 'dart:async';

import 'package:chess_app/chess/chess_rules.dart';
import 'package:chess_app/cubits/game_screen_state.dart';
import 'package:chess_app/models/game_state.dart';
import 'package:chess_app/services/game_socket_service.dart';
import 'package:chess_app/services/game_sound_service.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

class GameCubit extends Cubit<GameScreenState> {
  final GameSoundService _soundService = GameSoundService();

  final ChessRules _chessRules = ChessRules();
  final GameSocketService _socket;
  final String _playerId;
  bool _whiteLowTimeSoundPlayed = false;
  bool _blackLowTimeSoundPlayed = false;
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

    // Evaluate the initial draw state from the provided GameState object.
    // This ensures correct UI state if the player reconnects or opens the screen mid-game.
    _updateGame(game);
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
      final newTime = state.whiteTime - 1000;

      if (newTime <= 10_000 &&
          !_whiteLowTimeSoundPlayed &&
          _playerId == state.game.white?.id) {
        _whiteLowTimeSoundPlayed = true;
        _soundService.playLowTime();
      }

      emit(
        GameScreenState(
          game: state.game,
          whiteTime: newTime,
          blackTime: state.blackTime,
          selectedSquare: state.selectedSquare,
          legalMoves: state.legalMoves,
          promotionFrom: state.promotionFrom,
          promotionTo: state.promotionTo,
          drawOfferState: state.drawOfferState,
        ),
      );
    } else {
      if (state.blackTime <= 0) return;
      final newTime = state.blackTime - 1000;

      if (newTime <= 10_000 &&
          !_blackLowTimeSoundPlayed &&
          _playerId == state.game.black?.id) {
        _blackLowTimeSoundPlayed = true;
        _soundService.playLowTime();
      }

      emit(
        GameScreenState(
          game: state.game,
          whiteTime: state.whiteTime,
          blackTime: newTime,
          selectedSquare: state.selectedSquare,
          legalMoves: state.legalMoves,
          promotionFrom: state.promotionFrom,
          promotionTo: state.promotionTo,
          drawOfferState: state.drawOfferState,
        ),
      );
    }
  }

  void selectSquare(String square) {
    if (state.game.status != GameStatus.playing) {
      return;
    }

    if (!_isMyTurn) {
      return;
    }

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

    if (legalMoves.isEmpty) {
      return;
    }

    emit(
      GameScreenState(
        game: state.game,
        whiteTime: state.whiteTime,
        blackTime: state.blackTime,
        selectedSquare: square,
        legalMoves: legalMoves,
        promotionFrom: state.promotionFrom,
        promotionTo: state.promotionTo,
        drawOfferState: state.drawOfferState,
      ),
    );
  }

  void _clearSelection() {
    emit(
      GameScreenState(
        game: state.game,
        whiteTime: state.whiteTime,
        blackTime: state.blackTime,
        drawOfferState: state.drawOfferState,
      ),
    );
  }

  void _makeMove(String from, String to) {
    if (!_isMyTurn) return;

    if (_chessRules.isPromotionMove(state.game.fen, from)) {
      emit(
        GameScreenState(
          game: state.game,
          whiteTime: state.whiteTime,
          blackTime: state.blackTime,
          promotionFrom: from,
          promotionTo: to,
          drawOfferState: state.drawOfferState,
        ),
      );
      return;
    }

    _socket.sendMove(from: from, to: to);
    _clearSelection();
  }

  void promote(String promotion) {
    final from = state.promotionFrom;
    final to = state.promotionTo;

    if (from == null || to == null) {
      return;
    }

    _socket.sendMove(from: from, to: to, promotion: promotion);
    _clearSelection();
  }

  bool _isCapture(String previousFen, String newFen) {
    final previousBoard = previousFen.split(' ')[0];
    final newBoard = newFen.split(' ')[0];

    int countPieces(String board) {
      return RegExp(r'[prnbqkPRNBQK]').allMatches(board).length;
    }

    return countPieces(newBoard) < countPieces(previousBoard);
  }

  void _handleMessage(GameSocketMessage message) {
    switch (message) {
      case GameStartedMessage():
        _updateGame(message.game);

      case GameStateMessage():
        _updateGame(message.game);

      case ErrorMessage():
        _handleError(message);

      // REMOVED: DrawOfferedMessage and DrawDeclinedMessage handlers.
      // WHY: The backend broadcasts an authoritative GameStateMessage whenever
      // a draw is offered, accepted, declined, or auto-cleared by a move.
      case UnknownMessage():
      case DrawOfferedMessage():
      case DrawDeclinedMessage():
        break;
    }
  }

  void _handleError(ErrorMessage message) {
    // If a draw offer or response fails on the server (e.g., 'draw_already_offered', 'no_draw_offer'),
    // we recalculate the state based on the current game model to keep the UI in sync.
    if (message.code == 'draw_already_offered' ||
        message.code == 'no_draw_offer' ||
        message.code == 'invalid_draw_response') {
      _updateGame(state.game);
    }
  }

  /// Single source of truth for game updates.
  /// Derives [DrawOfferState] reactively from [game.drawOfferedBy].
  void _updateGame(GameState game) {
    final previousGame = state.game;

    if (previousGame.fen != game.fen) {
      if (_isCapture(previousGame.fen, game.fen)) {
        _soundService.playCapture();
      } else {
        _soundService.playMove();
      }
    }

    DrawOfferState drawState = DrawOfferState.none;

    // Determine draw offer state based on who proposed it:
    if (game.drawOfferedBy != null && game.drawOfferedBy!.isNotEmpty) {
      if (game.drawOfferedBy == _playerId) {
        drawState = DrawOfferState.sent;
      } else {
        drawState = DrawOfferState.received;
        if (game.drawOfferedBy != _playerId) {
          _soundService.playDrawOffer();
        }
      }
    }

    emit(
      GameScreenState(
        game: game,
        whiteTime: game.whiteTime,
        blackTime: game.blackTime,
        selectedSquare: state.selectedSquare,
        legalMoves: state.legalMoves,
        promotionFrom: state.promotionFrom,
        promotionTo: state.promotionTo,
        drawOfferState: drawState,
      ),
    );
  }

  void offerDraw() {
    if (state.game.status != GameStatus.playing) {
      return;
    }

    if (state.drawOfferState != DrawOfferState.none) {
      return;
    }

    // Send command to the socket. We do not immediately emit local state here;
    // we wait for the server's state broadcast to confirm and update the UI.
    _socket.sendOfferDraw();
  }

  void respondToDraw({required bool accepted}) {
    if (state.drawOfferState != DrawOfferState.received) {
      return;
    }

    // Send response command to the socket. The server will broadcast the new state
    // (clearing drawOfferedBy or finishing the game in draw).
    _socket.sendRespondDraw(accepted: accepted);
  }

  void resign() {
    if (state.game.status != GameStatus.playing) {
      return;
    }

    _socket.sendResign();
  }

  @override
  Future<void> close() async {
    await _messageSubscription.cancel();
    _clockTimer?.cancel();
    await _soundService.dispose();

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
