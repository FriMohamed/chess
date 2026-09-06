import 'package:flutter/material.dart';

class CustomButton extends StatefulWidget {
  final String label;
  final Color backgroundColor;
  final Color textColor;
  final Color? borderColor;
  final FontWeight fontWeight;
  final VoidCallback onPressed;

  const CustomButton({
    required this.label,
    required this.backgroundColor,
    required this.textColor,
    required this.fontWeight,
    required this.onPressed,
    this.borderColor,
  });

  @override
  State<CustomButton> createState() => _CustomButtonState();
}

class _CustomButtonState extends State<CustomButton> {
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _isPressed = true),
      onTapUp: (_) => setState(() => _isPressed = false),
      onTapCancel: () => setState(() => _isPressed = false),
      onTap: widget.onPressed,
      child: AnimatedOpacity(
        duration: const Duration(milliseconds: 100),
        opacity: _isPressed ? 0.7 : 1.0,
        child: Container(
          width: double.infinity,
          constraints: const BoxConstraints(minHeight: 58),
          decoration: BoxDecoration(
            color: widget.backgroundColor,
            borderRadius: BorderRadius.circular(14),
            border: widget.borderColor != null
                ? Border.all(color: widget.borderColor!)
                : null,
          ),
          alignment: Alignment.center,
          child: Text(
            widget.label,
            style: TextStyle(
              color: widget.textColor,
              fontSize: 15,
              fontWeight: widget.fontWeight,
              letterSpacing: 1,
            ),
          ),
        ),
      ),
    );
  }
}