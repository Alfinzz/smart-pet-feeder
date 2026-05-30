import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../models/dashboard_models.dart';
import '../services/dashboard_service.dart';
import '../services/settings_service.dart';
import '../widgets/app_header.dart';
import '../widgets/log_vitals_bottom_sheet.dart';

class HealthScreen extends StatefulWidget {
  const HealthScreen({
    super.key,
    required this.dashboardService,
    required this.settingsService,
  });

  final DashboardService dashboardService;
  final SettingsService settingsService;

  @override
  State<HealthScreen> createState() => _HealthScreenState();
}

class _HealthScreenState extends State<HealthScreen> {
  bool _loading = true;
  HealthSummary? _summary;

  @override
  void initState() {
    super.initState();
    _loadHealth();
  }

  Future<void> _loadHealth() async {
    setState(() => _loading = true);
    try {
      final summary = await widget.dashboardService.fetchHealthSummary();
      if (!mounted) return;
      setState(() {
        _summary = summary;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
    }
  }

  Future<void> _showLogVitalsBottomSheet(BuildContext context) async {
    final submitted = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) =>
          LogVitalsBottomSheet(settingsService: widget.settingsService),
    );

    if (mounted && submitted == true) {
      await _loadHealth();
    }
  }

  Future<bool> _markTaskCompleted(CareTask task) async {
    try {
      await widget.dashboardService.markCareTaskCompleted(task.id);
      await _createNextTaskIfNeeded(task);
      if (!mounted) return true;
      final summary = _summary;
      if (summary != null) {
        setState(() {
          _summary = summary.copyWith(
            upcomingTasks: summary.upcomingTasks
                .where((item) => item.id != task.id)
                .toList(),
          );
        });
      }
      await _loadHealth();
      if (!mounted) return true;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('${task.title} marked completed.')),
      );
      return true;
    } catch (error) {
      if (!mounted) return false;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to update task: ${error.toString()}'),
          backgroundColor: Colors.red,
        ),
      );
      return false;
    }
  }

  Future<void> _createNextTaskIfNeeded(CareTask task) async {
    final dueDate = _nextDueDateForTask(task, _summary?.pet.ageInMonths ?? 0);
    if (dueDate == null) return;
    await widget.dashboardService.createCareTask(
      category: task.category,
      title: task.title,
      description: task.description.isEmpty ? task.title : task.description,
      dueDate: dueDate,
      priority: task.priority,
    );
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: _loadHealth,
        child: ListView(
          padding: const EdgeInsets.only(bottom: 36),
          children: [
            const AppHeader(),
            const SizedBox(height: 16),
            if (_loading && _summary == null)
              const Padding(
                padding: EdgeInsets.all(40),
                child: Center(child: CircularProgressIndicator()),
              )
            else if (_summary == null)
              _buildEmptyHealth()
            else ...[
              _buildHealthScore(context),
              const SizedBox(height: 28),
              _buildVitalSigns(context),
              const SizedBox(height: 28),
              _buildUpcomingTasks(context),
              const SizedBox(height: 24),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildHealthScore(BuildContext context) {
    final summary = _summary;
    final score = summary?.score ?? 0;
    final statusLabel = summary?.statusLabel ?? 'Unknown';
    final headline = summary?.headline ?? 'No Health Data';
    final description =
        summary?.description ?? 'Health metrics are not available yet.';

    return Column(
      children: [
        // Circular score
        SizedBox(
          width: 180,
          height: 180,
          child: CustomPaint(
            painter: _HealthScorePainter(score: score.toDouble()),
            child: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    '$score',
                    style: TextStyle(
                      fontSize: 48,
                      fontWeight: FontWeight.w800,
                      color: Color(0xFF1E293B),
                    ),
                  ),
                  Text(
                    'SCORE',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: Color(0xFF94A3B8),
                      letterSpacing: 2,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
        const SizedBox(height: 16),
        // Excellent badge
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
          decoration: BoxDecoration(
            color: const Color(0xFFF0FDF4),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.check_circle, size: 16, color: Colors.green.shade600),
              const SizedBox(width: 6),
              Text(
                statusLabel,
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: Colors.green.shade700,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 14),
        Text(
          headline,
          style: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.w700,
            color: Color(0xFF1E293B),
          ),
        ),
        const SizedBox(height: 8),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 40),
          child: Text(
            description,
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 13,
              color: Colors.grey.shade600,
              height: 1.5,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildVitalSigns(BuildContext context) {
    final vitals = _summary?.vitals;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'Vital Signs',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF1E293B),
                ),
              ),
              TextButton.icon(
                onPressed: () => _showLogVitalsBottomSheet(context),
                style: TextButton.styleFrom(
                  foregroundColor: const Color(0xFF5B7FFF),
                  padding: const EdgeInsets.symmetric(horizontal: 10),
                  minimumSize: const Size(0, 36),
                  tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                ),
                icon: const Icon(Icons.add, size: 18),
                label: const Text(
                  'Log',
                  style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          Row(
            children: [
              Expanded(
                child: _VitalCard(
                  icon: Icons.monitor_weight_outlined,
                  iconColor: const Color(0xFF1565C0),
                  iconBgColor: const Color(0xFFE3F2FD),
                  value: (vitals?.weightKg ?? 0).toStringAsFixed(1),
                  unit: 'kg Weight',
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _VitalCard(
                  icon: Icons.directions_run,
                  iconColor: const Color(0xFFF97316),
                  iconBgColor: const Color(0xFFFFF7ED),
                  value: '${vitals?.activityMinutes ?? 0}',
                  unit: 'mins\nActivity',
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _VitalCard(
                  icon: Icons.nightlight_round,
                  iconColor: const Color(0xFF6366F1),
                  iconBgColor: const Color(0xFFEEF2FF),
                  value: (vitals?.sleepHours ?? 0).toStringAsFixed(1),
                  unit: 'hrs Sleep',
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildUpcomingTasks(BuildContext context) {
    final tasks = _summary?.upcomingTasks ?? const <CareTask>[];

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'Upcoming Tasks',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF1E293B),
                ),
              ),
              Text(
                '${tasks.length} tasks',
                style: const TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF64748B),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          if (tasks.isEmpty)
            const Padding(
              padding: EdgeInsets.all(20),
              child: Text('No upcoming tasks.'),
            )
          else
            ...tasks.map((task) {
              final style = _taskStyle(task.category, task.priority);
              final taskItem = _buildTaskItem(
                icon: style.icon,
                iconColor: style.iconColor,
                iconBgColor: style.iconBgColor,
                title: task.title,
                subtitle: task.subtitle,
                trailing: task.displayDueLabel,
                trailingColor: task.isDueWithinSevenDays
                    ? const Color(0xFFEF4444)
                    : style.trailingColor,
              );
              return Padding(
                padding: const EdgeInsets.only(bottom: 10),
                child: task.id <= 0
                    ? taskItem
                    : Dismissible(
                        key: ValueKey('care-task-${task.id}'),
                        direction: DismissDirection.endToStart,
                        confirmDismiss: (_) => _markTaskCompleted(task),
                        background: Container(
                          alignment: Alignment.centerRight,
                          padding: const EdgeInsets.symmetric(horizontal: 20),
                          decoration: BoxDecoration(
                            color: const Color(0xFF10B981),
                            borderRadius: BorderRadius.circular(14),
                          ),
                          child: const Icon(
                            Icons.check_circle,
                            color: Colors.white,
                          ),
                        ),
                        child: taskItem,
                      ),
              );
            }),
        ],
      ),
    );
  }

  Widget _buildEmptyHealth() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 32, 20, 0),
      child: Container(
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFFE2E8F0)),
        ),
        child: Column(
          children: [
            Container(
              width: 52,
              height: 52,
              decoration: BoxDecoration(
                color: const Color(0xFFF0FDF4),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Icon(
                Icons.monitor_heart_outlined,
                color: Colors.green.shade700,
              ),
            ),
            const SizedBox(height: 14),
            const Text(
              'No health data yet',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w700,
                color: Color(0xFF1E293B),
              ),
            ),
            const SizedBox(height: 6),
            Text(
              'Log weight, activity, and sleep to start tracking health trends.',
              textAlign: TextAlign.center,
              style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
            ),
            const SizedBox(height: 18),
            SizedBox(
              width: double.infinity,
              height: 48,
              child: OutlinedButton.icon(
                onPressed: () => _showLogVitalsBottomSheet(context),
                icon: const Icon(Icons.add, size: 18),
                label: const Text('Log Vitals'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTaskItem({
    required IconData icon,
    required Color iconColor,
    required Color iconBgColor,
    required String title,
    required String subtitle,
    required String trailing,
    required Color trailingColor,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
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
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF1E293B),
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  subtitle,
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade500),
                ),
              ],
            ),
          ),
          Text(
            trailing,
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: trailingColor,
            ),
          ),
        ],
      ),
    );
  }
}

