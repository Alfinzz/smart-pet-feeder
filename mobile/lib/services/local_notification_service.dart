import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';

import '../models/dashboard_models.dart';

class LocalNotificationService {
  LocalNotificationService._();

  static final FlutterLocalNotificationsPlugin _plugin =
      FlutterLocalNotificationsPlugin();
  static const AndroidNotificationChannel _androidChannel =
      AndroidNotificationChannel(
        'smart_pet_alerts',
        'Smart Pet Alerts',
        description: 'Food, water, and medical schedule alerts.',
        importance: Importance.high,
      );

  static bool _initialized = false;

  static Future<void> initialize({bool requestPermissions = false}) async {
    if (kIsWeb || _initialized) return;

    const initializationSettings = InitializationSettings(
      android: AndroidInitializationSettings('@mipmap/ic_launcher'),
      iOS: DarwinInitializationSettings(),
    );

    await _plugin.initialize(settings: initializationSettings);
    await _plugin
        .resolvePlatformSpecificImplementation<
          AndroidFlutterLocalNotificationsPlugin
        >()
        ?.createNotificationChannel(_androidChannel);

    if (requestPermissions) {
      await _plugin
          .resolvePlatformSpecificImplementation<
            AndroidFlutterLocalNotificationsPlugin
          >()
          ?.requestNotificationsPermission();
      await _plugin
          .resolvePlatformSpecificImplementation<
            IOSFlutterLocalNotificationsPlugin
          >()
          ?.requestPermissions(alert: true, badge: true, sound: true);
    }

    _initialized = true;
  }

  static Future<void> showAlert(UserAlert alert) async {
    if (alert.title.isEmpty || alert.message.isEmpty) return;
    await initialize();

    const details = NotificationDetails(
      android: AndroidNotificationDetails(
        'smart_pet_alerts',
        'Smart Pet Alerts',
        channelDescription: 'Food, water, and medical schedule alerts.',
        importance: Importance.high,
        priority: Priority.high,
      ),
      iOS: DarwinNotificationDetails(),
    );

    await _plugin.show(
      id: _stableNotificationId(alert.id),
      title: alert.title,
      body: alert.message,
      notificationDetails: details,
      payload: alert.id,
    );
  }

  static int _stableNotificationId(String value) {
    var hash = 0;
    for (final codeUnit in value.codeUnits) {
      hash = (hash * 31 + codeUnit) & 0x7fffffff;
    }
    return hash == 0 ? 1 : hash;
  }
}
