import 'owner.dart';
import '../config/app_config.dart';

class PetProfile {
  const PetProfile({
    required this.id,
    required this.ownerId,
    required this.deviceId,
    required this.name,
    required this.species,
    required this.breed,
    required this.ageYears,
    required this.ageInMonths,
    required this.weightKg,
    required this.dailyFeedTargetGrams,
    required this.healthScore,
    required this.healthStatus,
    required this.healthHeadline,
    required this.healthDescription,
    required this.activityMinutes,
    required this.sleepHours,
    required this.photoUrl,
  });

  final int id;
  final int ownerId;
  final String deviceId;
  final String name;
  final String species;
  final String breed;
  final int ageYears;
  final int ageInMonths;
  final double weightKg;
  final double dailyFeedTargetGrams;
  final int healthScore;
  final String healthStatus;
  final String healthHeadline;
  final String healthDescription;
  final int activityMinutes;
  final double sleepHours;
  final String photoUrl;

  factory PetProfile.fromJson(Map<String, dynamic> json) {
    final ageYears = (json['age_years'] as num?)?.toInt() ?? 0;
    return PetProfile(
      id: (json['id'] as num?)?.toInt() ?? 0,
      ownerId: (json['owner_id'] as num?)?.toInt() ?? 0,
      deviceId: json['device_id'] as String? ?? '',
      name: json['name'] as String? ?? 'Fluffy',
      species: json['species'] as String? ?? 'Dog',
      breed: json['breed'] as String? ?? 'Golden Retriever',
      ageYears: ageYears,
      ageInMonths: (json['age_in_months'] as num?)?.toInt() ?? ageYears * 12,
      weightKg: (json['weight_kg'] as num?)?.toDouble() ?? 0,
      dailyFeedTargetGrams:
          (json['daily_feed_target_grams'] as num?)?.toDouble() ?? 150,
      healthScore: (json['health_score'] as num?)?.toInt() ?? 0,
      healthStatus: json['health_status'] as String? ?? '',
      healthHeadline: json['health_headline'] as String? ?? '',
      healthDescription: json['health_description'] as String? ?? '',
      activityMinutes: (json['activity_minutes'] as num?)?.toInt() ?? 0,
      sleepHours: (json['sleep_hours'] as num?)?.toDouble() ?? 0,
      photoUrl: AppConfig.publicUrl(json['photo_url'] as String? ?? ''),
    );
  }
}

class DeviceStatus {
  const DeviceStatus({
    required this.id,
    required this.name,
    required this.foodStockPercent,
    required this.foodStockLabel,
    required this.waterAvailable,
    required this.waterStatus,
    required this.manualFeedPortionGrams,
    required this.servoOpenDegrees,
    required this.servoClosedDegrees,
    required this.automationEnabled,
    required this.lastSeenAt,
  });

  final String id;
  final String name;
  final double foodStockPercent;
  final String foodStockLabel;
  final bool waterAvailable;
  final String waterStatus;
  final double manualFeedPortionGrams;
  final int servoOpenDegrees;
  final int servoClosedDegrees;
  final bool automationEnabled;
  final DateTime lastSeenAt;

  double get foodStockFraction =>
      ((foodStockPercent / 100).clamp(0.0, 1.0) as num).toDouble();

  factory DeviceStatus.fromJson(Map<String, dynamic> json) {
    return DeviceStatus(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? 'Smart Feeder',
      foodStockPercent: (json['food_stock_percent'] as num?)?.toDouble() ?? 0,
      foodStockLabel: json['food_stock_label'] as String? ?? '',
      waterAvailable: json['water_available'] as bool? ?? false,
      waterStatus: json['water_status'] as String? ?? '',
      manualFeedPortionGrams:
          (json['manual_feed_portion_grams'] as num?)?.toDouble() ?? 30,
      servoOpenDegrees: (json['servo_open_degrees'] as num?)?.toInt() ?? 25,
      servoClosedDegrees: (json['servo_closed_degrees'] as num?)?.toInt() ?? 55,
      automationEnabled: json['automation_enabled'] as bool? ?? false,
      lastSeenAt:
          DateTime.tryParse(json['last_seen_at'] as String? ?? '') ??
          DateTime.now(),
    );
  }
}

