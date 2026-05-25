import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../../core/network/api_client.dart';

final _friendsProvider = FutureProvider<List<dynamic>>((ref) async {
  final res = await ref.read(dioProvider).get('/api/v1/friends');
  return res.data as List? ?? [];
});

final _requestsProvider = FutureProvider<List<dynamic>>((ref) async {
  final res = await ref.read(dioProvider).get('/api/v1/friends/requests');
  return res.data as List? ?? [];
});

class FriendsScreen extends ConsumerStatefulWidget {
  const FriendsScreen({super.key});
  @override
  ConsumerState<FriendsScreen> createState() => _FriendsScreenState();
}

class _FriendsScreenState extends ConsumerState<FriendsScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tab;
  final _searchCtrl = TextEditingController();
  final _peopleCtrl = TextEditingController();
  List<dynamic> _peopleResults = [];
  bool _searchingPeople = false;
  String _lastQuery = '';
  final _pendingRequests = <String>{};
  final _sentRequests    = <String>{};

  @override
  void initState() {
    super.initState();
    _tab = TabController(length: 3, vsync: this);
    _peopleCtrl.addListener(_onPeopleSearch);
  }

  @override
  void dispose() {
    _tab.dispose();
    _searchCtrl.dispose();
    _peopleCtrl.dispose();
    super.dispose();
  }

  Future<void> _onPeopleSearch() async {
    final q = _peopleCtrl.text.trim();
    if (q == _lastQuery) return;
    _lastQuery = q;
    if (q.isEmpty) { setState(() => _peopleResults = []); return; }
    setState(() => _searchingPeople = true);
    try {
      final res = await ref.read(dioProvider).get('/api/v1/users/search', queryParameters: {'q': q, 'limit': '20'});
      if (mounted && _lastQuery == q) setState(() => _peopleResults = res.data as List? ?? []);
    } catch (_) {
    } finally {
      if (mounted) setState(() => _searchingPeople = false);
    }
  }

  Future<void> _loadAllPeople() async {
    setState(() => _searchingPeople = true);
    try {
      final res = await ref.read(dioProvider).get('/api/v1/users/search', queryParameters: {'q': '', 'limit': '30'});
      if (mounted) setState(() => _peopleResults = res.data as List? ?? []);
    } catch (_) {
    } finally {
      if (mounted) setState(() => _searchingPeople = false);
    }
  }

  Future<void> _sendRequest(String userId) async {
    if (_pendingRequests.contains(userId) || _sentRequests.contains(userId)) return;
    setState(() => _pendingRequests.add(userId));
    try {
      await ref.read(dioProvider).post('/api/v1/friends/$userId');
      if (mounted) {
        setState(() { _sentRequests.add(userId); _pendingRequests.remove(userId); });
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Заявка отправлена!')));
      }
    } catch (_) {
      if (mounted) setState(() => _pendingRequests.remove(userId));
    }
  }

  Future<void> _acceptRequest(String id) async {
    try {
      await ref.read(dioProvider).put('/api/v1/friends/$id/accept');
      ref.invalidate(_requestsProvider);
      ref.invalidate(_friendsProvider);
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Заявка принята!')));
    } catch (_) {}
  }

  Future<void> _rejectRequest(String id) async {
    try {
      await ref.read(dioProvider).put('/api/v1/friends/$id/reject');
      ref.invalidate(_requestsProvider);
    } catch (_) {}
  }

  Future<void> _removeFriend(String userId) async {
    final ok = await showDialog<bool>(context: context,
      builder: (_) => AlertDialog(
        title: const Text('Удалить из друзей?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Отмена')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Удалить')),
        ],
      ),
    );
    if (ok != true) return;
    try {
      await ref.read(dioProvider).delete('/api/v1/friends/$userId');
      ref.invalidate(_friendsProvider);
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Удалён из друзей')));
    } catch (_) {}
  }

  String _avatar(Map? user) {
    final url = user?['profile']?['avatar_url'] as String?;
    return url?.isNotEmpty == true
        ? url!
        : 'https://ui-avatars.com/api/?name=${user?['username'] ?? '?'}&size=48&background=2563eb&color=fff';
  }

  String _name(Map? user) {
    final p = user?['profile'] as Map?;
    final fn = p?['first_name'] as String? ?? '';
    final ln = p?['last_name'] as String? ?? '';
    final full = '$fn $ln'.trim();
    return full.isNotEmpty ? full : (user?['username'] as String? ?? '');
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final requests = ref.watch(_requestsProvider);
    final pendingCount = requests.maybeWhen(data: (l) => l.length, orElse: () => 0);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Друзья'),
        bottom: TabBar(
          controller: _tab,
          tabs: [
            const Tab(text: 'Друзья'),
            Tab(child: Row(mainAxisSize: MainAxisSize.min, children: [
              const Text('Заявки'),
              if (pendingCount > 0) ...[
                const SizedBox(width: 6),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                  decoration: BoxDecoration(color: scheme.error, borderRadius: BorderRadius.circular(10)),
                  child: Text('$pendingCount', style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.bold)),
                ),
              ],
            ])),
            const Tab(text: 'Найти'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tab,
        children: [
          _buildFriendsList(scheme),
          _buildRequests(),
          _buildPeopleSearch(scheme),
        ],
      ),
    );
  }

  // ─── Вкладка Друзья ───────────────────────────────────────
  Widget _buildFriendsList(ColorScheme scheme) {
    final friends = ref.watch(_friendsProvider);
    return Column(children: [
      Padding(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
        child: TextField(
          controller: _searchCtrl,
          onChanged: (_) => setState(() {}),
          decoration: InputDecoration(
            hintText: 'Поиск в друзьях...',
            prefixIcon: const Icon(Icons.search, size: 20),
            isDense: true,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
            filled: true,
            fillColor: scheme.surfaceContainerHighest,
            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          ),
        ),
      ),
      Expanded(
        child: friends.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (_, __) => const Center(child: Text('Ошибка загрузки')),
          data: (list) {
            final q = _searchCtrl.text.toLowerCase();
            final filtered = list.where((f) {
              if (q.isEmpty) return true;
              final user = f['user'] as Map?;
              return _name(user).toLowerCase().contains(q) ||
                  (user?['username'] as String? ?? '').toLowerCase().contains(q);
            }).toList();

            if (filtered.isEmpty) {
              return Center(
                child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                  Icon(Icons.people_outline, size: 56, color: scheme.outline),
                  const SizedBox(height: 12),
                  const Text('Нет друзей', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 4),
                  TextButton(onPressed: () { _tab.animateTo(2); _loadAllPeople(); }, child: const Text('Найти людей')),
                ]),
              );
            }

            return ListView.builder(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
              itemCount: filtered.length,
              itemBuilder: (_, i) {
                final f = filtered[i] as Map<String, dynamic>;
                final user = f['user'] as Map?;
                final profile = (user?['profile'] as Map?) ?? {};
                final isOnline = profile['is_online'] as bool? ?? false;
                return Card(
                  margin: const EdgeInsets.only(bottom: 8),
                  elevation: 0,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14),
                      side: BorderSide(color: scheme.outlineVariant.withValues(alpha: 0.5))),
                  child: ListTile(
                    contentPadding: const EdgeInsets.fromLTRB(12, 6, 8, 6),
                    leading: Stack(clipBehavior: Clip.none, children: [
                      CircleAvatar(radius: 24, backgroundImage: CachedNetworkImageProvider(_avatar(user))),
                      if (isOnline) Positioned(right: 0, bottom: 0,
                        child: Container(width: 12, height: 12,
                          decoration: BoxDecoration(color: Colors.green,
                            border: Border.all(color: scheme.surface, width: 2), shape: BoxShape.circle))),
                    ]),
                    title: Text(_name(user), style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                    subtitle: Text('@${user?['username'] ?? ''}${profile['tagline'] != null && (profile['tagline'] as String).isNotEmpty ? ' · ${profile['tagline']}' : ''}',
                        maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 12)),
                    trailing: Row(mainAxisSize: MainAxisSize.min, children: [
                      IconButton(
                        icon: Icon(Icons.chat_bubble_outline_rounded, color: scheme.primary, size: 20),
                        tooltip: 'Написать',
                        onPressed: () {
                          final uid = user?['id'] as String? ?? '';
                          if (uid.isNotEmpty) {
                            context.push('/chat/direct/$uid');
                          }
                        },
                      ),
                      IconButton(
                        icon: Icon(Icons.person_remove_outlined, color: scheme.error, size: 20),
                        tooltip: 'Удалить',
                        onPressed: () => _removeFriend(user?['id'] as String? ?? ''),
                      ),
                    ]),
                    onTap: () => context.push('/profile/${user?['username']}'),
                  ),
                );
              },
            );
          },
        ),
      ),
    ]);
  }

  // ─── Вкладка Заявки ───────────────────────────────────────
  Widget _buildRequests() {
    final requests = ref.watch(_requestsProvider);
    return requests.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, __) => const Center(child: Text('Ошибка')),
      data: (list) {
        if (list.isEmpty) {
          return Center(
            child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
              Icon(Icons.inbox_outlined, size: 56, color: Theme.of(context).colorScheme.outline),
              const SizedBox(height: 12),
              const Text('Нет входящих заявок', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            ]),
          );
        }
        final scheme = Theme.of(context).colorScheme;
        return ListView.builder(
          padding: const EdgeInsets.all(16),
          itemCount: list.length,
          itemBuilder: (_, i) {
            final r = list[i] as Map<String, dynamic>;
            final requester = r['requester'] as Map?;
            return Card(
              margin: const EdgeInsets.only(bottom: 10),
              elevation: 0,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14),
                  side: BorderSide(color: scheme.outlineVariant.withValues(alpha: 0.5))),
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Row(children: [
                  CircleAvatar(radius: 24, backgroundImage: CachedNetworkImageProvider(_avatar(requester))),
                  const SizedBox(width: 12),
                  Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text(_name(requester), style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                    Text('@${requester?['username'] ?? ''}',
                        style: TextStyle(fontSize: 12, color: scheme.onSurface.withValues(alpha: 0.6))),
                  ])),
                  IconButton(icon: Icon(Icons.close_rounded, color: scheme.error), onPressed: () => _rejectRequest(r['id'] as String)),
                  FilledButton(
                    onPressed: () => _acceptRequest(r['id'] as String),
                    style: FilledButton.styleFrom(padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8), minimumSize: Size.zero),
                    child: const Text('Принять', style: TextStyle(fontSize: 12)),
                  ),
                ]),
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildAddButton(String userId) {
    if (_sentRequests.contains(userId)) {
      return const Icon(Icons.check_circle_rounded, color: Colors.green, size: 24);
    }
    final loading = _pendingRequests.contains(userId);
    return FilledButton.tonal(
      onPressed: loading ? null : () => _sendRequest(userId),
      style: FilledButton.styleFrom(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        minimumSize: Size.zero,
      ),
      child: loading
          ? const SizedBox(width: 14, height: 14, child: CircularProgressIndicator(strokeWidth: 2))
          : const Text('Добавить', style: TextStyle(fontSize: 12)),
    );
  }

  // ─── Вкладка Найти людей ──────────────────────────────────
  Widget _buildPeopleSearch(ColorScheme scheme) {
    return Column(children: [
      Padding(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
        child: TextField(
          controller: _peopleCtrl,
          decoration: InputDecoration(
            hintText: 'Поиск по имени или нику...',
            prefixIcon: const Icon(Icons.search, size: 20),
            suffixIcon: _searchingPeople
                ? const Padding(padding: EdgeInsets.all(12), child: SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2)))
                : null,
            isDense: true,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide.none),
            filled: true,
            fillColor: scheme.surfaceContainerHighest,
            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          ),
        ),
      ),
      if (_peopleResults.isEmpty && !_searchingPeople)
        Expanded(child: Center(
          child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
            Icon(Icons.group_add_outlined, size: 56, color: scheme.outline),
            const SizedBox(height: 12),
            const Text('Введите имя или @никнейм', style: TextStyle(color: Colors.grey)),
            const SizedBox(height: 12),
            OutlinedButton.icon(
              icon: const Icon(Icons.explore, size: 18),
              label: const Text('Показать всех студентов'),
              onPressed: _loadAllPeople,
            ),
          ]),
        ))
      else
        Expanded(
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
            itemCount: _peopleResults.length,
            itemBuilder: (_, i) {
              final u = _peopleResults[i] as Map<String, dynamic>;
              final profile = (u['profile'] as Map?) ?? {};
              return Card(
                margin: const EdgeInsets.only(bottom: 8),
                elevation: 0,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14),
                    side: BorderSide(color: scheme.outlineVariant.withValues(alpha: 0.5))),
                child: ListTile(
                  contentPadding: const EdgeInsets.fromLTRB(12, 6, 8, 6),
                  leading: CircleAvatar(radius: 22, backgroundImage: CachedNetworkImageProvider(_avatar(u))),
                  title: Text(_name(u), style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                  subtitle: Text(
                    '@${u['username'] ?? ''}${(profile['university'] as String?)?.isNotEmpty == true ? ' · ${profile['university']}' : ''}',
                    maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 12),
                  ),
                  trailing: _buildAddButton(u['id'] as String? ?? ''),
                  onTap: () => context.push('/profile/${u['username']}'),
                ),
              );
            },
          ),
        ),
    ]);
  }
}
