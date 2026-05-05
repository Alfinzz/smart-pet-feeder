import '../models/feed_history_item.dart';
import 'api_client.dart';

class FeedService {
  const FeedService(this._apiClient);

  final ApiClient _apiClient;

  Future<List<FeedHistoryItem>> fetchHistory({int limit = 50}) async {
    final response = await _apiClient.dio.get<Map<String, dynamic>>(
      '/feed/history',
      queryParameters: {'limit': limit},
    );

    final items = response.data?['data'] as List<dynamic>? ?? [];
    return items
        .map((item) => FeedHistoryItem.fromJson(item as Map<String, dynamic>))
        .toList();
  }
}
