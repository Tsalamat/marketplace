import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/network/api_client.dart';
import '../../../shared/widgets/service_card.dart';

final _categoriesProvider = FutureProvider<List<dynamic>>((ref) async {
  final res = await ref.read(dioProvider).get('/api/v1/services/categories');
  return res.data as List;
});

final _servicesProvider = FutureProvider.family<Map<String, dynamic>, Map<String, dynamic>>((ref, params) async {
  final res = await ref.read(dioProvider).get('/api/v1/services', queryParameters: params);
  return res.data as Map<String, dynamic>;
});

class MarketplaceScreen extends ConsumerStatefulWidget {
  const MarketplaceScreen({super.key});

  @override
  ConsumerState<MarketplaceScreen> createState() => _MarketplaceScreenState();
}

class _MarketplaceScreenState extends ConsumerState<MarketplaceScreen> {
  final _searchCtrl = TextEditingController();
  Map<String, dynamic> _params = {'sort': 'trending', 'limit': 20};
  String _category = '';
  String _sort     = 'trending';
  Timer? _debounce;

  @override
  void dispose() {
    _searchCtrl.dispose();
    _debounce?.cancel();
    super.dispose();
  }

  void _fetch({String? category, String? sort}) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 350), () {
      if (!mounted) return;
      final cat = category ?? _category;
      final s   = sort      ?? _sort;
      final q   = _searchCtrl.text.trim();
      final p   = <String, dynamic>{'sort': s, 'limit': 20};
      if (q.isNotEmpty)   p['q']        = q;
      if (cat.isNotEmpty) p['category'] = cat;
      setState(() { _params = p; _category = cat; _sort = s; });
    });
  }

  void _clearFilters() {
    _searchCtrl.clear();
    setState(() { _params = {'sort': 'trending', 'limit': 20}; _category = ''; _sort = 'trending'; });
  }

  final _sortOptions = const [
    ('trending',   'Trending',          Icons.local_fire_department_rounded),
    ('newest',     'Newest',            Icons.new_releases_rounded),
    ('rating',     'Top Rated',         Icons.star_rounded),
    ('price_asc',  'Price: Low → High', Icons.arrow_upward_rounded),
    ('price_desc', 'Price: High → Low', Icons.arrow_downward_rounded),
  ];

  @override
  Widget build(BuildContext context) {
    final categories = ref.watch(_categoriesProvider);
    final services   = ref.watch(_servicesProvider(_params));
    final scheme     = Theme.of(context).colorScheme;
    final theme      = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Marketplace'),
        actions: [
          IconButton(
            icon: const Icon(Icons.tune_rounded),
            tooltip: 'Sort',
            onPressed: _showSort,
          ),
          const SizedBox(width: 4),
        ],
      ),
      body: Column(
        children: [
          // Search bar
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
            child: TextField(
              controller: _searchCtrl,
              decoration: InputDecoration(
                hintText: 'Search services...',
                prefixIcon: const Icon(Icons.search_rounded, size: 20),
                suffixIcon: _searchCtrl.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(Icons.close_rounded, size: 18),
                        onPressed: () { _searchCtrl.clear(); _fetch(); },
                      )
                    : null,
              ),
              onChanged: (_) => _fetch(),
              onSubmitted: (_) => _fetch(),
            ),
          ),

          // Category chips
          categories.when(
            loading: () => const SizedBox(height: 52),
            error: (_, __) => const SizedBox.shrink(),
            data: (cats) => SizedBox(
              height: 52,
              child: ListView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                children: [
                  _CategoryChip(
                    label: 'All', value: '', selected: _category == '',
                    onTap: () => _fetch(category: ''),
                  ),
                  ...cats.map((c) => _CategoryChip(
                    label: c['name'], value: c['slug'],
                    selected: _category == c['slug'],
                    onTap: () => _fetch(category: c['slug']),
                  )),
                ],
              ),
            ),
          ),

          Divider(height: 1, color: scheme.outlineVariant.withValues(alpha: 0.5)),

          // Services grid
          Expanded(
            child: services.when(
              loading: () => GridView.builder(
                padding: const EdgeInsets.all(16),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2, childAspectRatio: 0.72,
                  crossAxisSpacing: 12, mainAxisSpacing: 12,
                ),
                itemCount: 6,
                itemBuilder: (_, __) => const _SkeletonCard(),
              ),
              error: (e, _) => Center(
                child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                  Icon(Icons.error_outline_rounded, size: 48, color: scheme.error),
                  const SizedBox(height: 12),
                  Text('Failed to load', style: theme.textTheme.titleMedium),
                  const SizedBox(height: 8),
                  FilledButton.icon(
                    onPressed: () => ref.invalidate(_servicesProvider(_params)),
                    icon: const Icon(Icons.refresh_rounded),
                    label: const Text('Retry'),
                  ),
                ]),
              ),
              data: (data) {
                final list = data['data'] as List? ?? [];
                if (list.isEmpty) {
                  return Center(
                    child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                      Icon(Icons.search_off_rounded, size: 56,
                        color: scheme.onSurface.withValues(alpha: 0.3)),
                      const SizedBox(height: 16),
                      Text('No services found', style: theme.textTheme.titleMedium),
                      const SizedBox(height: 8),
                      Text('Try different keywords or filters',
                        style: theme.textTheme.bodySmall),
                      const SizedBox(height: 20),
                      OutlinedButton(
                        onPressed: _clearFilters,
                        child: const Text('Clear filters'),
                      ),
                    ]),
                  );
                }
                return RefreshIndicator(
                  onRefresh: () => ref.refresh(_servicesProvider(_params).future),
                  child: GridView.builder(
                    padding: const EdgeInsets.all(16),
                    gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                      crossAxisCount: 2, childAspectRatio: 0.72,
                      crossAxisSpacing: 12, mainAxisSpacing: 12,
                    ),
                    itemCount: list.length,
                    itemBuilder: (ctx, i) => ServiceCard(
                      service: list[i],
                      onTap: () => ctx.push('/marketplace/gig/${list[i]['slug']}'),
                    ),
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  void _showSort() {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => Padding(
        padding: const EdgeInsets.fromLTRB(16, 20, 16, 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Sort by',
              style: GoogleFonts.inter(fontSize: 16, fontWeight: FontWeight.w700)),
            const SizedBox(height: 12),
            ..._sortOptions.map((opt) => ListTile(
              leading: Icon(opt.$3,
                color: _sort == opt.$1
                  ? Theme.of(context).colorScheme.primary : null),
              title: Text(opt.$2),
              trailing: _sort == opt.$1
                ? Icon(Icons.check_rounded,
                    color: Theme.of(context).colorScheme.primary)
                : null,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12)),
              onTap: () {
                Navigator.pop(context);
                _fetch(sort: opt.$1);
              },
            )),
          ],
        ),
      ),
    );
  }
}

