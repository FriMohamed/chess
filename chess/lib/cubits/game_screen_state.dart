import 'package:chess/models/game_state.dart';

class GameScreenState {
  final GameState game;

  final int whiteTime;
  final int blackTime;

  final String? selectedSquare;
  final List<String> legalMoves;

  final String? promotionFrom;
  final String? promotionTo;

  GameScreenState({
    required this.game,
    int? whiteTime,
    int? blackTime,
    this.selectedSquare,
    this.legalMoves = const [],
    this.promotionFrom,
    this.promotionTo,
  }): whiteTime = whiteTime ?? game.whiteTime,
       blackTime = blackTime ?? game.blackTime;

  bool get isPromotionPending => promotionFrom != null && promotionTo != null;
}
