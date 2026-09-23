import 'player.dart';

enum ActiveColor { white, black }

enum GameStatus { waiting, playing, finished }

enum GameResult { none, whiteWins, blackWins, draw }

enum EndReason {
  none,
  checkmate,
  stalemate,
  threefoldRepetition,
  fivefoldRepetition,
  fiftyMoveRule,
  seventyFiveMoveRule,
  insufficientMaterial,
  resignation,
  drawAgreement,
  disconnect,
  quit,
  timeout,
}

class GameState {
  final String gameId;

  final Player? white;
  final Player? black;

  final String fen;

  final int whiteTime;
  final int blackTime;

  final ActiveColor active;

  final GameStatus status;
  final GameResult result;
  final EndReason endReason;
  final bool? check;
  final String? drawOfferedBy;

  const GameState({
    required this.gameId,
    required this.white,
    required this.black,
    required this.fen,
    required this.whiteTime,
    required this.blackTime,
    required this.active,
    required this.status,
    required this.result,
    required this.endReason,
    this.check,
    this.drawOfferedBy,
  });

  factory GameState.fromJson(Map<String, dynamic> json) {
    return GameState(
      gameId: json['game_id'] as String,

      white: json['white'] != null ? Player.fromJson(json['white']) : null,

      black: json['black'] != null ? Player.fromJson(json['black']) : null,

      fen: json['fen'] as String,

      whiteTime: json['white_time'] as int,
      blackTime: json['black_time'] as int,

      active: (json['active'] as int) == 0
          ? ActiveColor.white
          : ActiveColor.black,

      status: _gameStatus(json['status'] as String),
      result: _gameResult(json['result'] as String),
      endReason: _endReason(json['end_reason'] as String),

      check: json['check'] as bool,
      drawOfferedBy: json['draw_offered_by'] as String?,
    );
  }

  static GameStatus _gameStatus(String value) {
    return switch (value) {
      'waiting' => GameStatus.waiting,
      'playing' => GameStatus.playing,
      'finished' => GameStatus.finished,
      _ => GameStatus.waiting,
    };
  }

  static GameResult _gameResult(String value) {
    return switch (value) {
      'none' => GameResult.none,
      'white_wins' => GameResult.whiteWins,
      'black_wins' => GameResult.blackWins,
      'draw' => GameResult.draw,
      _ => GameResult.none,
    };
  }

  static EndReason _endReason(String value) {
    return switch (value) {
      'none' => EndReason.none,
      'checkmate' => EndReason.checkmate,
      'stalemate' => EndReason.stalemate,
      'threefold_repetition' => EndReason.threefoldRepetition,
      'fivefold_repetition' => EndReason.fivefoldRepetition,
      'fifty_move_rule' => EndReason.fiftyMoveRule,
      'seventy_five_move_rule' => EndReason.seventyFiveMoveRule,
      'insufficient_material' => EndReason.insufficientMaterial,
      'resignation' => EndReason.resignation,
      'draw_agreement' => EndReason.drawAgreement,
      'disconnect' => EndReason.disconnect,
      'quit' => EndReason.quit,
      'timeout' => EndReason.timeout,
      _ => EndReason.none,
    };
  }
}