class _CategoryChip extends StatelessWidget {
  final String label, value;
  final bool selected;
  final VoidCallback onTap;
  const _CategoryChip({
    required this.label, required this.value,
    required this.selected, required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: GestureDetector(
        onTap: onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
          decoration: BoxDecoration(
            color: selected ? scheme.primary : scheme.surfaceContainerHighest,
            borderRadius: BorderRadius.circular(20),
          ),
          child: Text(
            label,
            style: GoogleFonts.inter(
              fontSize: 13,
              fontWeight: selected ? FontWeight.w600 : FontWeight.w500,
              color: selected ? Colors.white
                : scheme.onSurface.withValues(alpha: 0.7),
            ),
          ),
        ),
      ),
    );
  }
}

class _SkeletonCard extends StatelessWidget {
  const _SkeletonCard();

  @override
  Widget build(BuildContext context) {
    final c = Theme.of(context).colorScheme.surfaceContainerHighest;
    return Card(
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(flex: 5, child: Container(color: c)),
          Expanded(
            flex: 6,
            child: Padding(
              padding: const EdgeInsets.all(10),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(height: 10, width: 60,
                    decoration: BoxDecoration(color: c,
                      borderRadius: BorderRadius.circular(4))),
                  const SizedBox(height: 8),
                  Container(height: 12,
                    decoration: BoxDecoration(color: c,
                      borderRadius: BorderRadius.circular(4))),
                  const SizedBox(height: 4),
                  Container(height: 12, width: 100,
                    decoration: BoxDecoration(color: c,
                      borderRadius: BorderRadius.circular(4))),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
