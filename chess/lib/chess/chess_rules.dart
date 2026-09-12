import 'package:dartchess/dartchess.dart';

class ChessRules {
  
  List<String> legalMoves(String fen, String from) {
    final setup = Setup.parseFen(fen);
    final position = Chess.fromSetup(setup);

    final square = Square.fromName(from);
    final moves = position.legalMovesOf(square);

    if (moves == null) {
      return [];
    }

    return moves.squares.map((square) => square.name).toList();
  }

  String makeMove(String fen, String from, String to) {
    final setup = Setup.parseFen(fen);
    final position = Chess.fromSetup(setup);

    final move = NormalMove(
      from: Square.fromName(from),
      to: Square.fromName(to),
    );

    final newPosition = position.play(move);

    return newPosition.fen;
  }
}
