import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:timeago/timeago.dart' as timeago;
import '../../../core/network/api_client.dart';

final _ordersProvider = FutureProvider.family<List<dynamic>, String>((ref, role) async {
  final res = await ref.read(dioProvider).get('/api/v1/orders', queryParameters: {'as': role});
  return res.data as List;
});

class OrdersScreen extends ConsumerStatefulWidget {
  const OrdersScreen({super.key});

  @override
  ConsumerState<OrdersScreen> createState() => _OrdersScreenState();
}

class _OrdersScreenState extends ConsumerState<OrdersScreen> with SingleTickerProviderStateMixin {
  late final TabController _tabs = TabController(length: 2, vsync: this);

  @override
  void dispose() { _tabs.dispose(); super.dispose(); }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Orders'),
        bottom: TabBar(
          controller: _tabs,
          tabs: const [Tab(text: 'As Buyer'), Tab(text: 'As Seller')],
        ),
      ),
      body: TabBarView(
        controller: _tabs,
        children: const [_OrderList(role: 'buyer'), _OrderList(role: 'seller')],
      ),
    );
  }
}

class _OrderList extends ConsumerWidget {
  final String role;
  const _OrderList({required this.role});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final orders = ref.watch(_ordersProvider(role));

    return orders.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (e, _) => Center(child: Text('Error: $e')),
      data: (list) {
        if (list.isEmpty) {
          return Center(
            child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
              Icon(Icons.inbox_outlined, size: 56, color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.25)),
              const SizedBox(height: 12),
              Text('No orders yet', style: Theme.of(context).textTheme.titleMedium),
              if (role == 'buyer') ...[
                const SizedBox(height: 8),
                FilledButton(onPressed: () => context.go('/marketplace'), child: const Text('Browse Services')),
              ],
            ]),
          );
        }
        return RefreshIndicator(
          onRefresh: () => ref.refresh(_ordersProvider(role).future),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: list.length,
            itemBuilder: (ctx, i) => _OrderCard(order: list[i], onTap: () => ctx.push('/orders/${list[i]['id']}')),
          ),
        );
      },
    );
  }
}

class _OrderCard extends StatelessWidget {
  final Map<String, dynamic> order;
  final VoidCallback onTap;
  const _OrderCard({required this.order, required this.onTap});

  static const _statusColors = {
    'pending':     Colors.orange,
    'in_progress': Colors.blue,
    'delivered':   Colors.purple,
    'revision':    Colors.amber,
    'completed':   Colors.green,
    'cancelled':   Colors.red,
    'disputed':    Colors.red,
  };

  @override
  Widget build(BuildContext context) {
    final status = order['status'] as String? ?? 'pending';
    final color  = _statusColors[status] ?? Colors.grey;
    final theme  = Theme.of(context);

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        onTap: onTap,
        contentPadding: const EdgeInsets.all(16),
        title: Text(order['service']?['title'] ?? 'Order', style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14), maxLines: 1, overflow: TextOverflow.ellipsis),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Text('\$${(order['amount'] as num? ?? 0).toStringAsFixed(0)} · ${order['package']?['name'] ?? ''}'.toUpperCase(),
              style: theme.textTheme.bodySmall),
            const SizedBox(height: 6),
            Text(
              order['created_at'] != null ? timeago.format(DateTime.parse(order['created_at'])) : '',
              style: theme.textTheme.bodySmall,
            ),
          ],
        ),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(20)),
              child: Text(status.replaceAll('_', ' '), style: TextStyle(fontSize: 11, fontWeight: FontWeight.w700, color: color)),
            ),
            const SizedBox(height: 4),
            const Icon(Icons.chevron_right, size: 18),
          ],
        ),
      ),
    );
  }
}
