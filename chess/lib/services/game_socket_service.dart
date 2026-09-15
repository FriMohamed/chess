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

class ErrorMessage extends GameSocketMessage {
  final String message;

  const ErrorMessage(this.message);
}

class UnknownMessage extends GameSocketMessage {
  const UnknownMessage();
}

class GameSocketService {
  static const String baseUrl = 'ws://192.168.11.108:8080';

  WebSocketChannel? _channel;

  final _messageController =
      StreamController<GameSocketMessage>.broadcast();

  final _errorController =
      StreamController<Object>.broadcast();

  Stream<GameSocketMessage> get messages =>
      _messageController.stream;

  Stream<Object> get errors =>
      _errorController.stream;

  void connect({
    required String gameId,
    required String playerId,
  }) {
    close();

    final uri = Uri.parse(
      '$baseUrl/games/$gameId/ws?playerId=$playerId',
    );

    final channel = WebSocketChannel.connect(uri);

    _channel = channel;

    channel.stream.listen(
      _handleMessage,
      onError: _handleError,
      onDone: _handleDone,
    );
  }

  void sendMove({
    required String from,
    required String to,
    String? promotion,
  }) {
    final message = {
      'type': 'move',
      'data': {
        'from': from,
        'to': to,
        if (promotion != null) 'promotion': promotion,
      },
    };

    _channel?.sink.add(jsonEncode(message));
  }

  void _handleMessage(dynamic rawMessage) {
    try {
      final decoded = jsonDecode(rawMessage as String);

      final type = decoded['type'];
      final data = decoded['data'];

      switch (type) {
        case 'game_started':
          _messageController.add(
            GameStartedMessage(
              GameState.fromJson(data),
            ),
          );

        case 'game_state':
          _messageController.add(
            GameStateMessage(
              GameState.fromJson(data),
            ),
          );

        case 'error':
          _messageController.add(
            ErrorMessage(
              data['message'] as String,
            ),
          );

        default:
          _messageController.add(
            const UnknownMessage(),
          );
      }
    } catch (error) {
      _errorController.add(error);
    }
  }

  void _handleError(Object error) {
    _errorController.add(error);
  }

  void _handleDone() {
    // Reconnection will be handled later.
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