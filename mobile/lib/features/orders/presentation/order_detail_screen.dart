import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/network/api_client.dart';

final _orderDetailProvider = FutureProvider.family<Map<String, dynamic>, String>((ref, id) async {
  final res = await ref.read(dioProvider).get('/api/v1/orders/$id');
  return res.data as Map<String, dynamic>;
});

class OrderDetailScreen extends ConsumerWidget {
  final String orderId;
  const OrderDetailScreen({super.key, required this.orderId});

  static const _statusColors = {
    'pending':     Color(0xFFFF9800),
    'in_progress': Color(0xFF2196F3),
    'delivered':   Color(0xFF9C27B0),
    'revision':    Color(0xFFFFC107),
    'completed':   Color(0xFF4CAF50),
    'cancelled':   Color(0xFFF44336),
  };

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final order  = ref.watch(_orderDetailProvider(orderId));
    final theme  = Theme.of(context);
    final scheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text('Order Details')),
      body: order.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (o) {
          final status = o['status'] as String? ?? 'pending';
          final color  = _statusColors[status] ?? Colors.grey;

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              // Status banner
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: color.withValues(alpha: 0.3)),
                ),
                child: Row(children: [
                  Icon(Icons.circle, color: color, size: 10),
                  const SizedBox(width: 10),
                  Text(status.replaceAll('_', ' ').toUpperCase(),
                    style: TextStyle(color: color, fontWeight: FontWeight.w700, fontSize: 13)),
                  const Spacer(),
                  Text('\$${(o['amount'] as num? ?? 0).toStringAsFixed(2)}',
                    style: TextStyle(fontWeight: FontWeight.w800, fontSize: 18, color: scheme.primary)),
                ]),
              ),
              const SizedBox(height: 20),

              // Service info
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text('Service', style: theme.textTheme.labelLarge?.copyWith(color: scheme.onSurface.withValues(alpha: 0.5))),
                    const SizedBox(height: 6),
                    Text(o['service']?['title'] ?? '', style: theme.textTheme.titleMedium),
                    const SizedBox(height: 4),
                    Text('Package: ${o['package']?['name'] ?? ''}'.toUpperCase(),
                      style: theme.textTheme.bodySmall?.copyWith(color: scheme.primary, fontWeight: FontWeight.w600)),
                    const SizedBox(height: 4),
                    Text('${o['package']?['delivery_days'] ?? 0} day delivery · ${o['max_revisions']} revisions',
                      style: theme.textTheme.bodySmall),
                  ]),
                ),
              ),
              const SizedBox(height: 12),

              // Parties
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    _userRow('Buyer',  o['buyer'],  theme),
                    const Divider(height: 20),
                    _userRow('Seller', o['seller'], theme),
                  ]),
                ),
              ),
              const SizedBox(height: 12),

              // Requirements
              if ((o['requirements'] as String? ?? '').isNotEmpty)
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                      Text('Requirements', style: theme.textTheme.labelLarge?.copyWith(color: scheme.onSurface.withValues(alpha: 0.5))),
                      const SizedBox(height: 8),
                      Text(o['requirements'], style: theme.textTheme.bodyMedium?.copyWith(height: 1.5)),
                    ]),
                  ),
                ),

              const SizedBox(height: 20),

              // Financials
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(children: [
                    _priceRow('Order amount',   '\$${(o['amount']        as num? ?? 0).toStringAsFixed(2)}', theme),
                    _priceRow('Platform fee',   '\$${(o['platform_fee']  as num? ?? 0).toStringAsFixed(2)}', theme),
                    const Divider(height: 16),
                    _priceRow('Seller earns',   '\$${(o['seller_amount'] as num? ?? 0).toStringAsFixed(2)}', theme, bold: true),
                  ]),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _userRow(String label, Map? user, ThemeData theme) => Row(children: [
    CircleAvatar(
      radius: 20,
      backgroundImage: NetworkImage(
        user?['profile']?['avatar_url'] ??
        'https://ui-avatars.com/api/?name=${user?['username']}&size=40&background=2563eb&color=fff',
      ),
    ),
    const SizedBox(width: 12),
    Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text(label, style: theme.textTheme.bodySmall),
      Text(user?['username'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600)),
    ]),
  ]);

  Widget _priceRow(String label, String value, ThemeData theme, {bool bold = false}) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 4),
    child: Row(children: [
      Text(label, style: theme.textTheme.bodyMedium),
      const Spacer(),
      Text(value, style: TextStyle(fontWeight: bold ? FontWeight.w700 : FontWeight.normal)),
    ]),
  );
}