class DashboardOverview {
  const DashboardOverview({
    required this.greetingTitle,
    required this.greetingSubtitle,
    required this.pet,
    required this.device,
  });

  final String greetingTitle;
  final String greetingSubtitle;
  final PetProfile pet;
  final DeviceStatus device;

  factory DashboardOverview.fromJson(Map<String, dynamic> json) {
    return DashboardOverview(
      greetingTitle: json['greeting_title'] as String? ?? 'Hello!',
      greetingSubtitle: json['greeting_subtitle'] as String? ?? '',
      pet: PetProfile.fromJson(json['pet'] as Map<String, dynamic>? ?? {}),
      device: DeviceStatus.fromJson(
        json['device'] as Map<String, dynamic>? ?? {},
      ),
    );
  }
}

class DailyConsumption {
  const DailyConsumption({
    required this.date,
    required this.dayLabel,
    required this.totalGrams,
  });

  final DateTime date;
  final String dayLabel;
  final double totalGrams;

  factory DailyConsumption.fromJson(Map<String, dynamic> json) {
    return DailyConsumption(
      date: DateTime.parse(json['date'] as String),
      dayLabel: json['day_label'] as String? ?? '',
      totalGrams: (json['total_grams'] as num?)?.toDouble() ?? 0,
    );
  }
}

class WeeklyConsumption {
  const WeeklyConsumption({
    required this.days,
    required this.dailyTargetGrams,
    required this.totalGrams,
    required this.averageGrams,
    required this.recommendedDaysCount,
  });

  final List<DailyConsumption> days;
  final double dailyTargetGrams;
  final double totalGrams;
  final double averageGrams;
  final int recommendedDaysCount;

  factory WeeklyConsumption.fromJson(Map<String, dynamic> json) {
    final items = json['data'] as List<dynamic>? ?? [];
    return WeeklyConsumption(
      days: items
          .map(
            (item) => DailyConsumption.fromJson(item as Map<String, dynamic>),
          )
          .toList(),
      dailyTargetGrams: (json['daily_target_grams'] as num?)?.toDouble() ?? 150,
      totalGrams: (json['total_grams'] as num?)?.toDouble() ?? 0,
      averageGrams: (json['average_grams'] as num?)?.toDouble() ?? 0,
      recommendedDaysCount:
          (json['recommended_days_count'] as num?)?.toInt() ?? 7,
    );
  }
}

class HealthVitals {
  const HealthVitals({
    required this.weightKg,
    required this.activityMinutes,
    required this.sleepHours,
  });

  final double weightKg;
  final int activityMinutes;
  final double sleepHours;

  factory HealthVitals.fromJson(Map<String, dynamic> json) {
    return HealthVitals(
      weightKg: (json['weight_kg'] as num?)?.toDouble() ?? 0,
      activityMinutes: (json['activity_minutes'] as num?)?.toInt() ?? 0,
      sleepHours: (json['sleep_hours'] as num?)?.toDouble() ?? 0,
    );
  }
}

class CareTask {
  const CareTask({
    required this.id,
    required this.category,
    required this.title,
    required this.subtitle,
    required this.description,
    required this.dueLabel,
    required this.dueDate,
    required this.status,
    required this.priority,
  });

  final int id;
  final String category;
  final String title;
  final String subtitle;
  final String description;
  final String dueLabel;
  final DateTime? dueDate;
  final String status;
  final String priority;

  int? get daysUntilDue {
    final due = dueDate;
    if (due == null) return null;
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final dueDay = DateTime(due.year, due.month, due.day);
    return dueDay.difference(today).inDays;
  }

  bool get isDueWithinSevenDays {
    final days = daysUntilDue;
    return days != null && days <= 7;
  }

  String get displayDueLabel {
    final days = daysUntilDue;
    if (days == null) return dueLabel;
    if (days < 0) return 'Overdue';
    if (days == 0) return 'Due today';
    if (days <= 7) return days == 1 ? 'Due in 1 day' : 'Due in $days days';
    return _formatShortDate(dueDate!);
  }

