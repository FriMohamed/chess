import 'package:chess/models/game_state.dart';

class GameScreenState {
  final GameState game;
  final String? selectedSquare;
  final List<String> legalMoves;

  const GameScreenState({
    required this.game,
    this.selectedSquare,
    this.legalMoves = const [],
  });
}