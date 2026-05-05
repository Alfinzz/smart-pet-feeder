class Owner {
  const Owner({required this.id, required this.name, required this.email});

  final int id;
  final String name;
  final String email;

  factory Owner.fromJson(Map<String, dynamic> json) {
    return Owner(
      id: json['id'] as int,
      name: json['name'] as String,
      email: json['email'] as String,
    );
  }
}
