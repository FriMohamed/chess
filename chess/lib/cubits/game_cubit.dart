import 'package:chess/chess/chess_rules.dart';
import 'package:chess/cubits/game_screen_state.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:chess/models/game_state.dart';

class GameCubit extends Cubit<GameScreenState> {
  final ChessRules _chessRules = ChessRules();

  GameCubit(GameState game) : super(GameScreenState(game: game));

  void selectSquare(String square) {
    final selected = state.selectedSquare;

    if (selected == null) {
      _selectPiece(square);
      return;
    }

    if (state.selectedSquare == square) {
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
    final newFen = _chessRules.makeMove(state.game.fen, from, to);

    final newGame = GameState(
      gameId: state.game.gameId,
      white: state.game.white,
      black: state.game.black,
      fen: newFen,
      whiteTime: state.game.whiteTime,
      blackTime: state.game.blackTime,
      active: state.game.active,
    );

    emit(GameScreenState(game: newGame));
  }
}
