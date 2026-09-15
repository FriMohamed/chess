import 'package:chess/screens/home_screen.dart';
import 'package:chess/screens/quick_game_screen.dart';
import 'package:chess/theme/app_theme.dart';
import 'package:flutter/material.dart';

void main() {
  runApp(const Myapp());
}

class Myapp extends StatelessWidget {
  const Myapp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      theme: AppTheme.dark,
      home: const HomeScreen(),
      routes: {'/quick-game': (context) => const QuickGameScreen()},
    );
  }
}
