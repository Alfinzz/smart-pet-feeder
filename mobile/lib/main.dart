import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import 'screens/login_screen.dart';
import 'screens/main_shell.dart';
import 'services/alert_polling_service.dart';
import 'services/api_client.dart';
import 'services/auth_service.dart';
import 'services/control_service.dart';
import 'services/dashboard_service.dart';
import 'services/feed_service.dart';
import 'services/settings_service.dart';
import 'services/token_storage.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await AlertPollingService.initialize();

  final tokenStorage = TokenStorage();
  final apiClient = ApiClient(tokenStorage);

  runApp(
    SmartPetApp(
      tokenStorage: tokenStorage,
      authService: AuthService(apiClient, tokenStorage),
      feedService: FeedService(apiClient),
      controlService: ControlService(apiClient),
      dashboardService: DashboardService(apiClient),
      settingsService: SettingsService(apiClient),
    ),
  );
}

class SmartPetApp extends StatelessWidget {
  const SmartPetApp({
    super.key,
    required this.tokenStorage,
    required this.authService,
    required this.feedService,
    required this.controlService,
    required this.dashboardService,
    required this.settingsService,
  });

  final TokenStorage tokenStorage;
  final AuthService authService;
  final FeedService feedService;
  final ControlService controlService;
  final DashboardService dashboardService;
  final SettingsService settingsService;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Smart Pet Feeder',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF1565C0),
          brightness: Brightness.light,
        ),
        scaffoldBackgroundColor: const Color(0xFFF5F7FA),
        textTheme: GoogleFonts.interTextTheme(Theme.of(context).textTheme),
        useMaterial3: true,
      ),
      routes: {
        '/login': (_) => LoginScreen(authService: authService),
        '/home': (_) => MainShell(
          feedService: feedService,
          controlService: controlService,
          dashboardService: dashboardService,
          settingsService: settingsService,
          tokenStorage: tokenStorage,
        ),
      },
      home: AuthGate(
        tokenStorage: tokenStorage,
        authService: authService,
        feedService: feedService,
        controlService: controlService,
        dashboardService: dashboardService,
        settingsService: settingsService,
      ),
    );
  }
}

class AuthGate extends StatelessWidget {
  const AuthGate({
    super.key,
    required this.tokenStorage,
    required this.authService,
    required this.feedService,
    required this.controlService,
    required this.dashboardService,
    required this.settingsService,
  });

  final TokenStorage tokenStorage;
  final AuthService authService;
  final FeedService feedService;
  final ControlService controlService;
  final DashboardService dashboardService;
  final SettingsService settingsService;

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<String?>(
      future: tokenStorage.readToken(),
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Scaffold(
            body: Center(child: CircularProgressIndicator()),
          );
        }

        if (snapshot.data == null || snapshot.data!.isEmpty) {
          return LoginScreen(authService: authService);
        }

        return MainShell(
          feedService: feedService,
          controlService: controlService,
          dashboardService: dashboardService,
          settingsService: settingsService,
          tokenStorage: tokenStorage,
        );
      },
    );
  }
}
