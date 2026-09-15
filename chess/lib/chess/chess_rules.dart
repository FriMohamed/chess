import 'package:dartchess/dartchess.dart';

class ChessRules {
  List<String> legalMoves(String fen, String from) {
    var chess = Chess.fromSetup(Setup.parseFen(fen));

    final square = Square.fromName(from);
    final piece = chess.board.pieceAt(square);

    if (piece == null) {
      return [];
    }

    if (piece.color != chess.turn) {
      chess = chess.copyWith(turn: piece.color, epSquare: null) as Chess;
    }

    return chess
        .legalMovesOf(square)
        .squares
        .map((square) => square.name)
        .toList();
  }

  bool isPromotionMove(String fen, String from) {
    final setup = Setup.parseFen(fen);
    final chess = Chess.fromSetup(setup);

    final fromSquare = Square.fromName(from);

    if (!chess.board.pawns.has(fromSquare)) {
      return false;
    }

    final promotionRank = chess.turn == Side.white ? 6 : 1;

    return fromSquare.rank == promotionRank;
  }
}