DateTime? _nextDueDateForTask(CareTask task, int ageInMonths) {
  final category = task.category.toLowerCase();
  final title = task.title.toLowerCase();
  if (category == 'vaccination' ||
      category == 'vaccine' ||
      title.contains('vaksin') ||
      title.contains('vaccin')) {
    return DateTime.now().add(Duration(days: ageInMonths < 4 ? 21 : 365));
  }
  if (category == 'checkup' ||
      title.contains('medical checkup') ||
      title.contains('vet checkup')) {
    return DateTime.now().add(const Duration(days: 180));
  }
  return null;
}

_TaskStyle _taskStyle(String category, String priority) {
  if (category == 'vaccination') {
    return const _TaskStyle(
      icon: Icons.vaccines,
      iconColor: Color(0xFFEF4444),
      iconBgColor: Color(0xFFFEF2F2),
      trailingColor: Color(0xFFEF4444),
    );
  }
  if (category == 'checkup') {
    return const _TaskStyle(
      icon: Icons.medical_services_outlined,
      iconColor: Color(0xFF0EA5E9),
      iconBgColor: Color(0xFFF0F9FF),
      trailingColor: Color(0xFF64748B),
    );
  }
  return _TaskStyle(
    icon: Icons.event_note_outlined,
    iconColor: priority == 'high'
        ? const Color(0xFFEF4444)
        : const Color(0xFF10B981),
    iconBgColor: priority == 'high'
        ? const Color(0xFFFEF2F2)
        : const Color(0xFFF0FDF4),
    trailingColor: priority == 'high'
        ? const Color(0xFFEF4444)
        : const Color(0xFF64748B),
  );
}

