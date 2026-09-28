import 'package:chess_app/chess/chess_rules.dart';
import 'package:chess_app/widgets/game_screen/chess_square.dart';
import 'package:flutter/material.dart';
import 'package:chess_app/models/game_state.dart';

class ChessBoard extends StatelessWidget {
  final ChessRules _chessRules = ChessRules();
  final GameState game;
  final String? selectedSquare;
  final bool isWhite;
  final void Function(String) onSquareTap;
  final List<String> legalMoves;

  ChessBoard({
    super.key,
    required this.game,
    required this.isWhite,
    required this.selectedSquare,
    required this.legalMoves,
    required this.onSquareTap,
  });

  @override
  Widget build(BuildContext context) {
    final board = _parseFen(game.fen);
    final checkedKingSquare = _chessRules.checkedKingSquare(game.fen);

    return AspectRatio(
      aspectRatio: 1,
      child: GridView.builder(
        physics: const NeverScrollableScrollPhysics(),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 8,
        ),
        itemCount: 64,
        itemBuilder: (context, index) {
          final visualRow = index ~/ 8;
          final visualCol = index % 8;

          final row = isWhite ? visualRow : 7 - visualRow;
          final col = isWhite ? visualCol : 7 - visualCol;

          final piece = board[row][col];
          final isLight = (row + col).isEven;

          final square = '${String.fromCharCode(97 + col)}${8 - row}';
          final isLegalMove = legalMoves.contains(square);

          return ChessSquare(
            piece: piece,
            isLight: isLight,
            isSelected: selectedSquare == square,
            isLegalMove: isLegalMove,
            isInCheck: checkedKingSquare == square,
            onTap: () => onSquareTap(square),
          );
        },
      ),
    );
  }

  List<List<String?>> _parseFen(String fen) {
    final board = List.generate(8, (_) => List<String?>.filled(8, null));

    final position = fen.split(' ').first;
    final ranks = position.split('/');

    for (int row = 0; row < 8; row++) {
      int col = 0;

      for (final char in ranks[row].split('')) {
        if (int.tryParse(char) != null) {
          col += int.parse(char);
        } else {
          board[row][col] = char;
          col++;
        }
      }
    }

    return board;
  }
}
