import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';

import '../models/dashboard_models.dart';
import '../models/feed_history_item.dart';
import '../services/dashboard_service.dart';
import '../services/feed_service.dart';
import '../widgets/app_header.dart';

class HistoryScreen extends StatefulWidget {
  const HistoryScreen({
    super.key,
    required this.feedService,
    required this.dashboardService,
  });

  final FeedService feedService;
  final DashboardService dashboardService;

  @override
  State<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends State<HistoryScreen> {
  bool _loading = true;
  List<FeedHistoryItem> _history = [];
  WeeklyConsumption? _weekly;

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    setState(() => _loading = true);
    try {
      final results = await Future.wait<dynamic>([
        widget.feedService.fetchHistory(limit: 50),
        widget.dashboardService.fetchWeeklyConsumption(days: 7),
      ]);
      if (!mounted) return;
      setState(() {
        _history = results[0] as List<FeedHistoryItem>;
        _weekly = results[1] as WeeklyConsumption;
        _loading = false;
      });
    } on DioException {
      if (!mounted) return;
      setState(() => _loading = false);
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: _loadHistory,
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            const AppHeader(),
            const SizedBox(height: 8),
            // 7-Day Consumption card
            _buildConsumptionCard(),
            const SizedBox(height: 24),
            // Recent Logs
            _buildRecentLogs(),
          ],
        ),
      ),
    );
  }

  Widget _buildConsumptionCard() {
    final weekly = _weekly;
    final target = weekly?.dailyTargetGrams ?? 150;
    final average = weekly?.averageGrams ?? 0;
    final total = weekly?.totalGrams ?? 0;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Container(
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
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  '7-Day\nConsumption',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF1E293B),
                    height: 1.3,
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF0F9FF),
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Container(
                        width: 8,
                        height: 8,
                        decoration: const BoxDecoration(
                          color: Color(0xFF0EA5E9),
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 6),
                      const Text(
                        'Daily Target',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w500,
                          color: Color(0xFF0EA5E9),
                          height: 1.3,
                        ),
                      ),
                      Text(
                        '${target.toStringAsFixed(0)}g',
                        style: const TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF0EA5E9),
                          height: 1.3,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 20),
            // Bar chart
            SizedBox(height: 140, child: _buildBarChart()),
            const SizedBox(height: 16),
            // Summary row
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Weekly Average',
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey.shade500,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '${average.toStringAsFixed(0)}g',
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF0EA5E9),
                        ),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                        'Total Dispensed',
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey.shade500,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        '${total.toStringAsFixed(0)}g',
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF1E293B),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBarChart() {
    if (_loading && _weekly == null) {
      return const Center(child: CircularProgressIndicator());
    }

    final items = _weekly?.days ?? [];
    if (items.isEmpty) {
      return const Center(child: Text('Belum ada data konsumsi.'));
    }

    final target = _weekly?.dailyTargetGrams ?? 150;
    final maxValue = items.fold<double>(
      target,
      (current, item) => math.max(current, item.totalGrams),
    );
    final maxVal = math.max(100.0, maxValue * 1.25);
    final now = DateTime.now();
    final todayIndex = items.indexWhere((item) {
      return item.date.year == now.year &&
          item.date.month == now.month &&
          item.date.day == now.day;
    });
    final highlightIndex = todayIndex >= 0 ? todayIndex : items.length - 1;

    return Row(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: List.generate(items.length, (index) {
        final item = items[index];
        final isHighlighted = index == highlightIndex;
        final barHeight = (item.totalGrams / maxVal) * 120;

        return Expanded(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              if (isHighlighted)
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 6,
                    vertical: 3,
                  ),
                  margin: const EdgeInsets.only(bottom: 4),
                  decoration: BoxDecoration(
                    color: const Color(0xFF1E293B),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Text(
                    '${item.totalGrams.toStringAsFixed(0)}g',
                    style: const TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.w600,
                      color: Colors.white,
                    ),
                  ),
                ),
              Container(
                height: barHeight,
                margin: const EdgeInsets.symmetric(horizontal: 6),
                decoration: BoxDecoration(
                  color: isHighlighted
                      ? const Color(0xFF5B7FFF)
                      : const Color(0xFFE2E8F0),
                  borderRadius: BorderRadius.circular(6),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                item.dayLabel,
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: isHighlighted ? FontWeight.w700 : FontWeight.w500,
                  color: isHighlighted
                      ? const Color(0xFF1E293B)
                      : Colors.grey.shade500,
                ),
              ),
            ],
          ),
        );
      }),
    );
  }

  Widget _buildRecentLogs() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                'Recent Logs',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF1E293B),
                ),
              ),
              GestureDetector(
                onTap: () {},
                child: const Text(
                  'View All',
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                    color: Color(0xFF5B7FFF),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          if (_loading)
            const Padding(
              padding: EdgeInsets.all(20),
              child: CircularProgressIndicator(),
            )
          else if (_history.isEmpty)
            ..._buildDemoLogs()
          else
            ..._history
                .take(4)
                .map(
                  (item) => _buildLogItem(
                    icon: Icons.restaurant,
                    color: const Color(0xFF1565C0),
                    title: 'Feed',
                    subtitle: _formatDateTime(item.recordedAt),
                    amount: '${item.weightGrams.toStringAsFixed(0)} grams',
                  ),
                ),
        ],
      ),
    );
  }

  List<Widget> _buildDemoLogs() {
    return [
      _buildLogItem(
        icon: Icons.wb_sunny_outlined,
        color: const Color(0xFF1565C0),
        title: 'Morning Feed',
        subtitle: 'Today, 08:00 AM',
        amount: '50 grams',
      ),
      _buildLogItem(
        icon: Icons.nightlight_outlined,
        color: const Color(0xFF7C3AED),
        title: 'Evening Feed',
        subtitle: 'Yesterday, 06:30 PM',
        amount: '45 grams',
      ),
      _buildLogItem(
        icon: Icons.cookie_outlined,
        color: const Color(0xFF10B981),
        title: 'Manual Snack',
        subtitle: 'Yesterday, 02:15 PM',
        amount: '15 grams',
      ),
      _buildLogItem(
        icon: Icons.wb_sunny_outlined,
        color: const Color(0xFF1565C0),
        title: 'Morning Feed',
        subtitle: 'Yesterday, 08:00 AM',
        amount: '50 grams',
      ),
    ];
  }

  Widget _buildLogItem({
    required IconData icon,
    required Color color,
    required String title,
    required String subtitle,
    required String amount,
  }) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(14),
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
            width: 42,
            height: 42,
            decoration: BoxDecoration(
              color: color.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(icon, size: 20, color: color),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 14,
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
            amount,
            style: const TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: Color(0xFF1E293B),
            ),
          ),
        ],
      ),
    );
  }

  String _formatDateTime(DateTime time) {
    final now = DateTime.now();
    final isToday =
        time.year == now.year && time.month == now.month && time.day == now.day;
    final prefix = isToday ? 'Today' : 'Yesterday';
    final hour = time.hour.toString().padLeft(2, '0');
    final minute = time.minute.toString().padLeft(2, '0');
    return '$prefix, $hour:$minute';
  }
}
