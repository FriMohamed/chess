import 'package:flutter/material.dart';

class LoadingDots extends StatefulWidget {
  const LoadingDots({super.key});

  @override
  State<LoadingDots> createState() => _LoadingDotsState();
}

class _LoadingDotsState extends State<LoadingDots>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();

    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        final step = (_controller.value * 3).floor();

        return Row(
          mainAxisSize: MainAxisSize.min,
          children: List.generate(3, (index) {
            final active = index == step;

            return Padding(
              padding: const EdgeInsets.symmetric(horizontal: 3),
              child: AnimatedOpacity(
                duration: const Duration(milliseconds: 150),
                opacity: active ? 1.0 : 0.3,
                child: Text(
                  '•',
                  style: TextStyle(
                    fontSize: 20,
                    color: colors.primary
                  ),
                ),
              ),
            );
          }),
        );
      },
    );
  }
}