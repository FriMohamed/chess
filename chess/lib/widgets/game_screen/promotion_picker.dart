import 'package:chess/widgets/game_screen/chess_piece.dart';
import 'package:flutter/material.dart';

class PromotionPicker extends StatelessWidget {
  final bool isWhite;
  final void Function(String) onSelected;

  const PromotionPicker({
    super.key,
    required this.isWhite,
    required this.onSelected,
  });

  @override
  Widget build(BuildContext context) {
    const background = Color(0xFF181B21);
    const optionBackground = Color.fromARGB(255, 65, 68, 74);

    final pieces = ['q', 'r', 'b', 'n'];

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFF2A2E36)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          for (final piece in pieces)
            Padding(
              padding: const EdgeInsets.all(4),
              child: InkWell(
                onTap: () => onSelected(piece),
                borderRadius: BorderRadius.circular(10),
                child: Container(
                  width: 64,
                  height: 64,
                  decoration: BoxDecoration(
                    color: optionBackground,
                    borderRadius: BorderRadius.circular(10),
                  ),
                  padding: const EdgeInsets.all(6),
                  child: ChessPiece(
                    type: isWhite ? piece.toUpperCase() : piece,
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
