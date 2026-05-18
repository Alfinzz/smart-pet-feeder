import 'package:dio/dio.dart';

Exception apiExceptionFromDio(DioException error) {
  if (error.type == DioExceptionType.connectionTimeout ||
      error.type == DioExceptionType.receiveTimeout ||
      error.type == DioExceptionType.connectionError) {
    return Exception(
      'Koneksi ke server gagal. Periksa internet dan coba lagi.',
    );
  }

  if (error.response != null) {
    final data = error.response?.data;
    final message = data is Map<String, dynamic>
        ? data['error'] ?? data['message']
        : 'Server error: ${error.response?.statusCode}';
    return Exception(
      message?.toString() ?? 'Server error: ${error.response?.statusCode}',
    );
  }

  return Exception(error.message ?? 'Unknown network error');
}

Exception unexpectedApiException(Object error) {
  return Exception('Unexpected error occurred: $error');
}
