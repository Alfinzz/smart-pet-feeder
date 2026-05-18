import 'package:dio/dio.dart';
import '../config/app_config.dart';
import 'api_error.dart';
import 'api_payload.dart';
import 'api_client.dart';

class SettingsService {
  SettingsService(this._apiClient);

  final ApiClient _apiClient;

  // Pet Details
  Future<void> updatePetDetails(Map<String, dynamic> data) async {
    try {
      final profile = await _fetchProfile();
      final pet = objectMap(profile['pet']);

      await _apiClient.dio.put(
        '/profile/pet',
        data: {
          'name': stringValue(data['name'], pet['name'], 'Fluffy'),
          'species': stringValue(data['species'], pet['species'], 'Dog'),
          'breed': stringValue(data['breed'], pet['breed'], 'Unknown'),
          'age_years': intValue(
            data['age_years'] ?? data['age'],
            pet['age_years'],
          ),
          'weight_kg': doubleValue(
            data['weight_kg'] ?? data['weightKg'],
            pet['weight_kg'],
          ),
          'daily_feed_target_grams': doubleValue(
            data['daily_feed_target_grams'] ?? data['targetDailyPortion'],
            pet['daily_feed_target_grams'],
            fallback: 150,
          ),
          'health_score': intValue(pet['health_score'], null, fallback: 92),
          'health_status': stringValue(pet['health_status'], null, 'Excellent'),
          'health_headline': stringValue(
            pet['health_headline'],
            null,
            'Optimal Wellness',
          ),
          'health_description': stringValue(
            pet['health_description'],
            null,
            'Your pet health metrics are stable this week.',
          ),
          'activity_minutes': intValue(
            pet['activity_minutes'],
            null,
            fallback: 45,
          ),
          'sleep_hours': doubleValue(pet['sleep_hours'], null, fallback: 9.5),
          'device_id': stringValue(pet['device_id'], null, AppConfig.deviceId),
        },
      );
    } on DioException catch (e) {
      throw apiExceptionFromDio(e);
    } catch (e) {
      throw unexpectedApiException(e);
    }
  }

  Future<void> uploadPetPhoto(String imagePath) async {
    try {
      final formData = FormData.fromMap({
        'photo': await MultipartFile.fromFile(
          imagePath,
          filename: _fileNameFromPath(imagePath),
        ),
      });
      await _apiClient.dio.post(
        '/profile/pet/photo',
        data: formData,
        options: Options(contentType: 'multipart/form-data'),
      );
    } on DioException catch (e) {
      throw apiExceptionFromDio(e);
    } catch (e) {
      throw unexpectedApiException(e);
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
      throw apiExceptionFromDio(e);
    } catch (e) {
      throw unexpectedApiException(e);
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
      throw apiExceptionFromDio(e);
    } catch (e) {
      throw unexpectedApiException(e);
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
      throw apiExceptionFromDio(e);
    } catch (e) {
      throw unexpectedApiException(e);
    }
  }

  // Device Settings
  Future<void> updateDeviceSettings(Map<String, dynamic> data) async {
    try {
      await _apiClient.dio.patch(
        '/profile/device-settings',
        data: {
          'name': data['name'] ?? data['deviceName'] ?? '',
          'manual_feed_portion_grams': portionToGrams(
            data['manual_feed_portion_grams'] ?? data['manualPortionSize'],
          ),
        },
      );
    } on DioException catch (e) {
      throw apiExceptionFromDio(e);
    } catch (e) {
      throw unexpectedApiException(e);
    }
  }

  // Calibrate Sensor
  Future<void> calibrateSensor() async {
    try {
      await _apiClient.dio.post('/device/calibrate');
    } on DioException catch (e) {
      throw apiExceptionFromDio(e);
    } catch (e) {
      throw unexpectedApiException(e);
    }
  }

  Future<Map<String, dynamic>> _fetchProfile() async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>('/profile');
    return response.data ?? {};
  }

  String _fileNameFromPath(String path) {
    final normalized = path.replaceAll('\\', '/');
    final parts = normalized.split('/').where((part) => part.isNotEmpty);
    if (parts.isEmpty) return 'pet-photo.jpg';
    return parts.last;
  }
}
