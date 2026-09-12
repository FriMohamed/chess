import 'package:chess/widgets/game_screen/chess_piece.dart';
import 'package:flutter/material.dart';

class ChessSquare extends StatelessWidget {
  final String? piece;
  final bool isLight;
  final bool isSelected;
  final bool isLegalMove;

  final void Function() onTap;

  const ChessSquare({
    super.key,
    required this.piece,
    required this.isLight,
    required this.isSelected,
    required this.isLegalMove,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        color: isSelected
            ? const Color.fromARGB(205, 255, 231, 46)
            : isLight
            ? const Color(0xFFF0D9B5)
            : const Color(0xFFB58863),
        alignment: Alignment.center,
        child: Stack(
          alignment: Alignment.center,
          children: [
            if (piece != null)
              AnimatedScale(
                scale: isSelected ? 1.25 : 1.0,
                duration: const Duration(milliseconds: 120),
                curve: Curves.easeOut,
                child: ChessPiece(type: piece!),
              ),

            if (isLegalMove)
              Container(
                width: 12,
                height: 12,
                decoration: BoxDecoration(
                  color: Colors.black.withValues(alpha: 0.25),
                  shape: BoxShape.circle,
                ),
              ),
          ],
        ),
      ),
    );
  }
}

//  carriere@amanyspharma.com
