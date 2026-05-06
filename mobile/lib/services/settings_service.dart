import 'package:dio/dio.dart';
import 'api_client.dart';

class SettingsService {
  SettingsService(this._apiClient);

  final ApiClient _apiClient;

  // Pet Details
  Future<void> updatePetDetails(Map<String, dynamic> data) async {
    try {
      // Mocking endpoint, adjust to actual backend endpoint if different
      await _apiClient.dio.put('/pet/details', data: data);
    } on DioException catch (e) {
      throw _handleError(e);
    } catch (e) {
      throw Exception('Unexpected error occurred: $e');
    }
  }

  // Notification Preferences
  Future<void> updateNotificationPreferences(Map<String, dynamic> data) async {
    try {
      // Mocking endpoint, adjust to actual backend endpoint if different
      await _apiClient.dio.put('/user/preferences', data: data);
    } on DioException catch (e) {
      throw _handleError(e);
    } catch (e) {
      throw Exception('Unexpected error occurred: $e');
    }
  }

  // Health Vitals
  Future<void> submitVitals(double weight, int activity, double sleep) async {
    try {
      await _apiClient.dio.post(
        '/health/vitals',
        data: {
          'weight': weight,
          'activity': activity,
          'sleep': sleep,
        },
      );
    } on DioException catch (e) {
      throw _handleError(e);
    } catch (e) {
      throw Exception('Unexpected error occurred: $e');
    }
  }

  // Device Settings
  Future<void> updateDeviceSettings(Map<String, dynamic> data) async {
    try {
      // Mocking endpoint, adjust to actual backend endpoint if different
      await _apiClient.dio.put('/device/settings', data: data);
    } on DioException catch (e) {
      throw _handleError(e);
    } catch (e) {
      throw Exception('Unexpected error occurred: $e');
    }
  }

  // Calibrate Sensor
  Future<void> calibrateSensor() async {
    try {
      await _apiClient.dio.post('/device/calibrate');
    } on DioException catch (e) {
      throw _handleError(e);
    } catch (e) {
      throw Exception('Unexpected error occurred: $e');
    }
  }

  Exception _handleError(DioException e) {
    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.receiveTimeout ||
        e.type == DioExceptionType.connectionError) {
      return Exception('Connection failed. Is the server running?');
    }
    
    if (e.response != null) {
      final data = e.response?.data;
      final message = data is Map<String, dynamic> && data.containsKey('message') 
          ? data['message'] 
          : 'Server error: ${e.response?.statusCode}';
      return Exception(message);
    }
    
    return Exception(e.message ?? 'Unknown network error');
  }
}
