import '../models/dashboard_models.dart';
import 'api_client.dart';

class DashboardService {
  const DashboardService(this._apiClient);

  final ApiClient _apiClient;

  Future<DashboardOverview> fetchOverview() async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>(
      '/dashboard/overview',
    );

    return DashboardOverview.fromJson(response.data!);
  }

  Future<WeeklyConsumption> fetchWeeklyConsumption({int days = 7}) async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>(
      '/feed/weekly-consumption',
      queryParameters: {'days': days},
    );

    return WeeklyConsumption.fromJson(response.data!);
  }

  Future<HealthSummary> fetchHealthSummary() async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>(
      '/health/summary',
    );

    return HealthSummary.fromJson(response.data!);
  }

  Future<CareTask> markCareTaskCompleted(int taskId) async {
    final response = await _apiClient.dio.patch<Map<String, dynamic>>(
      '/health/tasks/$taskId/status',
      data: {'status': 'completed'},
    );

    return CareTask.fromJson(response.data!);
  }

  Future<CareTask> createCareTask({
    required String category,
    required String title,
    required String description,
    required DateTime dueDate,
    String status = 'pending',
    String priority = 'normal',
  }) async {
    final response = await _apiClient.dio.post<Map<String, dynamic>>(
      '/health/tasks',
      data: {
        'category': category,
        'title': title,
        'description': description,
        'due_date': _formatApiDate(dueDate),
        'status': status,
        'priority': priority,
      },
    );

    return CareTask.fromJson(response.data!);
  }

  Future<List<UserAlert>> fetchAlerts() async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>('/alerts');
    final items = response.data?['data'] as List<dynamic>? ?? [];
    return items
        .map((item) => UserAlert.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<ProfileSummary> fetchProfile() async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>('/profile');

    return ProfileSummary.fromJson(response.data!);
  }
}

String _formatApiDate(DateTime value) {
  final month = value.month.toString().padLeft(2, '0');
  final day = value.day.toString().padLeft(2, '0');
  return '${value.year}-$month-$day';
}
