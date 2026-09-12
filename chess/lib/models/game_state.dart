import 'player.dart';

enum ActiveColor {
  white,
  black,
}

class GameState {
  final String gameId;
  final Player? white;
  final Player? black;
  final String fen;
  final int whiteTime;
  final int blackTime;
  final ActiveColor active;

  const GameState({
    required this.gameId,
    required this.white,
    required this.black,
    required this.fen,
    required this.whiteTime,
    required this.blackTime,
    required this.active,
  });

  factory GameState.fromJson(Map<String, dynamic> json) {
    return GameState(
      gameId: json['game_id'] as String,
      white: json['white'] != null ? Player.fromJson(json['white']) : null,
      black: json['black'] != null ? Player.fromJson(json['black']) : null,
      fen: json['fen'] as String,
      whiteTime: json['white_time'] as int,
      blackTime: json['black_time'] as int,
      active: (json['active'] as int) == 0 ? ActiveColor.black : ActiveColor.white,
    );
  }
}
