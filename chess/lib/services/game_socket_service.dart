import 'dart:async';
import 'dart:convert';

import 'package:chess/models/game_session.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

import '../models/game_state.dart';

sealed class GameSocketMessage {
  const GameSocketMessage();
}

enum GameSocketStatus { disconnected, connecting, connected, reconnecting }

class GameStartedMessage extends GameSocketMessage {
  final GameState game;

  const GameStartedMessage(this.game);
}

class GameStateMessage extends GameSocketMessage {
  final GameState game;

  const GameStateMessage(this.game);
}

class DrawOfferedMessage extends GameSocketMessage {
  final String playerId;

  const DrawOfferedMessage(this.playerId);
}

class DrawDeclinedMessage extends GameSocketMessage {
  const DrawDeclinedMessage();
}

class ErrorMessage extends GameSocketMessage {
  final String code;
  final String message;

  const ErrorMessage({required this.code, required this.message});
}

class UnknownMessage extends GameSocketMessage {
  const UnknownMessage();
}

class GameSocketService {
  static const String baseUrl = 'ws://10.98.1.3:8080';

  static const List<Duration> _reconnectDelays = [
    Duration(seconds: 2),
    Duration(seconds: 3),
    Duration(seconds: 5),
    Duration(seconds: 8),
    Duration(seconds: 10),
  ];

  GameSession? _session;
  WebSocketChannel? _channel;

  bool _intentionalClose = false;

  int _reconnectAttempt = 0;
  Timer? _reconnectTimer;

  GameSocketStatus _status = GameSocketStatus.disconnected;

  final _messageController = StreamController<GameSocketMessage>.broadcast();

  final _errorController = StreamController<Object>.broadcast();

  final _statusController = StreamController<GameSocketStatus>.broadcast();

  Stream<GameSocketMessage> get messages => _messageController.stream;

  Stream<Object> get errors => _errorController.stream;

  GameSocketStatus get status => _status;

  Stream<GameSocketStatus> get statusStream => _statusController.stream;

  void connect({required GameSession session}) {
    _closeChannel();

    if (_session?.sessionId != session.sessionId) {
      _reconnectAttempt = 0;
    }

    _session = session;
    _intentionalClose = false;

    _reconnectTimer?.cancel();
    _reconnectTimer = null;

    _setStatus(GameSocketStatus.connecting);

    final uri = Uri.parse(
      '$baseUrl/games/${session.gameId}/ws'
      '?sessionId=${session.sessionId}',
    );

    final channel = WebSocketChannel.connect(uri);

    _channel = channel;

    channel.ready
        .then((_) {
          if (_channel != channel || _intentionalClose) {
            return;
          }

          _reconnectAttempt = 0;

          _setStatus(GameSocketStatus.connected);
        })
        .catchError((error) {
          if (_channel != channel || _intentionalClose) {
            return;
          }

          _handleConnectionError(channel, error);
        });

    channel.stream.listen(
      _handleMessage,
      onError: (error) {
        _handleConnectionError(channel, error);
      },
      onDone: () {
        _handleConnectionDone(channel);
      },
    );
  }

  void _setStatus(GameSocketStatus status) {
    _status = status;
    _statusController.add(status);
  }

  void sendMove({required String from, required String to, String? promotion}) {
    final message = {
      'type': 'move',
      'data': {
        'from': from,
        'to': to,
        if (promotion != null) 'promotion': promotion,
      },
    };

    _send(message);
  }

  void sendResign() {
    _send({'type': 'resign'});
  }

  void sendOfferDraw() {
    _send({'type': 'offer_draw'});
  }

  void sendRespondDraw({required bool accepted}) {
    _send({
      'type': 'respond_draw',
      'data': {'accepted': accepted},
    });
  }

  void _send(Map<String, dynamic> message) {
    _channel?.sink.add(jsonEncode(message));
  }

  void _handleMessage(dynamic rawMessage) {
    try {
      final decoded = jsonDecode(rawMessage as String);

      final type = decoded['type'];
      final data = decoded['data'];

      switch (type) {
        case 'game_started':
          _messageController.add(GameStartedMessage(GameState.fromJson(data)));

        case 'game_state':
          _messageController.add(GameStateMessage(GameState.fromJson(data)));

        case 'draw_offered':
          _messageController.add(
            DrawOfferedMessage(data['player_id'] as String),
          );

        case 'draw_declined':
          _messageController.add(const DrawDeclinedMessage());

        case 'error':
          _messageController.add(
            ErrorMessage(
              code: data['code'] as String,
              message: data['message'] as String? ?? '',
            ),
          );

        default:
          _messageController.add(const UnknownMessage());
      }
    } catch (error) {
      _errorController.add(error);
    }
  }

  void _handleConnectionError(WebSocketChannel channel, Object error) {
    if (_channel != channel || _intentionalClose) {
      return;
    }

    _errorController.add(error);

    _scheduleReconnect();
  }

  void _handleConnectionDone(WebSocketChannel channel) {
    if (_channel != channel || _intentionalClose) {
      return;
    }

    _scheduleReconnect();
  }

  void _scheduleReconnect() {
    if (_session == null || _intentionalClose) {
      return;
    }

    // A retry is already scheduled.
    if (_reconnectTimer != null) {
      return;
    }

    if (_reconnectAttempt >= _reconnectDelays.length) {
      _setStatus(GameSocketStatus.disconnected);
      return;
    }

    _setStatus(GameSocketStatus.reconnecting);

    final delay = _reconnectDelays[_reconnectAttempt];

    _reconnectAttempt++;

    _reconnectTimer = Timer(delay, () {
      _reconnectTimer = null;

      final session = _session;

      if (session == null || _intentionalClose) {
        return;
      }

      connect(session: session);
    });
  }

  void retry() {
    final session = _session;

    if (session == null) {
      return;
    }

    _reconnectTimer?.cancel();
    _reconnectTimer = null;

    _reconnectAttempt = 0;
    _intentionalClose = false;

    connect(session: session);
  }

  void close() {
    _intentionalClose = true;

    _reconnectTimer?.cancel();
    _reconnectTimer = null;

    _closeChannel();

    _setStatus(GameSocketStatus.disconnected);
  }

  void _closeChannel() {
    _channel?.sink.close();
    _channel = null;
  }

  Future<void> dispose() async {
    close();

    await _messageController.close();
    await _errorController.close();
    await _statusController.close();
  }
}
