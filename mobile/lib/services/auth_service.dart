import '../models/auth_response.dart';
import 'api_client.dart';
import 'token_storage.dart';

class AuthService {
  const AuthService(this._apiClient, this._tokenStorage);

  final ApiClient _apiClient;
  final TokenStorage _tokenStorage;

  Future<AuthResponse> login({
    required String email,
    required String password,
  }) async {
    final response = await _apiClient.dio.post<Map<String, dynamic>>(
      '/auth/login',
      data: {'email': email, 'password': password},
    );

    final auth = AuthResponse.fromJson(response.data!);
    await _tokenStorage.saveToken(auth.token);
    return auth;
  }
}
