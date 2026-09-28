import 'package:chess/chess.dart' as chess;

class ChessRules {
  
  List<String> legalMoves(String fen, String from) {
    final game = chess.Chess.fromFEN(fen);

    final moves = game.generate_moves({'square': from});

    return moves.map((move) => move.toAlgebraic).toList();
  }

  bool isPromotionMove(String fen, String from) {
    final game = chess.Chess.fromFEN(fen);

    final piece = game.get(from);

    if (piece == null) {
      return false;
    }

    if (piece.type != chess.Chess.PAWN) {
      return false;
    }

    // White pawn promotes from rank 7.
    // Black pawn promotes from rank 2.
    final rank = from[1];

    if (piece.color == chess.Chess.WHITE) {
      return rank == '7';
    }

    return rank == '2';
  }

  bool isCheck(String fen) {
    final game = chess.Chess.fromFEN(fen);

    return game.in_check;
  }

  String? checkedKingSquare(String fen) {
    final game = chess.Chess.fromFEN(fen);

    if (!game.in_check) {
      return null;
    }

    // The king of the side whose turn it is.
    final kingSquare = game.kings[game.turn];

    if (kingSquare == chess.Chess.EMPTY) {
      return null;
    }

    return chess.Chess.algebraic(kingSquare);
  }
}
