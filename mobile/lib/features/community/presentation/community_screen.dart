import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:timeago/timeago.dart' as timeago;
import '../../../core/network/api_client.dart';

final _feedProvider = FutureProvider<List<dynamic>>((ref) async {
  final res = await ref.read(dioProvider).get('/api/v1/posts', queryParameters: {'limit': 20});
  return res.data as List;
});

class CommunityScreen extends ConsumerStatefulWidget {
  const CommunityScreen({super.key});

  @override
  ConsumerState<CommunityScreen> createState() => _CommunityScreenState();
}

class _CommunityScreenState extends ConsumerState<CommunityScreen> {
  void _showCreatePost() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (_) => _CreatePostSheet(onCreated: () => ref.invalidate(_feedProvider)),
    );
  }

  @override
  Widget build(BuildContext context) {
    final feed = ref.watch(_feedProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Community'),
        actions: [
          IconButton(icon: const Icon(Icons.add_circle_outline), onPressed: _showCreatePost),
        ],
      ),
      body: feed.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
        data: (posts) => RefreshIndicator(
          onRefresh: () => ref.refresh(_feedProvider.future),
          child: ListView.separated(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: posts.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (_, i) => _PostCard(post: posts[i], onLike: () => _toggleLike(posts[i]['id'])),
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showCreatePost,
        child: const Icon(Icons.edit_outlined),
      ),
    );
  }

  Future<void> _toggleLike(String postId) async {
    try {
      await ref.read(dioProvider).post('/api/v1/posts/$postId/like');
      ref.invalidate(_feedProvider);
    } catch (_) {}
  }
}

class _PostCard extends StatelessWidget {
  final Map<String, dynamic> post;
  final VoidCallback onLike;
  const _PostCard({required this.post, required this.onLike});

  @override
  Widget build(BuildContext context) {
    final author  = post['author']  as Map? ?? {};
    final profile = author['profile'] as Map? ?? {};
    final images  = post['images'] as List? ?? [];
    final time    = post['created_at'] != null ? timeago.format(DateTime.parse(post['created_at'])) : '';
    final theme   = Theme.of(context);

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Row(children: [
            CircleAvatar(
              radius: 20,
              backgroundImage: NetworkImage(
                profile['avatar_url'] ?? 'https://ui-avatars.com/api/?name=${author['username']}&size=40',
              ),
            ),
            const SizedBox(width: 10),
            Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(author['username'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
              Text('${profile['university'] ?? ''} · $time', style: theme.textTheme.bodySmall),
            ]),
            const Spacer(),
            IconButton(icon: const Icon(Icons.more_horiz, size: 20), onPressed: () {}),
          ]),
          const SizedBox(height: 12),

          // Content
          Text(post['content'] ?? '', style: theme.textTheme.bodyMedium?.copyWith(height: 1.5)),

          // Images
          if (images.isNotEmpty) ...[
            const SizedBox(height: 10),
            ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: CachedNetworkImage(imageUrl: images[0] as String, fit: BoxFit.cover, height: 200, width: double.infinity),
            ),
          ],
          const SizedBox(height: 12),

          // Actions
          Row(children: [
            _ActionBtn(
              icon: post['is_liked'] == true ? Icons.favorite : Icons.favorite_border,
              color: post['is_liked'] == true ? Colors.red : null,
              label: '${post['likes_count'] ?? 0}',
              onTap: onLike,
            ),
            const SizedBox(width: 16),
            _ActionBtn(icon: Icons.chat_bubble_outline, label: '${post['comments_count'] ?? 0}', onTap: () {}),
            const SizedBox(width: 16),
            _ActionBtn(icon: Icons.share_outlined, label: 'Share', onTap: () {}),
          ]),
        ],
      ),
    );
  }
}

class _ActionBtn extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color? color;
  final VoidCallback onTap;
  const _ActionBtn({required this.icon, required this.label, required this.onTap, this.color});

  @override
  Widget build(BuildContext context) => GestureDetector(
    onTap: onTap,
    child: Row(children: [
      Icon(icon, size: 20, color: color ?? Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.6)),
      const SizedBox(width: 4),
      Text(label, style: Theme.of(context).textTheme.bodySmall),
    ]),
  );
}

class _CreatePostSheet extends ConsumerStatefulWidget {
  final VoidCallback onCreated;
  const _CreatePostSheet({required this.onCreated});

  @override
  ConsumerState<_CreatePostSheet> createState() => _CreatePostSheetState();
}

class _CreatePostSheetState extends ConsumerState<_CreatePostSheet> {
  final _ctrl = TextEditingController();
  bool _loading = false;

  @override
  void dispose() { _ctrl.dispose(); super.dispose(); }

  Future<void> _post() async {
    if (_ctrl.text.trim().isEmpty) return;
    setState(() => _loading = true);
    try {
      await ref.read(dioProvider).post('/api/v1/posts', data: {'content': _ctrl.text.trim()});
      widget.onCreated();
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Error: $e')));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.fromLTRB(16, 16, 16, MediaQuery.of(context).viewInsets.bottom + 16),
      child: Column(mainAxisSize: MainAxisSize.min, crossAxisAlignment: CrossAxisAlignment.stretch, children: [
        Row(children: [
          Text("What's on your mind?", style: Theme.of(context).textTheme.titleMedium),
          const Spacer(),
          TextButton(onPressed: _loading ? null : _post, child: _loading ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2)) : const Text('Post')),
        ]),
        const SizedBox(height: 12),
        TextField(controller: _ctrl, maxLines: 5, autofocus: true, decoration: const InputDecoration(hintText: 'Share something with the community...')),
      ]),
    );
  }
}
