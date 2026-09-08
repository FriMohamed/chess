class Player {
  final String id;
  final String nickname;

  const Player({
    required this.id,
    required this.nickname,
  });

  factory Player.fromJson(Map<String, dynamic> json) {
    return Player(
      id: json['id'] as String,
      nickname: json['nickname'] as String,
    );
  }
}