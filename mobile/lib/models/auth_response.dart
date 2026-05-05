import 'owner.dart';

class AuthResponse {
  const AuthResponse({
    required this.token,
    required this.expiresAt,
    required this.owner,
  });

  final String token;
  final DateTime expiresAt;
  final Owner owner;

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      token: json['token'] as String,
      expiresAt: DateTime.parse(json['expires_at'] as String),
      owner: Owner.fromJson(json['owner'] as Map<String, dynamic>),
    );
  }
}
