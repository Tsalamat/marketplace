import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:google_fonts/google_fonts.dart';

class ServiceCard extends StatelessWidget {
  final Map<String, dynamic> service;
  final VoidCallback? onTap;
  const ServiceCard({super.key, required this.service, this.onTap});

  double get _minPrice {
    final packages = service['packages'] as List? ?? [];
    if (packages.isEmpty) return 0;
    return packages.map<double>((p) => (p['price'] as num).toDouble()).reduce((a, b) => a < b ? a : b);
  }

  double get _rating => (service['rating'] as num? ?? 0).toDouble();
  int    get _reviews => service['total_reviews'] as int? ?? 0;

  String get _avatar {
    final url = service['seller']?['profile']?['avatar_url'] as String?;
    if (url != null && url.isNotEmpty) return url;
    return 'https://ui-avatars.com/api/?name=${service['seller']?['username'] ?? 'U'}&background=2563eb&color=fff&size=40';
  }

  String? get _gallery => (service['gallery'] as List?)?.firstOrNull as String?;

  IconData _categoryIcon(String? slug) {
    return switch (slug) {
      'programming'  => Icons.code_rounded,
      'design'       => Icons.palette_rounded,
      'tutoring'     => Icons.menu_book_rounded,
      'writing'      => Icons.edit_note_rounded,
      'video'        => Icons.videocam_rounded,
      'photography'  => Icons.photo_camera_rounded,
      'delivery'     => Icons.local_shipping_rounded,
      'fitness'      => Icons.fitness_center_rounded,
      'music'        => Icons.music_note_rounded,
      'business'     => Icons.business_center_rounded,
      _              => Icons.work_rounded,
    };
  }

  @override
  Widget build(BuildContext context) {
    final scheme  = Theme.of(context).colorScheme;
    final catSlug = service['category']?['slug'] as String?;

    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Thumbnail
            Expanded(
              flex: 5,
              child: Stack(
                fit: StackFit.expand,
                children: [
                  _gallery != null
                      ? CachedNetworkImage(
                          imageUrl: _gallery!,
                          fit: BoxFit.cover,
                          placeholder: (_, __) => Container(color: scheme.surfaceContainerHighest),
                          errorWidget: (_, __, ___) => _placeholder(scheme, catSlug),
                        )
                      : _placeholder(scheme, catSlug),
                  if (service['is_featured'] == true)
                    Positioned(
                      top: 8, left: 8,
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
                        decoration: BoxDecoration(color: Colors.amber[600], borderRadius: BorderRadius.circular(6)),
                        child: Row(mainAxisSize: MainAxisSize.min, children: [
                          const Icon(Icons.star_rounded, size: 10, color: Colors.white),
                          const SizedBox(width: 2),
                          Text('Featured', style: GoogleFonts.inter(fontSize: 9, fontWeight: FontWeight.w700, color: Colors.white)),
                        ]),
                      ),
                    ),
                ],
              ),
            ),

            // Info
            Expanded(
              flex: 6,
              child: Padding(
                padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Seller
                    Row(children: [
                      CircleAvatar(radius: 11, backgroundImage: NetworkImage(_avatar)),
                      const SizedBox(width: 5),
                      Expanded(
                        child: Text(
                          service['seller']?['username'] ?? '',
                          style: GoogleFonts.inter(fontSize: 11, fontWeight: FontWeight.w600, color: scheme.onSurface.withValues(alpha: 0.65)),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      if (service['seller']?['profile']?['is_online'] == true)
                        Container(width: 7, height: 7, decoration: const BoxDecoration(color: Colors.green, shape: BoxShape.circle)),
                    ]),
                    const SizedBox(height: 6),

                    // Title
                    Expanded(
                      child: Text(
                        service['title'] ?? '',
                        style: GoogleFonts.inter(fontSize: 12.5, fontWeight: FontWeight.w600, height: 1.35),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),

                    // Stars
                    Row(children: [
                      ...List.generate(5, (i) => Icon(
                        i < _rating.round() ? Icons.star_rounded : Icons.star_outline_rounded,
                        size: 12,
                        color: i < _rating.round() ? Colors.amber[600] : scheme.onSurface.withValues(alpha: 0.25),
                      )),
                      const SizedBox(width: 4),
                      Text(
                        _rating > 0 ? _rating.toStringAsFixed(1) : 'New',
                        style: GoogleFonts.inter(fontSize: 10.5, fontWeight: FontWeight.w600, color: scheme.onSurface.withValues(alpha: 0.6)),
                      ),
                      if (_reviews > 0)
                        Text(
                          ' ($_reviews)',
                          style: GoogleFonts.inter(fontSize: 10.5, color: scheme.onSurface.withValues(alpha: 0.45)),
                        ),
                    ]),
                    const SizedBox(height: 5),

                    // Price row
                    Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
                      Text('From', style: GoogleFonts.inter(fontSize: 10.5, color: scheme.onSurface.withValues(alpha: 0.5))),
                      Text(
                        '\$${_minPrice.toStringAsFixed(0)}',
                        style: GoogleFonts.inter(fontSize: 14, fontWeight: FontWeight.w800, color: scheme.primary),
                      ),
                    ]),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _placeholder(ColorScheme scheme, String? catSlug) => Container(
    color: scheme.primary.withValues(alpha: 0.06),
    child: Center(
      child: Icon(_categoryIcon(catSlug), size: 36, color: scheme.primary.withValues(alpha: 0.35)),
    ),
  );
}
