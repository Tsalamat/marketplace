import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:timeago/timeago.dart' as timeago;
import '../../../core/network/api_client.dart';

final _notifsProvider = FutureProvider<List<dynamic>>((ref) async {
  final res = await ref.read(dioProvider).get('/api/v1/notifications');
  return res.data as List;
});

class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key});

  static const _icons = {
    'order':          (icon: Icons.receipt_long,       color: Color(0xFF2563EB)),
    'message':        (icon: Icons.chat_bubble,         color: Color(0xFF7C3AED)),
    'review':         (icon: Icons.star,                color: Color(0xFFF59E0B)),
    'follow':         (icon: Icons.person_add,          color: Color(0xFF059669)),
    'like':           (icon: Icons.favorite,            color: Color(0xFFEF4444)),
    'comment':        (icon: Icons.comment,             color: Color(0xFF0891B2)),
    'friend_request': (icon: Icons.people,              color: Color(0xFF7C3AED)),
    'system':         (icon: Icons.notifications,       color: Color(0xFF6B7280)),
  };

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifs = ref.watch(_notifsProvider);
    final theme  = Theme.of(context);
    final scheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Notifications'),
        actions: [
          TextButton(
            onPressed: () => _markAllRead(ref),
            child: const Text('Mark all read'),
          ),
        ],
      ),
      body: notifs.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
        data: (list) {
          if (list.isEmpty) {
            return Center(
              child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                Icon(Icons.notifications_none_rounded, size: 56, color: theme.colorScheme.onSurface.withValues(alpha: 0.25)),
                const SizedBox(height: 12),
                Text('No notifications yet', style: theme.textTheme.titleMedium),
                const SizedBox(height: 4),
                Text("You're all caught up!", style: theme.textTheme.bodySmall),
              ]),
            );
          }

          return RefreshIndicator(
            onRefresh: () => ref.refresh(_notifsProvider.future),
            child: ListView.separated(
              itemCount: list.length,
              separatorBuilder: (_, __) => const Divider(height: 1, indent: 72),
              itemBuilder: (ctx, i) {
                final n       = list[i] as Map<String, dynamic>;
                final type    = n['type'] as String? ?? 'system';
                final info    = _icons[type] ?? _icons['system']!;
                final isRead  = n['is_read'] as bool? ?? false;
                final timeStr = n['created_at'] != null
                    ? timeago.format(DateTime.parse(n['created_at']))
                    : '';

                return Dismissible(
                  key: ValueKey(n['id']),
                  direction: DismissDirection.endToStart,
                  background: Container(
                    color: scheme.error,
                    alignment: Alignment.centerRight,
                    padding: const EdgeInsets.only(right: 20),
                    child: const Icon(Icons.delete_outline, color: Colors.white),
                  ),
                  onDismissed: (_) => _dismiss(ref, n['id'] as String),
                  child: ListTile(
                    tileColor: isRead ? null : scheme.primary.withValues(alpha: 0.05),
                    leading: Container(
                      width: 44, height: 44,
                      decoration: BoxDecoration(
                        color: info.color.withValues(alpha: 0.12),
                        shape: BoxShape.circle,
                      ),
                      child: Icon(info.icon, color: info.color, size: 22),
                    ),
                    title: Text(
                      n['title'] ?? '',
                      style: TextStyle(fontWeight: isRead ? FontWeight.normal : FontWeight.w600, fontSize: 14),
                    ),
                    subtitle: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const SizedBox(height: 2),
                        Text(n['body'] ?? '', style: const TextStyle(fontSize: 13), maxLines: 2, overflow: TextOverflow.ellipsis),
                        const SizedBox(height: 4),
                        Text(timeStr, style: TextStyle(fontSize: 11, color: scheme.onSurface.withValues(alpha: 0.5))),
                      ],
                    ),
                    trailing: !isRead
                        ? Container(width: 8, height: 8, decoration: BoxDecoration(color: scheme.primary, shape: BoxShape.circle))
                        : null,
                    onTap: () => _markRead(ref, n['id'] as String),
                  ),
                );
              },
            ),
          );
        },
      ),
    );
  }

  Future<void> _markRead(WidgetRef ref, String id) async {
    try {
      await ref.read(dioProvider).patch('/api/v1/notifications/$id/read');
      ref.invalidate(_notifsProvider);
    } catch (_) {}
  }

  Future<void> _markAllRead(WidgetRef ref) async {
    try {
      await ref.read(dioProvider).post('/api/v1/notifications/read-all');
      ref.invalidate(_notifsProvider);
    } catch (_) {}
  }

  Future<void> _dismiss(WidgetRef ref, String id) async {
    try {
      await ref.read(dioProvider).delete('/api/v1/notifications/$id');
    } catch (_) {}
  }
}
