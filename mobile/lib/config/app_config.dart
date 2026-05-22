class AppConfig {
  const AppConfig._();

  static const apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://103.47.224.190:8001/api/v1',
  );

  static const deviceId = String.fromEnvironment(
    'DEVICE_ID',
    defaultValue: 'ESP32-001',
  );

  static String publicUrl(String value) {
    final trimmed = value.trim();
    if (trimmed.isEmpty ||
        trimmed.startsWith('http://') ||
        trimmed.startsWith('https://')) {
      return trimmed;
    }

    final baseUri = Uri.tryParse(apiBaseUrl);
    if (baseUri == null || !baseUri.hasScheme || baseUri.host.isEmpty) {
      return trimmed;
    }

    final origin = baseUri.replace(path: '', query: null, fragment: null);
    final originText = origin.toString().replaceFirst(RegExp(r'/$'), '');
    final path = trimmed.startsWith('/') ? trimmed : '/$trimmed';
    return '$originText$path';
  }
}
