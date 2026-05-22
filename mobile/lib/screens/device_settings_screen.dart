import 'package:flutter/material.dart';
import '../models/dashboard_models.dart';
import '../services/settings_service.dart';

class DeviceSettingsScreen extends StatefulWidget {
  const DeviceSettingsScreen({
    super.key,
    required this.settingsService,
    this.initialDevice,
  });

  final SettingsService settingsService;
  final DeviceStatus? initialDevice;

  @override
  State<DeviceSettingsScreen> createState() => _DeviceSettingsScreenState();
}

class _DeviceSettingsScreenState extends State<DeviceSettingsScreen> {
  final _formKey = GlobalKey<FormState>();

  String _deviceName = 'Kitchen Feeder';
  double _portionSliderValue = 1; // 0: 1/8c, 1: 1/4c, 2: 1/2c, 3: 1c
  double _servoOpenValue = 25;
  double _servoClosedValue = 55;
  bool _automationEnabled = false;

  bool _isSaving = false;
  bool _isCalibrating = false;
  bool _isTestingServo = false;

  final List<String> _portionLabels = ['1/8c', '1/4c', '1/2c', '1c'];

  @override
  void initState() {
    super.initState();
    final device = widget.initialDevice;
    if (device != null && device.name.isNotEmpty) {
      _deviceName = device.name;
    }
    if (device != null) {
      _servoOpenValue = device.servoOpenDegrees.toDouble();
      _servoClosedValue = device.servoClosedDegrees.toDouble();
      _automationEnabled = device.automationEnabled;
      final grams = device.manualFeedPortionGrams;
      if (grams <= 15) {
        _portionSliderValue = 0;
      } else if (grams <= 30) {
        _portionSliderValue = 1;
      } else if (grams <= 60) {
        _portionSliderValue = 2;
      } else {
        _portionSliderValue = 3;
      }
    }
    _loadDeviceSettings();
  }

  Future<void> _loadDeviceSettings() async {
    try {
      final settings = await widget.settingsService.fetchDeviceSettings();
      if (!mounted) return;
      setState(() {
        final name = settings['name'] as String? ?? '';
        if (name.isNotEmpty) _deviceName = name;
        _servoOpenValue =
            (settings['servo_open_degrees'] as num?)?.toDouble() ??
            _servoOpenValue;
        _servoClosedValue =
            (settings['servo_closed_degrees'] as num?)?.toDouble() ??
            _servoClosedValue;
        _automationEnabled =
            settings['automation_enabled'] as bool? ?? _automationEnabled;
        final grams =
            (settings['manual_feed_portion_grams'] as num?)?.toDouble() ?? 30;
        if (grams <= 15) {
          _portionSliderValue = 0;
        } else if (grams <= 30) {
          _portionSliderValue = 1;
        } else if (grams <= 60) {
          _portionSliderValue = 2;
        } else {
          _portionSliderValue = 3;
        }
      });
    } catch (_) {
      // Keep initial dashboard values if the settings endpoint is unavailable.
    }
  }

