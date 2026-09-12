import 'package:chess/widgets/game_screen/chess_square.dart';
import 'package:flutter/material.dart';
import 'package:chess/models/game_state.dart';

class ChessBoard extends StatelessWidget {
  final GameState game;
  final String? selectedSquare;
  final void Function(String) onSquareTap;
  final List<String> legalMoves;

  const ChessBoard({
    super.key,
    required this.game,
    required this.selectedSquare,
    required this.legalMoves,
    required this.onSquareTap,
  });

  @override
  Widget build(BuildContext context) {
    final board = _parseFen(game.fen);

    return AspectRatio(
      aspectRatio: 1,
      child: GridView.builder(
        physics: const NeverScrollableScrollPhysics(),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 8,
        ),
        itemCount: 64,
        itemBuilder: (context, index) {
          final row = index ~/ 8;
          final col = index % 8;

          final piece = board[row][col];
          final isLight = (row + col).isEven;

          final square = '${String.fromCharCode(97 + col)}${8 - row}';
          final isLegalMove = legalMoves.contains(square);

          return ChessSquare(
            piece: piece,
            isLight: isLight,
            isSelected: selectedSquare == square,
            isLegalMove: isLegalMove,
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
