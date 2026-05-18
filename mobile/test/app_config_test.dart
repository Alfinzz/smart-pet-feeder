import 'package:flutter_test/flutter_test.dart';
import 'package:smart_pet_monitoring/config/app_config.dart';

void main() {
  group('AppConfig.publicUrl', () {
    test('keeps absolute URLs unchanged', () {
      expect(
        AppConfig.publicUrl('https://cdn.example.com/pet.jpg'),
        'https://cdn.example.com/pet.jpg',
      );
    });

    test('converts backend upload paths to production absolute URLs', () {
      expect(
        AppConfig.publicUrl('/uploads/pets/pet.jpg'),
        'https://smart-pet-feeder.alfian-gading.my.id/uploads/pets/pet.jpg',
      );
    });
  });
}