class _TaskStyle {
  const _TaskStyle({
    required this.icon,
    required this.iconColor,
    required this.iconBgColor,
    required this.trailingColor,
  });

  final IconData icon;
  final Color iconColor;
  final Color iconBgColor;
  final Color trailingColor;
}

class _VitalCard extends StatelessWidget {
  const _VitalCard({
    required this.icon,
    required this.iconColor,
    required this.iconBgColor,
    required this.value,
    required this.unit,
  });

  final IconData icon;
  final Color iconColor;
  final Color iconBgColor;
  final String value;
  final String unit;

  @override
  Widget build(BuildContext context) {
    return Container(
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
          const SizedBox(height: 10),
          Text(
            value,
            style: const TextStyle(
              fontSize: 22,
              fontWeight: FontWeight.w800,
              color: Color(0xFF1E293B),
            ),
          ),
          const SizedBox(height: 2),
          Text(
            unit,
            textAlign: TextAlign.center,
            style: TextStyle(fontSize: 11, color: Colors.grey.shade500),
          ),
        ],
      ),
    );
  }
}

class _HealthScorePainter extends CustomPainter {
  _HealthScorePainter({required this.score});

  final double score;

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = math.min(size.width, size.height) / 2 - 10;

    // Background arc
    final bgPaint = Paint()
      ..color = const Color(0xFFE2E8F0)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 12
      ..strokeCap = StrokeCap.round;

    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      -math.pi * 0.75,
      math.pi * 1.5,
      false,
      bgPaint,
    );

    // Score arc
    final scorePaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 12
      ..strokeCap = StrokeCap.round
      ..shader = const SweepGradient(
        startAngle: -math.pi * 0.75,
        endAngle: math.pi * 0.75,
        colors: [Color(0xFF10B981), Color(0xFF059669)],
      ).createShader(Rect.fromCircle(center: center, radius: radius));

    final sweepAngle = (score / 100) * math.pi * 1.5;
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      -math.pi * 0.75,
      sweepAngle,
      false,
      scorePaint,
    );
  }

  @override
  bool shouldRepaint(covariant _HealthScorePainter oldDelegate) {
    return oldDelegate.score != score;
  }
}
