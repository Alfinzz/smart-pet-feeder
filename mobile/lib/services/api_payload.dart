Map<String, dynamic> objectMap(Object? value) {
  if (value is Map<String, dynamic>) return value;
  if (value is Map) return Map<String, dynamic>.from(value);
  return {};
}

String stringValue(Object? value, Object? fallbackValue, String fallback) {
  final candidate = value ?? fallbackValue;
  if (candidate is String && candidate.trim().isNotEmpty) {
    return candidate.trim();
  }
  return fallback;
}

int intValue(Object? value, Object? fallbackValue, {int fallback = 0}) {
  final candidate = value ?? fallbackValue;
  if (candidate is num) return candidate.toInt();
  if (candidate is String) return int.tryParse(candidate) ?? fallback;
  return fallback;
}

double doubleValue(
  Object? value,
  Object? fallbackValue, {
  double fallback = 0,
}) {
  final candidate = value ?? fallbackValue;
  if (candidate is num) return candidate.toDouble();
  if (candidate is String) return double.tryParse(candidate) ?? fallback;
  return fallback;
}

double portionToGrams(Object? value) {
  if (value is num) return value.toDouble();
  if (value is String) {
    switch (value) {
      case '1/8c':
        return 15;
      case '1/4c':
        return 30;
      case '1/2c':
        return 60;
      case '1c':
        return 120;
    }
    return double.tryParse(value) ?? 30;
  }
  return 30;
}
