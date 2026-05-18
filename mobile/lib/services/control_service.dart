import '../models/manual_command.dart';
import '../config/app_config.dart';
import 'api_client.dart';

class ControlService {
  const ControlService(this._apiClient);

  final ApiClient _apiClient;

  Future<ManualCommand> sendManualCommand({
    required String action,
    String deviceId = AppConfig.deviceId,
  }) async {
    final response = await _apiClient.dio.post<Map<String, dynamic>>(
      '/control/manual',
      data: {'device_id': deviceId, 'action': action},
    );

    return ManualCommand.fromJson(response.data!);
  }
}
