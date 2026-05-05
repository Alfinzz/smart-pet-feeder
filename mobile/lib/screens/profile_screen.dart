import 'package:flutter/material.dart';

import '../models/dashboard_models.dart';
import '../services/dashboard_service.dart';
import '../services/token_storage.dart';
import '../widgets/app_header.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({
    super.key,
    required this.tokenStorage,
    required this.dashboardService,
  });

  final TokenStorage tokenStorage;
  final DashboardService dashboardService;

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  bool _loading = true;
  ProfileSummary? _profile;

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
        // Pet avatar
        Container(
          width: 100,
          height: 100,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: const Color(0xFFF1F5F9),
            border: Border.all(color: Colors.white, width: 4),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.08),
                blurRadius: 16,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: const Icon(Icons.pets, size: 44, color: Color(0xFF1565C0)),
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
            _MenuItem(
              icon: Icons.pets,
              iconColor: const Color(0xFF1565C0),
              iconBgColor: const Color(0xFFE3F2FD),
              title: 'Pet Details',
              subtitle:
                  '${pet?.species ?? 'Pet'}, ${(pet?.weightKg ?? 0).toStringAsFixed(1)} kg',
              onTap: () {},
            ),
            Divider(height: 1, color: Colors.grey.shade100, indent: 70),
            _MenuItem(
              icon: Icons.settings_outlined,
              iconColor: const Color(0xFF10B981),
              iconBgColor: const Color(0xFFF0FDF4),
              title: 'Device Settings',
              subtitle:
                  '${device?.name ?? 'Smart Feeder'}, ${(device?.foodStockPercent ?? 0).round()}% stock',
              onTap: () {},
            ),
            Divider(height: 1, color: Colors.grey.shade100, indent: 70),
            _MenuItem(
              icon: Icons.notifications_outlined,
              iconColor: const Color(0xFFF59E0B),
              iconBgColor: const Color(0xFFFFFBEB),
              title: 'Notification Preferences',
              subtitle: 'Alerts, Reminders',
              onTap: () {},
            ),
            Divider(height: 1, color: Colors.grey.shade100, indent: 70),
            _MenuItem(
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

class _MenuItem extends StatelessWidget {
  const _MenuItem({
    required this.icon,
    required this.iconColor,
    required this.iconBgColor,
    required this.title,
    this.subtitle,
    this.titleColor,
    this.showChevron = true,
    required this.onTap,
  });

  final IconData icon;
  final Color iconColor;
  final Color iconBgColor;
  final String title;
  final String? subtitle;
  final Color? titleColor;
  final bool showChevron;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
            Container(
              width: 42,
              height: 42,
              decoration: BoxDecoration(
                color: iconBgColor,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(icon, size: 22, color: iconColor),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                      color: titleColor ?? const Color(0xFF1E293B),
                    ),
                  ),
                  if (subtitle != null) ...[
                    const SizedBox(height: 2),
                    Text(
                      subtitle!,
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey.shade500,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            if (showChevron)
              Icon(Icons.chevron_right, size: 22, color: Colors.grey.shade400),
          ],
        ),
      ),
    );
  }
}
