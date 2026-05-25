import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_rating_bar/flutter_rating_bar.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../../core/network/api_client.dart';

final _profileProvider = FutureProvider.family<Map<String, dynamic>, String>((ref, username) async {
  final res = await ref.read(dioProvider).get('/api/v1/users/$username');
  return res.data as Map<String, dynamic>;
});

class ProfileScreen extends ConsumerWidget {
  final String username;
  const ProfileScreen({super.key, required this.username});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profile = ref.watch(_profileProvider(username));
    final theme   = Theme.of(context);
    final scheme  = theme.colorScheme;

    return Scaffold(
      body: profile.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => const Center(child: Text('User not found')),
        data: (data) {
          final p        = data['profile'] as Map? ?? {};
          final skills   = (p['skills']   as List? ?? []).cast<String>();
          final services = data['services'] as List? ?? [];
          final avatarUrl = p['avatar_url'] as String? ?? '';
          final coverUrl  = p['cover_url']  as String? ?? '';

          return CustomScrollView(
            slivers: [
              // Cover + Avatar
              SliverAppBar(
                expandedHeight: 200,
                pinned: true,
                flexibleSpace: FlexibleSpaceBar(
                  background: Stack(
                    fit: StackFit.expand,
                    children: [
                      coverUrl.isNotEmpty
                          ? CachedNetworkImage(imageUrl: coverUrl, fit: BoxFit.cover)
                          : Container(color: scheme.primary),
                      Positioned(
                        bottom: 16, left: 16,
                        child: CircleAvatar(
                          radius: 40,
                          backgroundColor: scheme.surface,
                          child: CircleAvatar(
                            radius: 37,
                            backgroundImage: NetworkImage(
                              avatarUrl.isNotEmpty ? avatarUrl
                                  : 'https://ui-avatars.com/api/?name=${data['username']}&size=80&background=2563eb&color=fff',
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),

              SliverPadding(
                padding: const EdgeInsets.all(16),
                sliver: SliverList(
                  delegate: SliverChildListDelegate([
                    // Name + follow
                    Row(children: [
                      Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                        Text(
                          '${p['first_name'] ?? ''} ${p['last_name'] ?? ''}'.trim().isNotEmpty
                              ? '${p['first_name']} ${p['last_name']}'
                              : data['username'],
                          style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w700),
                        ),
                        Text('@${data['username']}', style: theme.textTheme.bodySmall),
                      ]),
                      const Spacer(),
                      FilledButton.tonal(onPressed: () {}, child: const Text('Follow')),
                      const SizedBox(width: 8),
                      OutlinedButton(onPressed: () {}, child: const Text('Message')),
                    ]),
                    const SizedBox(height: 12),

                    // Bio
                    if ((p['tagline'] as String? ?? '').isNotEmpty)
                      Text(p['tagline'], style: theme.textTheme.bodyMedium?.copyWith(color: scheme.primary, fontWeight: FontWeight.w500)),
                    if ((p['bio'] as String? ?? '').isNotEmpty) ...[
                      const SizedBox(height: 6),
                      Text(p['bio'], style: theme.textTheme.bodyMedium?.copyWith(height: 1.5)),
                    ],
                    const SizedBox(height: 16),

                    // Stats row
                    Row(children: [
                      _stat((p['rating'] as num? ?? 0).toStringAsFixed(1), 'Rating', theme),
                      _stat('${p['total_reviews'] ?? 0}', 'Reviews', theme),
                      _stat('${p['completed_jobs'] ?? 0}', 'Jobs done', theme),
                    ]),
                    const SizedBox(height: 8),

                    // Rating stars
                    if ((p['total_reviews'] as int? ?? 0) > 0)
                      RatingBarIndicator(
                        rating: (p['rating'] as num? ?? 0).toDouble(),
                        itemBuilder: (_, __) => const Icon(Icons.star, color: Colors.amber),
                        itemCount: 5,
                        itemSize: 20,
                      ),
                    const SizedBox(height: 16),

                    // Info chips
                    Wrap(spacing: 8, runSpacing: 8, children: [
                      if ((p['university'] as String? ?? '').isNotEmpty)
                        _infoChip(Icons.school_outlined, p['university'], scheme),
                      if ((p['location'] as String? ?? '').isNotEmpty)
                        _infoChip(Icons.location_on_outlined, p['location'], scheme),
                      if ((p['department'] as String? ?? '').isNotEmpty)
                        _infoChip(Icons.book_outlined, p['department'], scheme),
                    ]),
                    const SizedBox(height: 16),

                    // Skills
                    if (skills.isNotEmpty) ...[
                      Text('Skills', style: theme.textTheme.titleMedium),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 8, runSpacing: 8,
                        children: skills.map((s) => Chip(
                          label: Text(s, style: const TextStyle(fontSize: 12)),
                          visualDensity: VisualDensity.compact,
                          padding: const EdgeInsets.symmetric(horizontal: 4),
                        )).toList(),
                      ),
                      const SizedBox(height: 20),
                    ],

                    // Services
                    if (services.isNotEmpty) ...[
                      Text('Services (${services.length})', style: theme.textTheme.titleMedium),
                      const SizedBox(height: 12),
                      ...services.take(3).map((s) => Card(
                        margin: const EdgeInsets.only(bottom: 8),
                        child: ListTile(
                          title: Text(s['title'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13), maxLines: 1, overflow: TextOverflow.ellipsis),
                          subtitle: Row(mainAxisSize: MainAxisSize.min, children: [
                            const Icon(Icons.star_rounded, size: 13, color: Colors.amber),
                            Text(' ${(s['rating'] as num? ?? 0).toStringAsFixed(1)} · ${s['orders_count']} orders', style: const TextStyle(fontSize: 12)),
                          ]),
                          trailing: Text(
                            'From \$${((s['packages'] as List?)?.map<double>((p) => (p['price'] as num).toDouble()).fold(double.infinity, (a, b) => a < b ? a : b) ?? 0).toStringAsFixed(0)}',
                            style: TextStyle(fontWeight: FontWeight.w700, color: scheme.primary, fontSize: 13),
                          ),
                        ),
                      )),
                    ],
                  ]),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _stat(String value, String label, ThemeData theme) => Expanded(
    child: Column(children: [
      Text(value, style: theme.textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w800)),
      Text(label,  style: theme.textTheme.bodySmall),
    ]),
  );

  Widget _infoChip(IconData icon, String label, ColorScheme scheme) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
    decoration: BoxDecoration(
      color: scheme.surfaceContainerHighest,
      borderRadius: BorderRadius.circular(20),
    ),
    child: Row(mainAxisSize: MainAxisSize.min, children: [
      Icon(icon, size: 14, color: scheme.onSurface.withValues(alpha: 0.6)),
      const SizedBox(width: 5),
      Text(label, style: TextStyle(fontSize: 12, color: scheme.onSurface.withValues(alpha: 0.8))),
    ]),
  );
}
