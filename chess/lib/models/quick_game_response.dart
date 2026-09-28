import 'package:chess_app/models/game_session.dart';

class QuickGameResponse {
  final GameSession session;

  const QuickGameResponse({
    required this.session,
  });

  factory QuickGameResponse.fromJson(Map<String, dynamic> json) {
    return QuickGameResponse(
      session: GameSession.fromJson(json),
    );
  }
}