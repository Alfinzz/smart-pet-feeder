class ManualCommand {
  const ManualCommand({
    required this.id,
    required this.deviceId,
    required this.action,
    required this.status,
    required this.attemptCount,
    required this.lastError,
    required this.createdAt,
    required this.updatedAt,
    required this.completedAt,
  });

  final int id;
  final String deviceId;
  final String action;
  final String status;
  final int attemptCount;
  final String lastError;
  final DateTime createdAt;
  final DateTime updatedAt;
  final DateTime? completedAt;

  bool get isTerminal => status == 'completed' || status == 'failed';

  factory ManualCommand.fromJson(Map<String, dynamic> json) {
    final completedAtText = json['completed_at'] as String?;
    return ManualCommand(
      id: (json['id'] as num).toInt(),
      deviceId: json['device_id'] as String? ?? '',
      action: json['action'] as String? ?? '',
      status: json['status'] as String? ?? '',
      attemptCount: (json['attempt_count'] as num?)?.toInt() ?? 0,
      lastError: json['last_error'] as String? ?? '',
      createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
      updatedAt:
          DateTime.tryParse(json['updated_at'] as String? ?? '')?.toLocal() ??
          DateTime.parse(json['created_at'] as String).toLocal(),
      completedAt: completedAtText == null
          ? null
          : DateTime.tryParse(completedAtText)?.toLocal(),
    );
  }
}
