import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';

class ChessPiece extends StatelessWidget {
  final String type;

  const ChessPiece({super.key, required this.type});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(4),
      child: SvgPicture.asset(
        "assets/pieces/${_color(type)}/$type.svg",
      ),
    );
  }

  String _color(String piece) {
    return piece == piece.toUpperCase() ? "white" : "black";
  }
}
