class GameSession {
  final String gameId;
  final String sessionId;
  final String playerId;

  const GameSession({
    required this.gameId,
    required this.sessionId,
    required this.playerId,
  });

  factory GameSession.fromJson(Map<String, dynamic> json) {
    return GameSession(
      gameId: json['game_id'] as String,
      sessionId: json['session_id'] as String,
      playerId: json['player_id'] as String,
    );
  }
}
