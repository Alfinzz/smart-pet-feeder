import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';

import '../models/dashboard_models.dart';
import '../models/feed_history_item.dart';
import '../models/manual_command.dart';
import '../services/control_service.dart';
import '../services/dashboard_service.dart';
import '../services/feed_service.dart';
import '../services/token_storage.dart';
import '../widgets/feed_history_chart.dart';

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({
    super.key,
    required this.feedService,
    required this.controlService,
    required this.dashboardService,
    required this.tokenStorage,
  });

  final FeedService feedService;
  final ControlService controlService;
  final DashboardService dashboardService;
  final TokenStorage tokenStorage;

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  bool _loading = true;
  String? _error;
  List<FeedHistoryItem> _history = [];
  DashboardOverview? _overview;
  String? _busyAction;

  @override
  void initState() {
    super.initState();
    _loadDashboard();
  }

  Future<void> _loadDashboard() async {
    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final results = await Future.wait<dynamic>([
        widget.dashboardService.fetchOverview(),
        widget.feedService.fetchHistory(limit: 50),
      ]);
      if (!mounted) return;
      setState(() {
        _overview = results[0] as DashboardOverview;
        _history = results[1] as List<FeedHistoryItem>;
        _loading = false;
      });
    } on DioException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = _messageFromDio(error);
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _error = 'Gagal memuat riwayat pakan.';
        _loading = false;
      });
    }
  }

  Future<void> _sendCommand(String action) async {
    if (_busyAction != null) return;

    setState(() => _busyAction = action);
    try {
      final command = await widget.controlService.sendManualCommand(
        action: action,
        deviceId: _overview?.device.id,
      );
      if (!mounted) return;
      _showCommandSnackBar(command, isPending: true);
      final finalCommand = await _waitForCommandResult(command);
      if (!mounted) return;
      _showCommandSnackBar(finalCommand);
      await _loadDashboard();
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: const Text('Command failed. Check connection.'),
          backgroundColor: Colors.red.shade400,
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(10),
          ),
        ),
      );
    } finally {
      if (mounted) setState(() => _busyAction = null);
    }
  }

  Future<ManualCommand> _waitForCommandResult(ManualCommand command) async {
    var current = command;
    final deadline = DateTime.now().add(const Duration(seconds: 60));

    while (!current.isTerminal && DateTime.now().isBefore(deadline)) {
      await Future<void>.delayed(const Duration(seconds: 2));
      current = await widget.controlService.fetchManualCommand(command.id);
    }

    return current;
  }

  void _showCommandSnackBar(ManualCommand command, {bool isPending = false}) {
    final isFeed = command.action == 'feed';
    final messenger = ScaffoldMessenger.of(context);
    messenger.hideCurrentSnackBar();

    var message = isFeed
        ? 'Feeding command is waiting for the feeder.'
        : 'Water refill command is waiting for the feeder.';
    var color = const Color(0xFF1565C0);

    if (!isPending) {
      if (command.status == 'completed') {
        message = isFeed ? 'Feeding completed.' : 'Water refill completed.';
        color = Colors.green.shade600;
      } else if (command.status == 'failed') {
        message = command.lastError.isEmpty
            ? 'Command failed on the feeder.'
            : 'Command failed: ${command.lastError}';
        color = Colors.red.shade400;
      } else {
        message = 'Still waiting for feeder confirmation.';
        color = Colors.orange.shade700;
      }
    }

    messenger.showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: color,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      ),
    );
  }

  String _messageFromDio(DioException error) {
    if (error.response?.statusCode == 401) {
      return 'Sesi berakhir. Silakan login ulang.';
    }
    return 'Server tidak dapat dihubungi.';
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      bottom: false,
      child: RefreshIndicator(
        onRefresh: _loadDashboard,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
          children: [
            // Greeting header
            _buildGreetingHeader(),
            const SizedBox(height: 20),
            // Status cards row
            _buildStatusCards(),
            const SizedBox(height: 20),
            // Weekly Nutrition chart
            _buildNutritionCard(),
            const SizedBox(height: 24),
            // Action buttons
            _buildActionButtons(),
          ],
        ),
      ),
    );
  }

  Widget _buildGreetingHeader() {
    final title = _overview?.greetingTitle ?? 'Hello, Fluffy!';
    final subtitle = _overview?.greetingSubtitle ?? 'Ready for breakfast?';

    return Row(
      children: [
        // Pet avatar
        Container(
          width: 50,
          height: 50,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: const Color(0xFFE3F2FD),
            border: Border.all(color: Colors.white, width: 2),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.08),
                blurRadius: 8,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: const Icon(Icons.pets, size: 24, color: Color(0xFF1565C0)),
        ),
        const SizedBox(width: 14),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF1E293B),
                ),
              ),
              Text(
                subtitle,
                style: TextStyle(fontSize: 14, color: Colors.grey.shade500),
              ),
            ],
          ),
        ),
        // Notification bell
        Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.06),
                blurRadius: 8,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: const Icon(
            Icons.notifications_outlined,
            size: 22,
            color: Color(0xFF64748B),
          ),
        ),
      ],
    );
  }

  Widget _buildStatusCards() {
    final device = _overview?.device;
    final stockPercent = device?.foodStockPercent ?? 0;
    final stockLabel = device?.foodStockLabel ?? 'Loading';
    final waterAvailable = device?.waterAvailable ?? false;
    final waterStatus = device?.waterStatus ?? 'Loading';

    return Row(
      children: [
        // Food Stock card
        Expanded(
          child: Container(
            padding: const EdgeInsets.all(16),
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
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Food Stock',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w500,
                        color: Colors.grey.shade600,
                      ),
                    ),
                    const Icon(
                      Icons.restaurant_menu,
                      size: 18,
                      color: Color(0xFF1565C0),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                Text(
                  '${stockPercent.round()}%',
                  style: TextStyle(
                    fontSize: 28,
                    fontWeight: FontWeight.w800,
                    color: Color(0xFF1565C0),
                  ),
                ),
                Text(
                  stockLabel,
                  style: TextStyle(fontSize: 13, color: Colors.grey.shade500),
                ),
                const SizedBox(height: 10),
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: device?.foodStockFraction ?? 0,
                    minHeight: 6,
                    backgroundColor: Colors.grey.shade200,
                    valueColor: const AlwaysStoppedAnimation<Color>(
                      Color(0xFF1565C0),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(width: 14),
        // Water card
        Expanded(
          child: Container(
            padding: const EdgeInsets.all(16),
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
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Water',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w500,
                        color: Colors.grey.shade600,
                      ),
                    ),
                    const Icon(
                      Icons.water_drop,
                      size: 18,
                      color: Color(0xFF0EA5E9),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                Text(
                  waterAvailable ? 'Available' : 'Unavailable',
                  style: TextStyle(
                    fontSize: 22,
                    fontWeight: FontWeight.w800,
                    color: waterAvailable
                        ? const Color(0xFF0EA5E9)
                        : const Color(0xFFEF4444),
                  ),
                ),
                const SizedBox(height: 6),
                Row(
                  children: [
                    Icon(
                      waterAvailable ? Icons.check_circle : Icons.error,
                      size: 14,
                      color: waterAvailable
                          ? Colors.green.shade500
                          : Colors.red.shade500,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      waterStatus,
                      style: TextStyle(
                        fontSize: 12,
                        color: waterAvailable
                            ? Colors.green.shade600
                            : Colors.red.shade500,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 14),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildNutritionCard() {
    return Container(
      padding: const EdgeInsets.all(20),
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
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'Weekly Nutrition',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF1E293B),
                ),
              ),
              GestureDetector(
                onTap: () {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Detail nutrisi tersedia di tab History.'),
                    ),
                  );
                },
                child: Row(
                  children: [
                    Text(
                      'Details',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w500,
                        color: Colors.grey.shade500,
                      ),
                    ),
                    Icon(
                      Icons.chevron_right,
                      size: 18,
                      color: Colors.grey.shade500,
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          SizedBox(height: 180, child: _buildChartArea()),
        ],
      ),
    );
  }

  Widget _buildChartArea() {
    if (_loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null) {
      return Center(child: Text(_error!, textAlign: TextAlign.center));
    }
    return FeedHistoryChart(history: _history);
  }

  Widget _buildActionButtons() {
    return Row(
      children: [
        Expanded(
          child: _ActionButton(
            label: 'Feed Now',
            icon: Icons.restaurant,
            color: const Color(0xFF0F4C75),
            isLoading: _busyAction == 'feed',
            onTap: () => _sendCommand('feed'),
          ),
        ),
        const SizedBox(width: 14),
        Expanded(
          child: _ActionButton(
            label: 'Refill Water',
            icon: Icons.water_drop,
            color: const Color(0xFF0EA5E9),
            isLoading: _busyAction == 'drink',
            onTap: () => _sendCommand('drink'),
          ),
        ),
      ],
    );
  }
}

class _ActionButton extends StatelessWidget {
  const _ActionButton({
    required this.label,
    required this.icon,
    required this.color,
    required this.isLoading,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final Color color;
  final bool isLoading;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: color,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: isLoading ? null : onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 16),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              if (isLoading)
                const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    color: Colors.white,
                  ),
                )
              else ...[
                Icon(icon, size: 20, color: Colors.white),
                const SizedBox(width: 8),
                Text(
                  label,
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: Colors.white,
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
