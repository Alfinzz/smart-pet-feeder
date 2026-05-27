import 'dart:ui';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:workmanager/workmanager.dart';

import 'api_client.dart';
import 'dashboard_service.dart';
import 'local_notification_service.dart';
import 'token_storage.dart';

const smartPetAlertPollingTask = 'smartPetAlertPollingTask';
const _smartPetAlertPollingUniqueName = 'smartPetAlertPolling';

@pragma('vm:entry-point')
void smartPetAlertCallbackDispatcher() {
  Workmanager().executeTask((task, inputData) async {
    try {
      WidgetsFlutterBinding.ensureInitialized();
      DartPluginRegistrant.ensureInitialized();

      final tokenStorage = TokenStorage();
      final token = await tokenStorage.readToken();
      if (token == null || token.isEmpty) return true;

      final dashboardService = DashboardService(ApiClient(tokenStorage));
      final alerts = await dashboardService.fetchAlerts();
      await LocalNotificationService.initialize();
      for (final alert in alerts) {
        await LocalNotificationService.showAlert(alert);
      }
      return true;
    } catch (_) {
      return false;
    }
  });
}

class AlertPollingService {
  AlertPollingService._();

  static Future<void> initialize() async {
    if (kIsWeb) return;
    if (defaultTargetPlatform != TargetPlatform.android &&
        defaultTargetPlatform != TargetPlatform.iOS) {
      return;
    }

    try {
      await LocalNotificationService.initialize(requestPermissions: true);
      await Workmanager().initialize(smartPetAlertCallbackDispatcher);
      await Workmanager().registerPeriodicTask(
        _smartPetAlertPollingUniqueName,
        smartPetAlertPollingTask,
        frequency: const Duration(minutes: 15),
        constraints: Constraints(networkType: NetworkType.connected),
      );
    } catch (_) {
      // Background polling is best effort; the foreground app should still open.
    }
  }
}
