import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter_rating_bar/flutter_rating_bar.dart';
import '../../../core/network/api_client.dart';

final _gigProvider = FutureProvider.family<Map<String, dynamic>, String>((ref, slug) async {
  final res = await ref.read(dioProvider).get('/api/v1/services/$slug');
  return res.data as Map<String, dynamic>;
});

class GigDetailScreen extends ConsumerStatefulWidget {
  final String slug;
  const GigDetailScreen({super.key, required this.slug});

  @override
  ConsumerState<GigDetailScreen> createState() => _GigDetailScreenState();
}

class _GigDetailScreenState extends ConsumerState<GigDetailScreen> {
  int _selectedPackage = 0;

  @override
  Widget build(BuildContext context) {
    final gig    = ref.watch(_gigProvider(widget.slug));
    final scheme = Theme.of(context).colorScheme;
    final theme  = Theme.of(context);

    return Scaffold(
      body: gig.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (service) {
          final packages  = service['packages'] as List? ?? [];
          final reviews   = service['reviews']  as List? ?? [];
          final faqs      = service['faqs']     as List? ?? [];
          final gallery   = service['gallery']  as List? ?? [];
          final seller    = service['seller']   as Map? ?? {};
          final profile   = seller['profile']   as Map? ?? {};
          final sellerPkg = packages.isNotEmpty ? packages[_selectedPackage] : null;

          return CustomScrollView(
            slivers: [
              // App bar with gallery
              SliverAppBar(
                expandedHeight: 260,
                pinned: true,
                flexibleSpace: FlexibleSpaceBar(
                  background: gallery.isNotEmpty
                      ? CachedNetworkImage(imageUrl: gallery[0], fit: BoxFit.cover)
                      : Container(color: scheme.surfaceContainerHighest,
                          child: Center(child: Icon(Icons.work_outline_rounded, size: 56, color: scheme.onSurface.withValues(alpha: 0.2)))),
                ),
              ),

              SliverPadding(
                padding: const EdgeInsets.all(16),
                sliver: SliverList(
                  delegate: SliverChildListDelegate([
                    // Category badge
                    if (service['category'] != null)
                      Chip(
                        label: Text(service['category']['name'], style: const TextStyle(fontSize: 12)),
                        padding: EdgeInsets.zero,
                        visualDensity: VisualDensity.compact,
                      ),
                    const SizedBox(height: 8),

                    // Title
                    Text(service['title'] ?? '', style: theme.textTheme.headlineMedium),
                    const SizedBox(height: 12),

                    // Rating row
                    Row(children: [
                      RatingBarIndicator(
                        rating: (service['rating'] as num? ?? 0).toDouble(),
                        itemBuilder: (_, __) => const Icon(Icons.star, color: Colors.amber),
                        itemCount: 5,
                        itemSize: 18,
                      ),
                      const SizedBox(width: 6),
                      Text('${(service['rating'] as num? ?? 0).toStringAsFixed(1)} (${service['total_reviews']} reviews)',
                          style: theme.textTheme.bodySmall),
                      const Spacer(),
                      Icon(Icons.visibility_outlined, size: 16, color: scheme.onSurface.withValues(alpha: 0.5)),
                      const SizedBox(width: 4),
                      Text('${service['views']}', style: theme.textTheme.bodySmall),
                    ]),
                    const SizedBox(height: 20),

                    // Seller card
                    Card(
                      child: ListTile(
                        leading: CircleAvatar(
                          backgroundImage: NetworkImage(
                            profile['avatar_url'] as String? ??
                            'https://ui-avatars.com/api/?name=${seller['username']}&size=48&background=2563eb&color=fff',
                          ),
                          radius: 24,
                        ),
                        title: Text(seller['username'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600)),
                        subtitle: Row(mainAxisSize: MainAxisSize.min, children: [
                          if ((profile['university'] as String? ?? '').isNotEmpty) ...[
                            Text('${profile['university']} · '),
                          ],
                          const Icon(Icons.star_rounded, size: 13, color: Colors.amber),
                          Text(' ${(profile['rating'] as num? ?? 0).toStringAsFixed(1)}'),
                        ]),
                        trailing: OutlinedButton(
                          onPressed: () {},
                          child: const Text('Contact'),
                        ),
                      ),
                    ),
                    const SizedBox(height: 20),

                    // Packages
                    if (packages.isNotEmpty) ...[
                      Text('Packages', style: theme.textTheme.titleLarge),
                      const SizedBox(height: 12),
                      Row(
                        children: List.generate(packages.length, (i) {
                          final pkg = packages[i];
                          final selected = _selectedPackage == i;
                          return Expanded(
                            child: GestureDetector(
                              onTap: () => setState(() => _selectedPackage = i),
                              child: AnimatedContainer(
                                duration: const Duration(milliseconds: 200),
                                margin: EdgeInsets.only(right: i < packages.length - 1 ? 8 : 0),
                                padding: const EdgeInsets.all(12),
                                decoration: BoxDecoration(
                                  color: selected ? scheme.primary : scheme.surfaceContainerHighest,
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(color: selected ? scheme.primary : Colors.transparent, width: 2),
                                ),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      (pkg['name'] as String).toUpperCase(),
                                      style: TextStyle(
                                        fontSize: 11, fontWeight: FontWeight.w700,
                                        color: selected ? Colors.white70 : scheme.onSurface.withValues(alpha: 0.6),
                                      ),
                                    ),
                                    const SizedBox(height: 4),
                                    Text(
                                      '\$${(pkg['price'] as num).toStringAsFixed(0)}',
                                      style: TextStyle(
                                        fontSize: 18, fontWeight: FontWeight.w800,
                                        color: selected ? Colors.white : scheme.primary,
                                      ),
                                    ),
                                    const SizedBox(height: 2),
                                    Text(
                                      '${pkg['delivery_days']}d delivery',
                                      style: TextStyle(fontSize: 11, color: selected ? Colors.white70 : scheme.onSurface.withValues(alpha: 0.6)),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          );
                        }),
                      ),
                      if (sellerPkg != null) ...[
                        const SizedBox(height: 12),
                        Card(
                          child: Padding(
                            padding: const EdgeInsets.all(14),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(sellerPkg['title'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600)),
                                if ((sellerPkg['description'] as String? ?? '').isNotEmpty) ...[
                                  const SizedBox(height: 6),
                                  Text(sellerPkg['description'], style: theme.textTheme.bodySmall),
                                ],
                                const SizedBox(height: 10),
                                ...((sellerPkg['features'] as List? ?? []).map((f) => Padding(
                                  padding: const EdgeInsets.only(bottom: 4),
                                  child: Row(children: [
                                    Icon(Icons.check_circle, size: 16, color: scheme.primary),
                                    const SizedBox(width: 8),
                                    Text(f.toString(), style: theme.textTheme.bodySmall),
                                  ]),
                                ))),
                                const SizedBox(height: 4),
                                Row(children: [
                                  Icon(Icons.refresh, size: 14, color: scheme.onSurface.withValues(alpha: 0.5)),
                                  const SizedBox(width: 4),
                                  Text('${sellerPkg['revisions']} revision(s)', style: theme.textTheme.bodySmall),
                                ]),
                              ],
                            ),
                          ),
                        ),
                      ],
                    ],
                    const SizedBox(height: 20),

                    // Description
                    Text('About this service', style: theme.textTheme.titleLarge),
                    const SizedBox(height: 8),
                    Text(service['description'] ?? '', style: theme.textTheme.bodyMedium?.copyWith(height: 1.6)),
                    const SizedBox(height: 20),

                    // FAQs
                    if (faqs.isNotEmpty) ...[
                      Text('FAQ', style: theme.textTheme.titleLarge),
                      const SizedBox(height: 8),
                      ...faqs.map((faq) => ExpansionTile(
                        title: Text(faq['question'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                        children: [Padding(padding: const EdgeInsets.fromLTRB(16, 0, 16, 12), child: Text(faq['answer'] ?? ''))],
                      )),
                      const SizedBox(height: 20),
                    ],

                    // Reviews
                    if (reviews.isNotEmpty) ...[
                      Text('Reviews (${service['total_reviews']})', style: theme.textTheme.titleLarge),
                      const SizedBox(height: 12),
                      ...reviews.take(5).map((r) => Padding(
                        padding: const EdgeInsets.only(bottom: 16),
                        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                          Row(children: [
                            CircleAvatar(radius: 16, backgroundImage: NetworkImage(
                              r['reviewer']?['profile']?['avatar_url'] ??
                              'https://ui-avatars.com/api/?name=${r['reviewer']?['username']}&size=32',
                            )),
                            const SizedBox(width: 8),
                            Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                              Text(r['reviewer']?['username'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
                              RatingBarIndicator(
                                rating: (r['rating'] as num? ?? 0).toDouble(),
                                itemBuilder: (_, __) => const Icon(Icons.star, color: Colors.amber),
                                itemCount: 5,
                                itemSize: 14,
                              ),
                            ]),
                          ]),
                          if ((r['content'] as String? ?? '').isNotEmpty) ...[
                            const SizedBox(height: 6),
                            Text(r['content'], style: theme.textTheme.bodySmall?.copyWith(height: 1.5)),
                          ],
                          const Divider(height: 24),
                        ]),
                      )),
                    ],

                    const SizedBox(height: 100),
                  ]),
                ),
              ),
            ],
          );
        },
      ),
      bottomNavigationBar: gig.whenData((service) {
        final packages = service['packages'] as List? ?? [];
        if (packages.isEmpty) return const SizedBox.shrink();
        final pkg = packages[_selectedPackage];
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: FilledButton(
              onPressed: () => _placeOrder(service, pkg),
              style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 16)),
              child: Text(
                'Order Now · \$${(pkg['price'] as num).toStringAsFixed(0)}',
                style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700),
              ),
            ),
          ),
        );
      }).value,
    );
  }

  void _placeOrder(Map service, Map pkg) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (_) => _OrderSheet(service: service, package: pkg),
    );
  }
}

