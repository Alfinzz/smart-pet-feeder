import 'package:dio/dio.dart';
import '../config/app_config.dart';
import 'token_storage.dart';

class ApiClient {
  ApiClient(this._tokenStorage, {Dio? dio})
    : dio =
          dio ??
          Dio(
            BaseOptions(
              baseUrl: baseUrl,
              connectTimeout: const Duration(seconds: 10),
              receiveTimeout: const Duration(seconds: 10),
              headers: {'Content-Type': 'application/json'},
            ),
          ) {
    this.dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await _tokenStorage.readToken();
          if (token != null && token.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
      ),
    );
  }

  static const baseUrl = AppConfig.apiBaseUrl;

  final TokenStorage _tokenStorage;
  final Dio dio;
}