  Future<void> _saveSettings() async {
    if (!_formKey.currentState!.validate()) return;
    _formKey.currentState!.save();
    if (!_validateServoRange()) return;

    setState(() => _isSaving = true);

    try {
      await widget.settingsService.updateDeviceSettings({
        'deviceName': _deviceName,
        'manualPortionSize': _portionLabels[_portionSliderValue.toInt()],
        'servo_open_degrees': _servoOpenValue.round(),
        'servo_closed_degrees': _servoClosedValue.round(),
        'automation_enabled': _automationEnabled,
      });

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Settings saved successfully!'),
          backgroundColor: Colors.green,
        ),
      );
      Navigator.pop(context);
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to save settings: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      if (mounted) {
        setState(() => _isSaving = false);
      }
    }
  }

  Future<void> _testServo() async {
    if (_isTestingServo) return;
    if (!_validateServoRange()) return;
    setState(() => _isTestingServo = true);

    try {
      await widget.settingsService.updateDeviceSettings({
        'deviceName': _deviceName,
        'manualPortionSize': _portionLabels[_portionSliderValue.toInt()],
        'servo_open_degrees': _servoOpenValue.round(),
        'servo_closed_degrees': _servoClosedValue.round(),
        'automation_enabled': _automationEnabled,
      });
      await widget.settingsService.testServo(
        deviceId: widget.initialDevice?.id,
      );

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Servo test command sent.'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to test servo: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      if (mounted) {
        setState(() => _isTestingServo = false);
      }
    }
  }

  bool _validateServoRange() {
    if (_servoOpenValue.round() == _servoClosedValue.round()) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Servo open and closed angles must be different.'),
          backgroundColor: Colors.red,
        ),
      );
      return false;
    }
    return true;
  }

  Future<void> _calibrate() async {
    setState(() => _isCalibrating = true);

    try {
      await widget.settingsService.calibrateSensor();

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Sensor calibrated successfully!'),
          backgroundColor: Colors.green,
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to calibrate: ${e.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      if (mounted) {
        setState(() => _isCalibrating = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final device = widget.initialDevice;
    final isRecentlySeen = device == null
        ? false
        : DateTime.now().difference(device.lastSeenAt).inMinutes < 10;
    final statusText = isRecentlySeen
        ? 'Online & Connected'
        : 'Waiting for device';
    final statusColor = isRecentlySeen ? Colors.green : Colors.orange;

    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        iconTheme: IconThemeData(color: Colors.blue[800]),
        title: Text(
          'Device Settings',
          style: TextStyle(
            color: Colors.blue[800],
            fontWeight: FontWeight.bold,
          ),
        ),
        centerTitle: true,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24.0),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Status Card
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.blue[50]?.withValues(alpha: 0.5),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: Colors.blue[100]!),
                ),
                child: Column(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Icon(
                        Icons.pets,
                        size: 40,
                        color: Colors.blue[600],
                      ),
                    ),
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Container(
                          width: 8,
                          height: 8,
                          decoration: BoxDecoration(
                            color: statusColor,
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Text(
                          statusText,
                          style: TextStyle(
                            color: statusColor,
                            fontWeight: FontWeight.w600,
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Text(
                      _deviceName,
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 18,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Food stock ${(device?.foodStockPercent ?? 0).round()}% - ${device?.waterStatus ?? 'Water status unknown'}',
                      style: TextStyle(color: Colors.grey[600], fontSize: 13),
                    ),
                    const SizedBox(height: 16),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        _buildBadge(
                          Icons.wifi,
                          isRecentlySeen ? 'Connected' : 'No recent signal',
                        ),
                        const SizedBox(width: 8),
                        _buildBadge(
                          Icons.water_drop,
                          device?.waterAvailable == true
                              ? 'Water available'
                              : 'Check water',
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),

              // Device Name Input
              TextFormField(
                initialValue: _deviceName,
                decoration: InputDecoration(
                  labelText: 'Device Name',
                  prefixIcon: Icon(
                    Icons.edit,
                    color: Colors.blue[600],
                    size: 20,
                  ),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                    borderSide: BorderSide(color: Colors.grey[300]!),
                  ),
                  filled: true,
                  fillColor: Colors.white,
                  helperText:
                      'This name will appear in notifications and dashboard.',
                ),
                validator: (val) =>
                    val == null || val.isEmpty ? 'Required' : null,
                onSaved: (val) => _deviceName = val ?? '',
              ),
              const SizedBox(height: 24),

              // Portion Slider
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.grey[200]!),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.02),
                      blurRadius: 8,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Manual Feed Portion Size',
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                          ),
                        ),
                        Container(
                          padding: const EdgeInsets.all(6),
                          decoration: BoxDecoration(
                            color: Colors.blue[50],
                            shape: BoxShape.circle,
                          ),
                          child: Icon(
                            Icons.restaurant,
                            color: Colors.blue[600],
                            size: 16,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Amount dispensed when "Feed Now" is pressed.',
                      style: TextStyle(color: Colors.grey[600], fontSize: 12),
                    ),
                    const SizedBox(height: 16),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          _portionLabels[_portionSliderValue.toInt()],
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 24,
                            color: Colors.blue[700],
                          ),
                        ),
                        const SizedBox(width: 8),
                        const Padding(
                          padding: EdgeInsets.only(bottom: 4.0),
                          child: Text(
                            'Cup',
                            style: TextStyle(fontWeight: FontWeight.w600),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Slider(
                      value: _portionSliderValue,
                      min: 0,
                      max: 3,
                      divisions: 3,
                      activeColor: Colors.blue[600],
                      inactiveColor: Colors.blue[100],
                      onChanged: (val) {
                        setState(() => _portionSliderValue = val);
                      },
                    ),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: _portionLabels
                          .map(
                            (e) => Text(
                              e,
                              style: TextStyle(
                                color: Colors.grey[500],
                                fontSize: 12,
                              ),
                            ),
                          )
                          .toList(),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),

              // Servo Calibration
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.grey[200]!),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Feed Gate Servo',
                          style: TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                          ),
                        ),
                        Icon(Icons.tune, color: Colors.blue[600], size: 20),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Use a smaller movement range if the physical gate has limited space.',
                      style: TextStyle(color: Colors.grey[600], fontSize: 12),
                    ),
                    const SizedBox(height: 18),
                    _buildServoSlider(
                      label: 'Open angle',
                      value: _servoOpenValue,
                      onChanged: (value) => setState(() {
                        _servoOpenValue = value;
                      }),
                    ),
                    const SizedBox(height: 12),
                    _buildServoSlider(
                      label: 'Closed angle',
                      value: _servoClosedValue,
                      onChanged: (value) => setState(() {
                        _servoClosedValue = value;
                      }),
                    ),
                    const SizedBox(height: 12),
                    SwitchListTile(
                      value: _automationEnabled,
                      contentPadding: EdgeInsets.zero,
                      activeColor: Colors.blue[700],
                      title: const Text(
                        'Automatic feeding and water refill',
                        style: TextStyle(
                          fontWeight: FontWeight.w700,
                          fontSize: 14,
                        ),
                      ),
                      subtitle: Text(
                        'Keep this off until the real sensors are calibrated.',
                        style: TextStyle(color: Colors.grey[600], fontSize: 12),
                      ),
                      onChanged: (value) {
                        setState(() => _automationEnabled = value);
                      },
                    ),
                    const SizedBox(height: 12),
                    SizedBox(
                      width: double.infinity,
                      height: 48,
                      child: OutlinedButton.icon(
                        onPressed: _isTestingServo ? null : _testServo,
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.blue[700],
                          side: BorderSide(color: Colors.blue[300]!),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                        icon: _isTestingServo
                            ? const SizedBox(
                                height: 16,
                                width: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.play_arrow, size: 20),
                        label: const Text(
                          'Save & Test Servo',
                          style: TextStyle(fontWeight: FontWeight.w600),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),

              // Calibrate Sensor
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: Colors.grey[200]!),
                ),
                child: Column(
                  children: [
                    const Text(
                      'Scale Calibration',
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 14,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Reset the feeder\'s internal scale to zero. Ensure the bowl is completely empty before taring.',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.grey[600], fontSize: 13),
                    ),
                    const SizedBox(height: 16),
                    SizedBox(
                      width: double.infinity,
                      height: 48,
                      child: OutlinedButton.icon(
                        onPressed: _isCalibrating ? null : _calibrate,
                        style: OutlinedButton.styleFrom(
                          foregroundColor: Colors.blue[700],
                          side: BorderSide(color: Colors.blue[300]!),
                          backgroundColor: Colors.blue[50]?.withValues(
                            alpha: 0.5,
                          ),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                        icon: _isCalibrating
                            ? const SizedBox(
                                height: 16,
                                width: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.scale, size: 20),
                        label: const Text(
                          'Calibrate Sensor (Tare)',
                          style: TextStyle(fontWeight: FontWeight.w600),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 32),

              // Save Settings
              SizedBox(
                height: 56,
                child: ElevatedButton.icon(
                  onPressed: _isSaving ? null : _saveSettings,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blue[700],
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: 0,
                  ),
                  icon: _isSaving
                      ? const SizedBox.shrink()
                      : const Icon(Icons.save, size: 20),
                  label: _isSaving
                      ? const SizedBox(
                          height: 24,
                          width: 24,
                          child: CircularProgressIndicator(
                            color: Colors.white,
                            strokeWidth: 2,
                          ),
                        )
                      : const Text(
                          'Save Settings',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildBadge(IconData icon, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.blue[100]?.withValues(alpha: 0.5),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: Colors.blue[800]),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              color: Colors.blue[900],
              fontSize: 12,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildServoSlider({
    required String label,
    required double value,
    required ValueChanged<double> onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              label,
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
            ),
            Text(
              '${value.round()} deg',
              style: TextStyle(
                color: Colors.blue[700],
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ),
        Slider(
          value: value,
          min: 0,
          max: 180,
          divisions: 180,
          activeColor: Colors.blue[600],
          inactiveColor: Colors.blue[100],
          onChanged: onChanged,
        ),
      ],
    );
  }
}
