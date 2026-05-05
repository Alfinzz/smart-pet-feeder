import 'package:flutter/material.dart';

class AppHeader extends StatelessWidget {
  const AppHeader({super.key, this.showBackButton = false});

  final bool showBackButton;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
      child: Row(
        children: [
          if (showBackButton)
            IconButton(
              onPressed: () => Navigator.of(context).pop(),
              icon: const Icon(Icons.arrow_back_ios_new, size: 20),
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(),
            ),
          // Paw logo
          Container(
            width: 36,
            height: 36,
            decoration: BoxDecoration(
              color: const Color(0xFFE3F2FD),
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(Icons.pets, size: 20, color: Color(0xFF1565C0)),
          ),
          const SizedBox(width: 10),
          Text(
            'Smart Pet Feeder',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: const Color(0xFF1565C0),
                ),
          ),
          const Spacer(),
          // Notification bell
          Container(
            width: 38,
            height: 38,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(12),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.06),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: const Icon(
              Icons.notifications_outlined,
              size: 20,
              color: Color(0xFF64748B),
            ),
          ),
        ],
      ),
    );
  }
}
