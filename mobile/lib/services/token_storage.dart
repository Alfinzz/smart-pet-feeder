import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class TokenStorage {
  static const _tokenKey = 'jwt_token';

  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  Future<void> saveToken(String token) {
    return _storage.write(key: _tokenKey, value: token);
  }

  Future<String?> readToken() {
    return _storage.read(key: _tokenKey);
  }

  Future<void> clear() {
    return _storage.delete(key: _tokenKey);
  }
}
