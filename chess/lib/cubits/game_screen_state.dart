import 'package:chess_app/models/game_state.dart';

enum DrawOfferState { none, sent, received, declined }

class GameScreenState {
  final GameState game;

  final int whiteTime;
  final int blackTime;

  final String? selectedSquare;
  final List<String> legalMoves;

  final String? promotionFrom;
  final String? promotionTo;

  final DrawOfferState drawOfferState;

  GameScreenState({
    required this.game,
    int? whiteTime,
    int? blackTime,
    this.selectedSquare,
    this.legalMoves = const [],
    this.promotionFrom,
    this.promotionTo,
    this.drawOfferState = DrawOfferState.none,
  }) : whiteTime = whiteTime ?? game.whiteTime,
       blackTime = blackTime ?? game.blackTime;

  bool get isPromotionPending => promotionFrom != null && promotionTo != null;
}
