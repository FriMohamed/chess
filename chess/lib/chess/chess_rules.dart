import 'package:dartchess/dartchess.dart';

class ChessRules {
  List<String> legalMoves(String fen, String from) {
    var chess = Chess.fromSetup(Setup.parseFen(fen));

    final square = Square.fromName(from);
    final piece = chess.board.pieceAt(square);

    if (piece == null) {
      return [];
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

  bool isCheck(String fen) {
    final chess = Chess.fromSetup(Setup.parseFen(fen));

    return chess.isCheck;
  }

  String? checkedKingSquare(String fen) {
    final chess = Chess.fromSetup(Setup.parseFen(fen));

    if (!chess.isCheck) {
      return null;
    }

    final kingSquares = chess.board.kings.intersect(chess.board.occupied);

    for (final square in kingSquares.squares) {
      final piece = chess.board.pieceAt(square);

      if (piece?.color == chess.turn) {
        return square.name;
      }
    }

    return null;
  }
}
