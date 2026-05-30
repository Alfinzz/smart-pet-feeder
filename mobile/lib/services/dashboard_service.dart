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

  Future<CareTask> markCareTaskCompleted(int taskId, {int? ageInMonths}) async {
    final data = <String, dynamic>{'status': 'completed'};
    if (ageInMonths != null) {
      data['age_in_months'] = ageInMonths;
    }

    final response = await _apiClient.dio.patch<Map<String, dynamic>>(
      '/health/tasks/$taskId/status',
      data: data,
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