  factory CareTask.fromJson(Map<String, dynamic> json) {
    final description =
        (json['description'] as String?) ?? (json['subtitle'] as String?) ?? '';
    final dueValue =
        (json['due_date'] as String?) ?? (json['due_at'] as String?);
    return CareTask(
      id: (json['id'] as num?)?.toInt() ?? 0,
      category: json['category'] as String? ?? '',
      title: json['title'] as String? ?? '',
      subtitle: description,
      description: description,
      dueLabel: json['due_label'] as String? ?? '',
      dueDate: dueValue == null || dueValue.isEmpty
          ? null
          : DateTime.tryParse(dueValue),
      status: json['status'] as String? ?? 'pending',
      priority: json['priority'] as String? ?? 'normal',
    );
  }
}

String _formatShortDate(DateTime value) {
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec',
  ];
  return '${months[value.month - 1]} ${value.day}';
}

class HealthSummary {
  const HealthSummary({
    required this.pet,
    required this.score,
    required this.statusLabel,
    required this.headline,
    required this.description,
    required this.vitals,
    required this.upcomingTasks,
  });

  final PetProfile pet;
  final int score;
  final String statusLabel;
  final String headline;
  final String description;
  final HealthVitals vitals;
  final List<CareTask> upcomingTasks;

  factory HealthSummary.fromJson(Map<String, dynamic> json) {
    final tasks = json['upcoming_tasks'] as List<dynamic>? ?? [];
    return HealthSummary(
      pet: PetProfile.fromJson(json['pet'] as Map<String, dynamic>? ?? {}),
      score: (json['score'] as num?)?.toInt() ?? 0,
      statusLabel: json['status_label'] as String? ?? '',
      headline: json['headline'] as String? ?? '',
      description: json['description'] as String? ?? '',
      vitals: HealthVitals.fromJson(
        json['vitals'] as Map<String, dynamic>? ?? {},
      ),
      upcomingTasks: tasks
          .map((item) => CareTask.fromJson(item as Map<String, dynamic>))
          .toList(),
    );
  }

  HealthSummary copyWith({
    PetProfile? pet,
    int? score,
    String? statusLabel,
    String? headline,
    String? description,
    HealthVitals? vitals,
    List<CareTask>? upcomingTasks,
  }) {
    return HealthSummary(
      pet: pet ?? this.pet,
      score: score ?? this.score,
      statusLabel: statusLabel ?? this.statusLabel,
      headline: headline ?? this.headline,
      description: description ?? this.description,
      vitals: vitals ?? this.vitals,
      upcomingTasks: upcomingTasks ?? this.upcomingTasks,
    );
  }
}

class UserAlert {
  const UserAlert({
    required this.id,
    required this.type,
    required this.title,
    required this.message,
    required this.severity,
    required this.dueDate,
  });

  final String id;
  final String type;
  final String title;
  final String message;
  final String severity;
  final DateTime? dueDate;

  factory UserAlert.fromJson(Map<String, dynamic> json) {
    final dueValue = json['due_date'] as String?;
    return UserAlert(
      id: json['id'] as String? ?? '',
      type: json['type'] as String? ?? '',
      title: json['title'] as String? ?? '',
      message: json['message'] as String? ?? '',
      severity: json['severity'] as String? ?? 'info',
      dueDate: dueValue == null || dueValue.isEmpty
          ? null
          : DateTime.tryParse(dueValue),
    );
  }
}

class ProfileSummary {
  const ProfileSummary({
    required this.owner,
    required this.pet,
    required this.device,
  });

  final Owner owner;
  final PetProfile pet;
  final DeviceStatus device;

  factory ProfileSummary.fromJson(Map<String, dynamic> json) {
    return ProfileSummary(
      owner: Owner.fromJson(json['owner'] as Map<String, dynamic>? ?? {}),
      pet: PetProfile.fromJson(json['pet'] as Map<String, dynamic>? ?? {}),
      device: DeviceStatus.fromJson(
        json['device'] as Map<String, dynamic>? ?? {},
      ),
    );
  }
}
