class AppConfig {
  const AppConfig._();

  static const apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'https://smart-pet-feeder.alfian-gading.my.id/api/v1',
  );

  static const deviceId = String.fromEnvironment(
    'DEVICE_ID',
    defaultValue: 'ESP32-001',
  );
}
