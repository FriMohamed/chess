import '../../models/private_game_response.dart';
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

  static Future<PrivateGameResponse> createPrivateGame(String nickname) {
    return ApiClient.post(
      '/games/private',
      body: {'nickname': nickname},
      parser: (json) {
        return PrivateGameResponse.fromJson(json);
      },
    );
  }

  static Future<PrivateGameResponse> joinPrivateGame(
    String nickname,
    String code,
  ) {
    return ApiClient.post(
      '/games/private/join',
      body: {'nickname': nickname, 'code': code},
      parser: (json) {
        return PrivateGameResponse.fromJson(json);
      },
    );
  }

  static Future<void> quitGame({required String sessionId}) {
    return ApiClient.delete('/games?sessionId=$sessionId');
  }
}
