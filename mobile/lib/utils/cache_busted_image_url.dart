String cacheBustedImageUrl(String imageUrl) {
  final trimmed = imageUrl.trim();
  if (trimmed.isEmpty) return '';

  final version = DateTime.now().millisecondsSinceEpoch.toString();
  final uri = Uri.tryParse(trimmed);
  if (uri != null && uri.hasScheme) {
    return uri
        .replace(queryParameters: {...uri.queryParameters, 'v': version})
        .toString();
  }

  final separator = trimmed.contains('?') ? '&' : '?';
  return '$trimmed${separator}v=$version';
}
