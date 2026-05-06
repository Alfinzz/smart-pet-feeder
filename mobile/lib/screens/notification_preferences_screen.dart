import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import '../services/settings_service.dart';

class NotificationPreferencesScreen extends StatefulWidget {
  const NotificationPreferencesScreen({super.key, required this.settingsService});

  final SettingsService settingsService;

  @override
  State<NotificationPreferencesScreen> createState() => _NotificationPreferencesScreenState();
}

class _NotificationPreferencesScreenState extends State<NotificationPreferencesScreen> {
  bool _lowFoodAlert = true;
  bool _emptyWaterAlert = true;
  bool _feedSuccessReport = false;
  bool _healthAnomalies = true;
  
  bool _isLoading = false;

  Future<void> _updatePreference(String key, bool value) async {
    setState(() => _isLoading = true);
    
    // Save old state to revert if needed
    final oldState = _getState(key);
    _updateState(key, value);
    
    try {
      await widget.settingsService.updateNotificationPreferences({
        'lowFoodAlert': _lowFoodAlert,
        'emptyWaterAlert': _emptyWaterAlert,
        'feedSuccessReport': _feedSuccessReport,
        'healthAnomalies': _healthAnomalies,
      });
      // Optionally show a success snackbar
    } catch (e) {
      if (!mounted) return;
      // Revert state
      _updateState(key, oldState);
      
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to update preference: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }
  
  bool _getState(String key) {
    switch (key) {
      case 'lowFood': return _lowFoodAlert;
      case 'emptyWater': return _emptyWaterAlert;
      case 'feedSuccess': return _feedSuccessReport;
      case 'healthAnomalies': return _healthAnomalies;
      default: return false;
    }
  }
  
  void _updateState(String key, bool value) {
    setState(() {
      switch (key) {
        case 'lowFood': _lowFoodAlert = value; break;
        case 'emptyWater': _emptyWaterAlert = value; break;
        case 'feedSuccess': _feedSuccessReport = value; break;
        case 'healthAnomalies': _healthAnomalies = value; break;
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        iconTheme: IconThemeData(color: Colors.blue[800]),
        title: Text(
          'Notification Preferences',
          style: TextStyle(color: Colors.blue[800], fontWeight: FontWeight.bold),
        ),
        centerTitle: true,
      ),
      body: Stack(
        children: [
          ListView(
            padding: const EdgeInsets.all(24.0),
            children: [
              Text(
                'Alert Settings',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Colors.blue[900],
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Manage how Buddy communicates with you about your pet\'s feeding and health.',
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.grey[600],
                ),
              ),
              const SizedBox(height: 24),
              
              Container(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.05),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Column(
                  children: [
                    _buildPreferenceTile(
                      title: 'Low Food Alert',
                      description: 'Get notified when the hopper is below 10% capacity.',
                      icon: Icons.restaurant,
                      value: _lowFoodAlert,
                      onChanged: (val) => _updatePreference('lowFood', val),
                    ),
                    const Divider(height: 1, indent: 64, endIndent: 16),
                    _buildPreferenceTile(
                      title: 'Empty Water Alert',
                      description: 'Alerts if the water bowl requires immediate refilling.',
                      icon: Icons.water_drop,
                      value: _emptyWaterAlert,
                      onChanged: (val) => _updatePreference('emptyWater', val),
                    ),
                    const Divider(height: 1, indent: 64, endIndent: 16),
                    _buildPreferenceTile(
                      title: 'Feed Success Report',
                      description: 'Receive a summary after every scheduled feeding.',
                      icon: Icons.check_circle_outline,
                      value: _feedSuccessReport,
                      onChanged: (val) => _updatePreference('feedSuccess', val),
                    ),
                    const Divider(height: 1, indent: 64, endIndent: 16),
                    _buildPreferenceTile(
                      title: 'Health Anomalies',
                      description: 'Crucial alerts for irregular eating or drinking patterns.',
                      icon: Icons.monitor_heart,
                      value: _healthAnomalies,
                      onChanged: (val) => _updatePreference('healthAnomalies', val),
                      isLast: true,
                    ),
                  ],
                ),
              ),
            ],
          ),
          if (_isLoading)
            const Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: LinearProgressIndicator(),
            ),
        ],
      ),
    );
  }

  Widget _buildPreferenceTile({
    required String title,
    required String description,
    required IconData icon,
    required bool value,
    required ValueChanged<bool> onChanged,
    bool isLast = false,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8.0),
      child: ListTile(
        leading: Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: Colors.blue[50],
            shape: BoxShape.circle,
          ),
          child: Icon(icon, color: Colors.blue[600], size: 24),
        ),
        title: Text(
          title,
          style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 16),
        ),
        subtitle: Padding(
          padding: const EdgeInsets.only(top: 4.0),
          child: Text(
            description,
            style: TextStyle(color: Colors.grey[600], fontSize: 13),
          ),
        ),
        trailing: CupertinoSwitch(
          activeColor: Colors.blue[600],
          value: value,
          onChanged: _isLoading ? null : onChanged,
        ),
      ),
    );
  }
}
