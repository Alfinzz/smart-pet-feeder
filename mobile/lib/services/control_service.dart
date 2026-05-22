import '../models/manual_command.dart';
import 'api_client.dart';

class ControlService {
  const ControlService(this._apiClient);

  final ApiClient _apiClient;

  Future<ManualCommand> sendManualCommand({
    required String action,
    String? deviceId,
  }) async {
    final payload = <String, dynamic>{'action': action};
    final trimmedDeviceId = deviceId?.trim() ?? '';
    if (trimmedDeviceId.isNotEmpty) {
      payload['device_id'] = trimmedDeviceId;
    }

    final response = await _apiClient.dio.post<Map<String, dynamic>>(
      '/control/manual',
      data: payload,
    );

    return ManualCommand.fromJson(response.data!);
  }

  Future<ManualCommand> fetchManualCommand(int commandId) async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>(
      '/control/manual/$commandId',
    );

    return ManualCommand.fromJson(response.data!);
  }
}
