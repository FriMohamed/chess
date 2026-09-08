import '../../models/quick_game_response.dart';
import 'api_client.dart';

class GameApiService {
  static Future<QuickGameResponse> quickGame(String nickname) {
    return ApiClient.post(
      '/games/quick',
      body: {'nickname': nickname},
      parser: (json) {
        return QuickGameResponse.fromJson(json);
      },
    );
  }

  static Future<void> quitGame({
    required String gameId,
    required String playerId,
  }) {
    print('gameId: $gameId');
    print('playerId: $playerId');
    return ApiClient.delete('/games/$gameId/players/$playerId');
  }
}
