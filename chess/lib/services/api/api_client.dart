import 'dart:convert';
import 'package:http/http.dart' as http;

class ApiClient {
  static const String baseUrl = 'http://localhost:8080';

  static Future<T> post<T>(
    String endpoint, {
    Map<String, dynamic>? body,
    required T Function(dynamic json) parser,
  }) async {
    final response = await http.post(
      Uri.parse('$baseUrl$endpoint'),
      headers: {'Content-Type': 'application/json'},
      body: body == null ? null : jsonEncode(body),
    );

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('HTTP ${response.statusCode}: ${response.body}');
    }

    final json = jsonDecode(response.body);

    return parser(json);
  }

  static Future<void> delete(String endpoint) async {
    final response = await http.delete(Uri.parse('$baseUrl$endpoint'));

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('HTTP error: ${response.statusCode} - ${response.body}');
    }
  }
}
