import 'package:flutter/material.dart';

class PetPhotoAvatar extends StatelessWidget {
  const PetPhotoAvatar({
    super.key,
    required this.photoUrl,
    this.size = 100,
    this.iconSize = 44,
    this.borderWidth = 4,
    this.isUploading = false,
    this.showCameraBadge = false,
    this.onTap,
  });

  final String photoUrl;
  final double size;
  final double iconSize;
  final double borderWidth;
  final bool isUploading;
  final bool showCameraBadge;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final trimmedPhotoUrl = photoUrl.trim();

    return GestureDetector(
      onTap: isUploading ? null : onTap,
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          Container(
            width: size,
            height: size,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: const Color(0xFFF1F5F9),
              border: Border.all(color: Colors.white, width: borderWidth),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.08),
                  blurRadius: 16,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            clipBehavior: Clip.antiAlias,
            child: trimmedPhotoUrl.isEmpty
                ? _fallbackIcon()
                : Image.network(
                    trimmedPhotoUrl,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) =>
                        _fallbackIcon(),
                  ),
          ),
          if (isUploading)
            Positioned.fill(
              child: Container(
                decoration: BoxDecoration(
                  color: Colors.black.withValues(alpha: 0.35),
                  shape: BoxShape.circle,
                ),
                child: const Center(
                  child: SizedBox(
                    width: 24,
                    height: 24,
                    child: CircularProgressIndicator(
                      strokeWidth: 2.5,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            ),
          if (showCameraBadge)
            Positioned(
              right: -2,
              bottom: -2,
              child: Container(
                width: 32,
                height: 32,
                decoration: BoxDecoration(
                  color: const Color(0xFF1565C0),
                  shape: BoxShape.circle,
                  border: Border.all(color: Colors.white, width: 3),
                ),
                child: const Icon(
                  Icons.camera_alt,
                  size: 16,
                  color: Colors.white,
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _fallbackIcon() {
    return Icon(Icons.pets, size: iconSize, color: const Color(0xFF1565C0));
  }
}
