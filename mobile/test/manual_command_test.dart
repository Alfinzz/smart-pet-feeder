import 'package:flutter_test/flutter_test.dart';
import 'package:smart_pet_monitoring/models/manual_command.dart';

void main() {
  test('ManualCommand parses delivery status fields', () {
    final command = ManualCommand.fromJson({
      'id': 12,
      'device_id': 'ESP32-001',
      'action': 'feed',
      'status': 'failed',
      'attempt_count': 2,
      'last_error': 'load cell pakan tidak siap',
      'created_at': '2026-05-22T10:00:00Z',
      'updated_at': '2026-05-22T10:00:10Z',
      'completed_at': null,
    });

    expect(command.id, 12);
    expect(command.deviceId, 'ESP32-001');
    expect(command.attemptCount, 2);
    expect(command.lastError, 'load cell pakan tidak siap');
    expect(command.isTerminal, isTrue);
    expect(command.completedAt, isNull);
  });
}
