class QuickGameResponse {
  final String gameId;
  final String playerId;

  const QuickGameResponse({
    required this.gameId,
    required this.playerId,
  });

  factory QuickGameResponse.fromJson(Map<String, dynamic> json) {
    return QuickGameResponse(
      gameId: json['game_id'] as String,
      playerId: json['player_id'] as String,
    );
  }
}