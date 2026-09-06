import 'dart:math';
import 'package:shared_preferences/shared_preferences.dart';

class PlayerService {
  static const String _playerNameKey = 'player_name';

  static const List<String> defaultNames = [
    'CleverKnight',
    'SwiftBishop',
    'BraveRook',
    'SilentPawn',
    'RoyalKnight',
    'WiseBishop',
    'BoldRook',
    'LuckyPawn',
  ];

  static String _getRandomName() {
    final random = Random();
    return defaultNames[random.nextInt(defaultNames.length)];
  }

  static Future<String> getPlayerName() async {
    final prefs = await SharedPreferences.getInstance();
    final storedName = prefs.getString(_playerNameKey);

    if (storedName != null && storedName.isNotEmpty) {
      return storedName;
    }

    final randomName = _getRandomName();
    await prefs.setString(_playerNameKey, randomName);
    return randomName;
  }

  static Future<void> savePlayerName(String name) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_playerNameKey, name);
  }
}