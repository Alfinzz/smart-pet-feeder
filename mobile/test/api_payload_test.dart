import 'package:flutter_test/flutter_test.dart';
import 'package:smart_pet_monitoring/services/api_payload.dart';

void main() {
  group('api payload helpers', () {
    test('stringValue trims valid strings and falls back for empty values', () {
      expect(stringValue(' Buddy ', null, 'Fluffy'), 'Buddy');
      expect(stringValue('', ' Pet ', 'Fluffy'), 'Fluffy');
      expect(stringValue(null, null, 'Fluffy'), 'Fluffy');
    });

    test('numeric helpers parse numbers and strings', () {
      expect(intValue('3', null), 3);
      expect(intValue(null, 4.8), 4);
      expect(intValue('bad', null, fallback: 7), 7);

      expect(doubleValue('2.5', null), 2.5);
      expect(doubleValue(null, 4), 4);
      expect(doubleValue('bad', null, fallback: 1.5), 1.5);
    });

    test('portionToGrams keeps legacy cup mappings', () {
      expect(portionToGrams('1/8c'), 15);
      expect(portionToGrams('1/4c'), 30);
      expect(portionToGrams('1/2c'), 60);
      expect(portionToGrams('1c'), 120);
      expect(portionToGrams('42'), 42);
      expect(portionToGrams(null), 30);
    });
  });
}
