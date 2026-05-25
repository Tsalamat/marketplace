import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../../core/network/api_client.dart';

const _storage = FlutterSecureStorage(
  aOptions: AndroidOptions(encryptedSharedPreferences: true),
);

final _myProfileProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  final res = await ref.read(dioProvider).get('/api/v1/users/me');
  return res.data as Map<String, dynamic>;
});

class MyProfileScreen extends ConsumerWidget {
  const MyProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profile = ref.watch(_myProfileProvider);
    final scheme = Theme.of(context).colorScheme;

    return Scaffold(
      body: profile.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => _NotLoggedIn(),
        data: (data) {
          final p = (data['profile'] as Map?) ?? {};
          final username = data['username'] as String? ?? '';
          final firstName = p['first_name'] as String? ?? '';
          final lastName = p['last_name'] as String? ?? '';
          final displayName = '$firstName $lastName'.trim().isNotEmpty ? '$firstName $lastName'.trim() : username;
          final avatarUrl = p['avatar_url'] as String? ??
              'https://ui-avatars.com/api/?name=$username&size=120&background=2563eb&color=fff';
          final bio = p['bio'] as String? ?? '';
          final university = p['university'] as String? ?? '';
          final rating = (p['rating'] as num? ?? 0).toStringAsFixed(1);
          final jobsDone = p['completed_jobs'] as int? ?? 0;
          final skills = (p['skills'] as List? ?? []).cast<String>();

          return CustomScrollView(
            slivers: [
              SliverAppBar(
                expandedHeight: 200,
                pinned: true,
                backgroundColor: scheme.primary,
                flexibleSpace: FlexibleSpaceBar(
                  background: Container(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft, end: Alignment.bottomRight,
                        colors: [scheme.primary, scheme.primary.withValues(alpha: 0.7)],
                      ),
                    ),
                    child: SafeArea(
                      child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                        const SizedBox(height: 20),
                        CircleAvatar(
                          radius: 44,
                          backgroundColor: Colors.white24,
                          child: CircleAvatar(
                            radius: 41,
                            backgroundImage: CachedNetworkImageProvider(avatarUrl),
                          ),
                        ),
                        const SizedBox(height: 10),
                        Text(displayName, style: const TextStyle(color: Colors.white, fontSize: 20, fontWeight: FontWeight.w700)),
                        Text('@$username', style: const TextStyle(color: Colors.white70, fontSize: 13)),
                      ]),
                    ),
                  ),
                ),
                actions: [
                  IconButton(
                    icon: const Icon(Icons.edit_outlined, color: Colors.white),
                    onPressed: () => context.push('/settings'),
                    tooltip: 'Редактировать профиль',
                  ),
                ],
              ),

              SliverPadding(
                padding: const EdgeInsets.all(16),
                sliver: SliverList(
                  delegate: SliverChildListDelegate([

                    // Статистика
                    Row(children: [
                      _StatCard(value: rating, label: 'Рейтинг', icon: Icons.star_rounded, color: Colors.amber),
                      const SizedBox(width: 10),
                      _StatCard(value: '$jobsDone', label: 'Выполнено', icon: Icons.check_circle_outline, color: Colors.green),
                      const SizedBox(width: 10),
                      _StatCard(value: university.isNotEmpty ? 'Set' : '—', label: university.isNotEmpty ? university.split(' ').first : 'Универ', icon: Icons.school_outlined, color: scheme.primary),
                    ]),

                    if (bio.isNotEmpty) ...[
                      const SizedBox(height: 16),
                      Card(
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                        elevation: 0,
                        color: scheme.surfaceContainerHighest,
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                            const Text('О себе', style: TextStyle(fontWeight: FontWeight.w700, fontSize: 14)),
                            const SizedBox(height: 6),
                            Text(bio, style: TextStyle(color: scheme.onSurface.withValues(alpha: 0.8), height: 1.5)),
                          ]),
                        ),
                      ),
                    ],

                    if (skills.isNotEmpty) ...[
                      const SizedBox(height: 12),
                      const Text('Навыки', style: TextStyle(fontWeight: FontWeight.w700, fontSize: 14)),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 6, runSpacing: 6,
                        children: skills.map((s) => Chip(
                          label: Text(s, style: const TextStyle(fontSize: 12)),
                          visualDensity: VisualDensity.compact,
                          padding: EdgeInsets.zero,
                        )).toList(),
                      ),
                    ],

                    const SizedBox(height: 20),
                    const Divider(),
                    const SizedBox(height: 8),

                    // Меню
                    _MenuItem(icon: Icons.storefront_outlined, label: 'Мои объявления', onTap: () => context.go('/marketplace')),
                    _MenuItem(icon: Icons.receipt_long_outlined, label: 'Мои заказы', onTap: () => context.go('/orders')),
                    _MenuItem(icon: Icons.notifications_outlined, label: 'Уведомления', onTap: () => context.go('/notifications')),
                    _MenuItem(icon: Icons.settings_outlined, label: 'Настройки', onTap: () => context.push('/settings')),

                    const SizedBox(height: 8),
                    const Divider(),
                    const SizedBox(height: 8),

                    ListTile(
                      leading: Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: Colors.red.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: const Icon(Icons.logout_rounded, color: Colors.red, size: 20),
                      ),
                      title: const Text('Выйти из аккаунта', style: TextStyle(color: Colors.red, fontWeight: FontWeight.w600)),
                      onTap: () => _logout(context),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),

                    const SizedBox(height: 40),
                  ]),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<void> _logout(BuildContext context) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Выйти?'),
        content: const Text('Вы уверены, что хотите выйти из аккаунта?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Отмена')),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Выйти'),
          ),
        ],
      ),
    );
    if (ok != true) return;
    await _storage.deleteAll();
    if (context.mounted) context.go('/login');
  }
}

class _NotLoggedIn extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
          const Icon(Icons.person_outline, size: 80, color: Colors.grey),
          const SizedBox(height: 16),
          const Text('Вы не вошли в аккаунт', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          const Text('Войдите, чтобы видеть профиль', style: TextStyle(color: Colors.grey)),
          const SizedBox(height: 24),
          FilledButton.icon(
            icon: const Icon(Icons.login),
            label: const Text('Войти'),
            onPressed: () => context.go('/login'),
          ),
        ]),
      ),
    );
  }
}

class _StatCard extends StatelessWidget {
  final String value;
  final String label;
  final IconData icon;
  final Color color;
  const _StatCard({required this.value, required this.label, required this.icon, required this.color});

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 10),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: color.withValues(alpha: 0.2)),
        ),
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          Icon(icon, color: color, size: 22),
          const SizedBox(height: 4),
          Text(value, style: TextStyle(fontWeight: FontWeight.w800, fontSize: 16, color: color)),
          Text(label, style: const TextStyle(fontSize: 10, color: Colors.grey), textAlign: TextAlign.center, maxLines: 2),
        ]),
      ),
    );
  }
}

class _MenuItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;
  const _MenuItem({required this.icon, required this.label, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surfaceContainerHighest,
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(icon, size: 20),
      ),
      title: Text(label, style: const TextStyle(fontWeight: FontWeight.w500)),
      trailing: const Icon(Icons.chevron_right, size: 18),
      onTap: onTap,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
    );
  }
}
