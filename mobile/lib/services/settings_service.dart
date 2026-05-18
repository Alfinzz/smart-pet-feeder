import 'package:dio/dio.dart';
import 'api_client.dart';

class SettingsService {
  SettingsService(this._apiClient);

  final ApiClient _apiClient;

  // Pet Details
  Future<void> updatePetDetails(Map<String, dynamic> data) async {
    try {
      final profile = await _fetchProfile();
      final pet = _objectMap(profile['pet']);

      await _apiClient.dio.put(
        '/profile/pet',
        data: {
          'name': _stringValue(data['name'], pet['name'], 'Fluffy'),
          'species': _stringValue(data['species'], pet['species'], 'Dog'),
          'breed': _stringValue(data['breed'], pet['breed'], 'Unknown'),
          'age_years': _intValue(
            data['age_years'] ?? data['age'],
            pet['age_years'],
          ),
          'weight_kg': _doubleValue(
            data['weight_kg'] ?? data['weightKg'],
            pet['weight_kg'],
          ),
          'daily_feed_target_grams': _doubleValue(
            data['daily_feed_target_grams'] ?? data['targetDailyPortion'],
            pet['daily_feed_target_grams'],
            fallback: 150,
          ),
          'health_score': _intValue(pet['health_score'], null, fallback: 92),
          'health_status': _stringValue(
            pet['health_status'],
            null,
            'Excellent',
          ),
          'health_headline': _stringValue(
            pet['health_headline'],
            null,
            'Optimal Wellness',
          ),
          'health_description': _stringValue(
            pet['health_description'],
            null,
            'Your pet health metrics are stable this week.',
          ),
          'activity_minutes': _intValue(
            pet['activity_minutes'],
            null,
            fallback: 45,
          ),
          'sleep_hours': _doubleValue(pet['sleep_hours'], null, fallback: 9.5),
          'device_id': _stringValue(pet['device_id'], null, 'ESP32-001'),
        },
      );
    } on DioException catch (e) {
      throw _handleError(e);
    } catch (e) {
      throw Exception('Unexpected error occurred: $e');
    }
  }

  // Notification Preferences
  Future<Map<String, dynamic>> fetchNotificationPreferences() async {
    try {
      final response = await _apiClient.dio.get<Map<String, dynamic>>(
        '/profile/notification-preferences',
      );
      return response.data ?? {};
    } on DioException catch (e) {
      throw _handleError(e);
    } catch (e) {
      throw Exception('Unexpected error occurred: $e');
    }
  }

  Future<void> updateNotificationPreferences(Map<String, dynamic> data) async {
    try {
      await _apiClient.dio.put(
        '/profile/notification-preferences',
        data: {
          'low_food_alert':
              data['low_food_alert'] ?? data['lowFoodAlert'] ?? false,
          'empty_water_alert':
              data['empty_water_alert'] ?? data['emptyWaterAlert'] ?? false,
          'feeding_success_report':
              data['feeding_success_report'] ??
              data['feedSuccessReport'] ??
              false,
        },
      );
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
          'weight_kg': weight,
          'activity_minutes': activity,
          'sleep_hours': sleep,
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
      await _apiClient.dio.patch(
        '/profile/device-settings',
        data: {
          'name': data['name'] ?? data['deviceName'] ?? '',
          'manual_feed_portion_grams': _portionToGrams(
            data['manual_feed_portion_grams'] ?? data['manualPortionSize'],
          ),
        },
      );
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

  Future<Map<String, dynamic>> _fetchProfile() async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>('/profile');
    return response.data ?? {};
  }

  Map<String, dynamic> _objectMap(Object? value) {
    if (value is Map<String, dynamic>) return value;
    if (value is Map) return Map<String, dynamic>.from(value);
    return {};
  }

  String _stringValue(Object? value, Object? fallbackValue, String fallback) {
    final candidate = value ?? fallbackValue;
    if (candidate is String && candidate.trim().isNotEmpty) {
      return candidate.trim();
    }
    return fallback;
  }

  int _intValue(Object? value, Object? fallbackValue, {int fallback = 0}) {
    final candidate = value ?? fallbackValue;
    if (candidate is num) return candidate.toInt();
    if (candidate is String) return int.tryParse(candidate) ?? fallback;
    return fallback;
  }

  double _doubleValue(
    Object? value,
    Object? fallbackValue, {
    double fallback = 0,
  }) {
    final candidate = value ?? fallbackValue;
    if (candidate is num) return candidate.toDouble();
    if (candidate is String) return double.tryParse(candidate) ?? fallback;
    return fallback;
  }

  double _portionToGrams(Object? value) {
    if (value is num) return value.toDouble();
    if (value is String) {
      switch (value) {
        case '1/8c':
          return 15;
        case '1/4c':
          return 30;
        case '1/2c':
          return 60;
        case '1c':
          return 120;
      }
      return double.tryParse(value) ?? 30;
    }
    return 30;
  }

  Exception _handleError(DioException e) {
    if (e.type == DioExceptionType.connectionTimeout ||
        e.type == DioExceptionType.receiveTimeout ||
        e.type == DioExceptionType.connectionError) {
      return Exception(
        'Koneksi ke server gagal. Periksa internet dan coba lagi.',
      );
    }

    if (e.response != null) {
      final data = e.response?.data;
      final message = data is Map<String, dynamic>
          ? data['error'] ?? data['message']
          : 'Server error: ${e.response?.statusCode}';
      return Exception(
        message?.toString() ?? 'Server error: ${e.response?.statusCode}',
      );
    }

    return Exception(e.message ?? 'Unknown network error');
  }
}
