import 'dart:async';
import 'dart:convert';

import 'package:web_socket_channel/web_socket_channel.dart';

import '../../models/game_state.dart';

sealed class GameSocketMessage {
  const GameSocketMessage();
}

class GameStartedMessage extends GameSocketMessage {
  final GameState game;

  const GameStartedMessage(this.game);
}

class GameStateMessage extends GameSocketMessage {
  final GameState game;

  const GameStateMessage(this.game);
}

class InvalidMessage extends GameSocketMessage {
  const InvalidMessage();
}

class UnknownMessage extends GameSocketMessage {
  const UnknownMessage();
}

class GameSocketService {
  static const String baseUrl = 'ws://192.168.11.108:8080';

  WebSocketChannel? _channel;

  final StreamController<GameSocketMessage> _messageController =
      StreamController<GameSocketMessage>.broadcast();

  final StreamController<Object> _errorController =
      StreamController<Object>.broadcast();

  Stream<GameSocketMessage> get messages => _messageController.stream;

  Stream<Object> get errors => _errorController.stream;

  void connect({required String gameId, required String playerId}) {
    close();

    final uri = Uri.parse('$baseUrl/games/$gameId/ws?playerId=$playerId');

    _channel = WebSocketChannel.connect(uri);

    _channel!.stream.listen(
      _handleMessage,
      onError: (error) {
        _errorController.add(error);
      },
      onDone: () {
        // Connection closed.
      },
    );
  }

  void _handleMessage(dynamic rawMessage) {
    try {
      final decoded = jsonDecode(rawMessage as String);

      final type = decoded['type'];
      final data = decoded['data'];

      switch (type) {
        case 'game_started':
          _messageController.add(GameStartedMessage(GameState.fromJson(data)));
          break;

        case 'game_state':
          _messageController.add(GameStateMessage(GameState.fromJson(data)));
          break;

        case 'invalid_message':
          _messageController.add(const InvalidMessage());
          break;

        default:
          _messageController.add(const UnknownMessage());
      }
    } catch (error) {
      _errorController.add(error);
    }
  }

  void close() {
    _channel?.sink.close();
    _channel = null;
  }

  Future<void> dispose() async {
    close();

    await _messageController.close();
    await _errorController.close();
  }
}
