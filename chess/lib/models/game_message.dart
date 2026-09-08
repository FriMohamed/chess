enum MessageType {
  gameStarted,
  gameState,
  invalidMessage,
  unknown,
}

class WebSocketMessage {
  final MessageType type;
  final dynamic data;

  const WebSocketMessage({
    required this.type,
    this.data,
  });
}