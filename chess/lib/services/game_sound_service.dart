import 'package:audioplayers/audioplayers.dart';

class GameSoundService {
  final AudioPlayer _player = AudioPlayer();

  Future<void> playMove() async {
    await _player.play(AssetSource('sounds/move.wav'));
  }

  Future<void> playCapture() async {
    await _player.play(AssetSource('sounds/capture.wav'));
  }

  Future<void> playLowTime() async {
    await _player.play(AssetSource('sounds/lowtime.wav'));
  }

  Future<void> playDrawOffer() async {
    await _player.play(AssetSource('sounds/drawoffer.wav'));
  }

  Future<void> playTurnSwitch() async {
    await _player.play(AssetSource('sounds/turnswitch.wav'));
  }

  Future<void> dispose() async {
    await _player.dispose();
  }
}
