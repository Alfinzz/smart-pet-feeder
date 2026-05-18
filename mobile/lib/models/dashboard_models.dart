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
    return PetProfile(
      id: (json['id'] as num?)?.toInt() ?? 0,
      ownerId: (json['owner_id'] as num?)?.toInt() ?? 0,
      deviceId: json['device_id'] as String? ?? '',
      name: json['name'] as String? ?? 'Fluffy',
      species: json['species'] as String? ?? 'Dog',
      breed: json['breed'] as String? ?? 'Golden Retriever',
      ageYears: (json['age_years'] as num?)?.toInt() ?? 0,
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
    required this.lastSeenAt,
  });

  final String id;
  final String name;
  final double foodStockPercent;
  final String foodStockLabel;
  final bool waterAvailable;
  final String waterStatus;
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
    required this.dueLabel,
    required this.priority,
  });

  final int id;
  final String category;
  final String title;
  final String subtitle;
  final String dueLabel;
  final String priority;

  factory CareTask.fromJson(Map<String, dynamic> json) {
    return CareTask(
      id: (json['id'] as num?)?.toInt() ?? 0,
      category: json['category'] as String? ?? '',
      title: json['title'] as String? ?? '',
      subtitle: json['subtitle'] as String? ?? '',
      dueLabel: json['due_label'] as String? ?? '',
      priority: json['priority'] as String? ?? 'normal',
    );
  }
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
