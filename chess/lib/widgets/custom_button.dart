import 'package:flutter/material.dart';

class CustomButton extends StatelessWidget {
  final String label;
  final Color? backgroundColor;
  final Color? textColor;
  final Color? borderColor;
  final FontWeight fontWeight;
  final VoidCallback onPressed;

  const CustomButton({
    super.key,
    required this.label,
    required this.fontWeight,
    required this.onPressed,
    this.backgroundColor,
    this.textColor,
    this.borderColor,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      height: 58,
      child: ElevatedButton(
        onPressed: onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: backgroundColor,
          foregroundColor: textColor,
          shape: borderColor != null
              ? RoundedRectangleBorder(
                  borderRadius: const BorderRadius.all(Radius.circular(14)),
                  side: BorderSide(color: borderColor!),
                )
              : null,
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 15,
            fontWeight: fontWeight,
            letterSpacing: 1,
          ),
        ),
      ),
    );
  }
}