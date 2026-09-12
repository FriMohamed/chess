import 'package:chess/models/game_state.dart';
import 'package:chess/models/player.dart';
import 'package:chess/screens/game_screen.dart';
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
      const testGame = GameState(
      gameId: 'test-game-7f3a92',
      white: Player(id: 'player-white-123', nickname: 'Mohamed'),
      black: Player(id: 'player-black-456', nickname: 'Opponent'),
      fen: 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR',
      whiteTime: 600,
      blackTime: 600,
      active: ActiveColor.white,
    );

    return MaterialApp(
      debugShowCheckedModeBanner: false,
      theme: AppTheme.dark,
      home: const GameScreen(
        game: testGame,
        playerId: "string"
      ),
      routes: {'/quick-game': (context) => const QuickGameScreen()},
    );
  }
}
