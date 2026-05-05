class FeedHistoryItem {
  const FeedHistoryItem({
    required this.id,
    required this.deviceId,
    required this.weightGrams,
    required this.recordedAt,
    required this.createdAt,
  });

  final int id;
  final String deviceId;
  final double weightGrams;
  final DateTime recordedAt;
  final DateTime createdAt;

  factory FeedHistoryItem.fromJson(Map<String, dynamic> json) {
    return FeedHistoryItem(
      id: json['id'] as int,
      deviceId: json['device_id'] as String,
      weightGrams: (json['weight_grams'] as num).toDouble(),
      recordedAt: DateTime.parse(json['recorded_at'] as String).toLocal(),
      createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
    );
  }
}
