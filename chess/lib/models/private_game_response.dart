import 'game_session.dart';

class PrivateGameResponse {
  final GameSession session;
  final String code;

  const PrivateGameResponse({
    required this.session,  
    required this.code,
  });

  factory PrivateGameResponse.fromJson(Map<String, dynamic> json) {
    return PrivateGameResponse(
      session: GameSession.fromJson(json),
      code: json['code'] as String,
    );
  }
}