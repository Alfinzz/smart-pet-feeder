import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import '../models/dashboard_models.dart';
import '../services/dashboard_service.dart';
import '../services/settings_service.dart';
import '../services/token_storage.dart';
import '../widgets/app_header.dart';
import '../widgets/log_vitals_bottom_sheet.dart';
import '../widgets/pet_photo_avatar.dart';
import '../widgets/profile_menu_item.dart';
import 'device_settings_screen.dart';
import 'notification_preferences_screen.dart';
import 'pet_details_screen.dart';
import 'wifi_setup_screen.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({
    super.key,
    required this.tokenStorage,
    required this.dashboardService,
    required this.settingsService,
  });

  final TokenStorage tokenStorage;
  final DashboardService dashboardService;
  final SettingsService settingsService;

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  bool _loading = true;
  bool _uploadingPhoto = false;
  ProfileSummary? _profile;
  final ImagePicker _imagePicker = ImagePicker();

  @override
  void initState() {
    super.initState();
    _loadProfile();
  }

  Future<void> _loadProfile() async {
    setState(() => _loading = true);
    try {
      final profile = await widget.dashboardService.fetchProfile();
      if (!mounted) return;
      setState(() {
        _profile = profile;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: _loadProfile,
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            const AppHeader(),
            const SizedBox(height: 20),
            if (_loading && _profile == null)
              const Padding(
                padding: EdgeInsets.all(40),
                child: Center(child: CircularProgressIndicator()),
              )
            else ...[
              _buildPetProfile(),
              const SizedBox(height: 32),
              _buildMenuSection(context),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildPetProfile() {
    final pet = _profile?.pet;
    final petName = pet?.name ?? 'Fluffy';
    final petDetail = '${pet?.ageYears ?? 0} Years Old - ${pet?.breed ?? '-'}';

    return Column(
      children: [
        PetPhotoAvatar(
          photoUrl: pet?.photoUrl ?? '',
          isUploading: _uploadingPhoto,
          showCameraBadge: true,
          onTap: _uploadingPhoto ? null : _pickAndUploadPetPhoto,
        ),
        const SizedBox(height: 16),
        Text(
          petName,
          style: TextStyle(
            fontSize: 24,
            fontWeight: FontWeight.w700,
            color: Color(0xFF1E293B),
          ),
        ),
        const SizedBox(height: 4),
        Text(
          petDetail,
          style: TextStyle(fontSize: 14, color: Colors.grey.shade500),
        ),
      ],
    );
  }

  Future<void> _pickAndUploadPetPhoto() async {
    final image = await _imagePicker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1600,
      imageQuality: 85,
    );
    if (image == null) return;
    if (!mounted) return;

    setState(() => _uploadingPhoto = true);
    try {
      await widget.settingsService.uploadPetPhoto(image.path);
      await _loadProfile();
      if (!mounted) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Pet photo updated')));
    } catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(error.toString())));
    } finally {
      if (mounted) {
        setState(() => _uploadingPhoto = false);
      }
    }
  }

  Widget _buildMenuSection(BuildContext context) {
    final pet = _profile?.pet;
    final device = _profile?.device;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.05),
              blurRadius: 10,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          children: [
            ProfileMenuItem(
              icon: Icons.pets,
              iconColor: const Color(0xFF1565C0),
              iconBgColor: const Color(0xFFE3F2FD),
              title: 'Pet Details',
              subtitle:
                  '${pet?.species ?? 'Pet'}, ${(pet?.weightKg ?? 0).toStringAsFixed(1)} kg',
              onTap: () async {
                await Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (context) => PetDetailsScreen(
                      settingsService: widget.settingsService,
                      initialPet: pet,
                    ),
                  ),
                );
                if (mounted) await _loadProfile();
              },
            ),
            Divider(height: 1, color: Colors.grey.shade100, indent: 70),
            ProfileMenuItem(
              icon: Icons.settings_outlined,
              iconColor: const Color(0xFF10B981),
              iconBgColor: const Color(0xFFF0FDF4),
              title: 'Device Settings',
              subtitle:
                  '${device?.name ?? 'Smart Feeder'}, ${(device?.foodStockPercent ?? 0).round()}% stock',
              onTap: () async {
                await Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (context) => DeviceSettingsScreen(
                      settingsService: widget.settingsService,
                      initialDevice: device,
                    ),
                  ),
                );
                if (mounted) await _loadProfile();
              },
            ),
            Divider(height: 1, color: Colors.grey.shade100, indent: 70),
            ProfileMenuItem(
              icon: Icons.wifi_tethering,
              iconColor: const Color(0xFF0EA5E9),
              iconBgColor: const Color(0xFFE0F2FE),
              title: 'WiFi Setup',
              subtitle: 'Connect feeder to a new network',
              onTap: () async {
                await Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (context) => WifiSetupScreen(
                      settingsService: widget.settingsService,
                    ),
                  ),
                );
                if (mounted) await _loadProfile();
              },
            ),
            Divider(height: 1, color: Colors.grey.shade100, indent: 70),
            ProfileMenuItem(
              icon: Icons.notifications_outlined,
              iconColor: const Color(0xFFF59E0B),
              iconBgColor: const Color(0xFFFFFBEB),
              title: 'Notification Preferences',
              subtitle: 'Alerts, Reminders',
              onTap: () async {
                await Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (context) => NotificationPreferencesScreen(
                      settingsService: widget.settingsService,
                    ),
                  ),
                );
                if (mounted) await _loadProfile();
              },
            ),
            Divider(height: 1, color: Colors.grey.shade100, indent: 70),
            ProfileMenuItem(
              icon: Icons.monitor_heart_outlined,
              iconColor: const Color(0xFF8B5CF6),
              iconBgColor: const Color(0xFFF5F3FF),
              title: 'Log Health Vitals',
              subtitle: 'Weight, activity, sleep',
              onTap: () {
                showModalBottomSheet(
                  context: context,
                  isScrollControlled: true,
                  backgroundColor: Colors.transparent,
                  builder: (context) => LogVitalsBottomSheet(
                    settingsService: widget.settingsService,
                  ),
                );
              },
            ),
            Divider(height: 1, color: Colors.grey.shade100, indent: 70),
            ProfileMenuItem(
              icon: Icons.logout,
              iconColor: const Color(0xFFEF4444),
              iconBgColor: const Color(0xFFFEF2F2),
              title: 'Logout',
              titleColor: const Color(0xFFEF4444),
              showChevron: false,
              onTap: () => _logout(context),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _logout(BuildContext context) async {
    await widget.tokenStorage.clear();
    if (!context.mounted) return;
    Navigator.of(context).pushNamedAndRemoveUntil('/login', (_) => false);
  }
}