class _OrderSheet extends ConsumerStatefulWidget {
  final Map service;
  final Map package;
  const _OrderSheet({required this.service, required this.package});

  @override
  ConsumerState<_OrderSheet> createState() => _OrderSheetState();
}

class _OrderSheetState extends ConsumerState<_OrderSheet> {
  final _reqCtrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() { _reqCtrl.dispose(); super.dispose(); }

  Future<void> _confirm() async {
    setState(() => _loading = true);
    try {
      await ref.read(dioProvider).post('/api/v1/orders', data: {
        'service_id':   widget.service['id'],
        'package_id':   widget.package['id'],
        'requirements': _reqCtrl.text.trim(),
      });
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Order placed!'), backgroundColor: Colors.green));
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e'), backgroundColor: Colors.red));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.fromLTRB(24, 24, 24, MediaQuery.of(context).viewInsets.bottom + 24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text('Confirm Order', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 8),
          Text('${widget.package['title']} · \$${(widget.package['price'] as num).toStringAsFixed(0)} · ${widget.package['delivery_days']} days'),
          const SizedBox(height: 20),
          TextField(
            controller: _reqCtrl,
            maxLines: 4,
            decoration: const InputDecoration(labelText: 'Describe your requirements', hintText: 'Be as specific as possible...'),
          ),
          const SizedBox(height: 20),
          FilledButton(
            onPressed: _loading ? null : _confirm,
            child: _loading ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white)) : const Text('Place Order'),
          ),
        ],
      ),
    );
  }
}
