import 'dart:math' as math;

import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

import '../models/feed_history_item.dart';

class FeedHistoryChart extends StatelessWidget {
  const FeedHistoryChart({super.key, required this.history});

  final List<FeedHistoryItem> history;

  @override
  Widget build(BuildContext context) {
    if (history.isEmpty) {
      return const Center(child: Text('Belum ada data riwayat pakan.'));
    }

    // Group by day of week and sum
    final now = DateTime.now();
    final weekday = now.weekday; // 1=Mon ... 7=Sun
    final dayLabels = ['M', 'T', 'W', 'T', 'F', 'S', 'S'];

    // Compute daily totals for the last 7 days
    final dailyTotals = List<double>.filled(7, 0);
    for (final item in history) {
      final diff = now.difference(item.recordedAt).inDays;
      if (diff >= 0 && diff < 7) {
        final dayIndex = (item.recordedAt.weekday - 1) % 7;
        dailyTotals[dayIndex] += item.weightGrams;
      }
    }

    final maxWeight = dailyTotals.fold<double>(
      0,
      (current, value) => math.max(current, value),
    );
    final maxY = math.max(100.0, maxWeight * 1.3);
    final todayIndex = (weekday - 1) % 7;

    return BarChart(
      BarChartData(
        maxY: maxY,
        barTouchData: BarTouchData(
          touchTooltipData: BarTouchTooltipData(
            getTooltipItem: (group, groupIndex, rod, rodIndex) {
              return BarTooltipItem(
                '${rod.toY.toStringAsFixed(0)}g',
                const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w600,
                  fontSize: 12,
                ),
              );
            },
          ),
        ),
        gridData: FlGridData(
          show: true,
          drawVerticalLine: false,
          horizontalInterval: maxY / 4,
          getDrawingHorizontalLine: (value) {
            return FlLine(
              color: Colors.grey.shade200,
              strokeWidth: 1,
            );
          },
        ),
        borderData: FlBorderData(show: false),
        titlesData: FlTitlesData(
          topTitles: const AxisTitles(
            sideTitles: SideTitles(showTitles: false),
          ),
          rightTitles: const AxisTitles(
            sideTitles: SideTitles(showTitles: false),
          ),
          leftTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 40,
              interval: maxY / 4,
              getTitlesWidget: (value, meta) {
                return Text(
                  '${value.toInt()}g',
                  style: TextStyle(
                    fontSize: 10,
                    color: Colors.grey.shade500,
                  ),
                );
              },
            ),
          ),
          bottomTitles: AxisTitles(
            sideTitles: SideTitles(
              showTitles: true,
              reservedSize: 28,
              getTitlesWidget: (value, meta) {
                final index = value.toInt();
                if (index < 0 || index >= 7) {
                  return const SizedBox.shrink();
                }
                final isToday = index == todayIndex;
                return Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: Text(
                    dayLabels[index],
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight:
                          isToday ? FontWeight.w700 : FontWeight.w500,
                      color: isToday
                          ? const Color(0xFF1E293B)
                          : Colors.grey.shade500,
                    ),
                  ),
                );
              },
            ),
          ),
        ),
        barGroups: List.generate(7, (index) {
          final isToday = index == todayIndex;
          return BarChartGroupData(
            x: index,
            barRods: [
              BarChartRodData(
                toY: dailyTotals[index],
                width: 20,
                borderRadius: const BorderRadius.only(
                  topLeft: Radius.circular(6),
                  topRight: Radius.circular(6),
                ),
                color: isToday
                    ? const Color(0xFF5B7FFF)
                    : const Color(0xFFE2E8F0),
              ),
            ],
            showingTooltipIndicators: isToday ? [0] : [],
          );
        }),
      ),
    );
  }
}
