class ManualCommand {
  const ManualCommand({
    required this.id,
    required this.deviceId,
    required this.action,
    required this.status,
    required this.createdAt,
  });

  final int id;
  final String deviceId;
  final String action;
  final String status;
  final DateTime createdAt;

  factory ManualCommand.fromJson(Map<String, dynamic> json) {
    return ManualCommand(
      id: json['id'] as int,
      deviceId: json['device_id'] as String,
      action: json['action'] as String,
      status: json['status'] as String,
      createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
    );
  }
}
