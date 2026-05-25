import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:timeago/timeago.dart' as timeago;
import 'dart:convert';
import '../../../core/network/api_client.dart';
// WebSocket URL is centralised in api_client.dart → wsBaseUrl

final _chatsProvider = FutureProvider<List<dynamic>>((ref) async {
  final res = await ref.read(dioProvider).get('/api/v1/chat');
  return res.data as List;
});

final _messagesProvider = FutureProvider.family<List<dynamic>, String>((ref, chatId) async {
  final res = await ref.read(dioProvider).get('/api/v1/chat/$chatId/messages');
  return res.data as List;
});

class ChatScreen extends ConsumerStatefulWidget {
  /// When set, automatically opens (or creates) a direct chat with this user ID.
  final String? initialDirectUserId;
  const ChatScreen({super.key, this.initialDirectUserId});

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  Map<String, dynamic>? _activeChat;
  WebSocketChannel? _ws;
  final _messages = <Map<String, dynamic>>[];
  final _inputCtrl = TextEditingController();
  final _scrollCtrl = ScrollController();
  String? _myUserId;

  @override
  void initState() {
    super.initState();
    _connectWS();
    _loadUserId();
    if (widget.initialDirectUserId != null) {
      // Defer until first frame so the Dio provider is ready.
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _openDirectChat(widget.initialDirectUserId!);
      });
    }
  }

  Future<void> _loadUserId() async {
    final token = await readSecureValue('access_token');
    if (token == null) return;
    final parts = token.split('.');
    if (parts.length != 3) return;
    final payload = json.decode(utf8.decode(base64Url.decode(base64Url.normalize(parts[1]))));
    if (mounted) setState(() => _myUserId = payload['user_id'] as String?);
  }

  Future<void> _connectWS() async {
    if (!mounted) return;
    final token = await readSecureValue('access_token');
    if (token == null || !mounted) return;
    try {
      _ws?.sink.close();
      _ws = WebSocketChannel.connect(Uri.parse('$wsBaseUrl/ws/chat?token=$token'));
      await _ws!.ready;
      _ws!.stream.listen(
        _onWsEvent,
        onDone: () {
          if (mounted) Future.delayed(const Duration(seconds: 3), _connectWS);
        },
        onError: (_) {
          if (mounted) Future.delayed(const Duration(seconds: 3), _connectWS);
        },
        cancelOnError: true,
      );
    } catch (_) {
      if (mounted) Future.delayed(const Duration(seconds: 3), _connectWS);
    }
  }

  void _onWsEvent(dynamic data) {
    try {
      final event = json.decode(data as String) as Map<String, dynamic>;
      if (event['type'] == 'message' &&
          _activeChat != null &&
          event['chat_id'] == _activeChat!['id']) {
        if (mounted) {
          setState(() => _messages.add(event['payload'] as Map<String, dynamic>));
          _scrollToBottom();
        }
      }
    } catch (_) {}
  }

  void _sendMessage() {
    final content = _inputCtrl.text.trim();
    if (content.isEmpty || _activeChat == null || _ws == null) return;
    _ws!.sink.add(json.encode({'type': 'message', 'payload': {'chat_id': _activeChat!['id'], 'content': content}}));
    _inputCtrl.clear();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollCtrl.hasClients) {
        _scrollCtrl.animateTo(
          _scrollCtrl.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  Future<void> _openDirectChat(String userId) async {
    try {
      final res = await ref.read(dioProvider).post('/api/v1/chat/direct/$userId');
      if (mounted) await _openChat(res.data as Map<String, dynamic>);
    } catch (_) {}
  }

  Future<void> _openChat(Map<String, dynamic> chat) async {
    final msgs = await ref.read(_messagesProvider(chat['id'] as String).future);
    if (mounted) {
      setState(() { _activeChat = chat; _messages..clear()..addAll(msgs.cast()); });
      _scrollToBottom();
    }
  }

  @override
  void dispose() { _ws?.sink.close(); _inputCtrl.dispose(); _scrollCtrl.dispose(); super.dispose(); }

  @override
  Widget build(BuildContext context) {
    final chats  = ref.watch(_chatsProvider);
    final theme  = Theme.of(context);
    final scheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(title: _activeChat == null ? const Text('Messages') : _chatHeader(theme)),
      body: _activeChat == null
          ? chats.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(child: Text('$e')),
              data: (list) => list.isEmpty
                  ? const Center(child: Text('No conversations yet'))
                  : ListView.builder(
                      itemCount: list.length,
                      itemBuilder: (_, i) => _ChatTile(chat: list[i], myId: _myUserId, onTap: () => _openChat(list[i])),
                    ),
            )
          : Column(
              children: [
                Expanded(
                  child: ListView.builder(
                    controller: _scrollCtrl,
                    padding: const EdgeInsets.all(16),
                    itemCount: _messages.length,
                    itemBuilder: (_, i) => _MessageBubble(message: _messages[i], myId: _myUserId, scheme: scheme, theme: theme),
                  ),
                ),
                _buildInput(scheme),
              ],
            ),
    );
  }

  Widget _chatHeader(ThemeData theme) {
    final participants = _activeChat!['participants'] as List? ?? [];
    final other = participants.firstWhere((p) => p['user_id'] != _myUserId, orElse: () => participants.firstOrNull ?? {});
    final user = other['user'] as Map? ?? {};
    return Row(children: [
      CircleAvatar(radius: 16, backgroundImage: NetworkImage(
        user['profile']?['avatar_url'] ?? 'https://ui-avatars.com/api/?name=${user['username']}&size=32',
      )),
      const SizedBox(width: 8),
      Text(user['username'] ?? '', style: theme.textTheme.titleMedium),
    ]);
  }

  Widget _buildInput(ColorScheme scheme) => SafeArea(
    child: Container(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
      decoration: BoxDecoration(
        color: scheme.surface,
        border: Border(top: BorderSide(color: scheme.outlineVariant)),
      ),
      child: Row(children: [
        Expanded(child: TextField(
          controller: _inputCtrl,
          decoration: const InputDecoration(hintText: 'Type a message...', border: InputBorder.none, isDense: true),
          onSubmitted: (_) => _sendMessage(),
        )),
        IconButton(icon: Icon(Icons.send_rounded, color: scheme.primary), onPressed: _sendMessage),
      ]),
    ),
  );
}

class _ChatTile extends StatelessWidget {
  final Map<String, dynamic> chat;
  final String? myId;
  final VoidCallback onTap;
  const _ChatTile({required this.chat, required this.myId, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final participants = chat['participants'] as List? ?? [];
    final other = participants.firstWhere((p) => p['user_id'] != myId, orElse: () => participants.firstOrNull ?? {});
    final user  = (other['user'] as Map?) ?? {};
    final last  = chat['last_message'] as Map?;
    final unread = chat['unread_count'] as int? ?? 0;

    return ListTile(
      onTap: onTap,
      leading: CircleAvatar(
        backgroundImage: NetworkImage(user['profile']?['avatar_url'] ?? 'https://ui-avatars.com/api/?name=${user['username']}&size=44'),
      ),
      title: Text(user['username'] ?? '', style: const TextStyle(fontWeight: FontWeight.w600)),
      subtitle: Text(last?['content'] ?? 'No messages yet', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 13)),
      trailing: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
        if (last?['created_at'] != null)
          Text(timeago.format(DateTime.parse(last!['created_at'])), style: const TextStyle(fontSize: 11, color: Colors.grey)),
        if (unread > 0) ...[
          const SizedBox(height: 4),
          CircleAvatar(radius: 10, backgroundColor: Theme.of(context).colorScheme.primary,
            child: Text('$unread', style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.bold))),
        ],
      ]),
    );
  }
}

class _MessageBubble extends StatelessWidget {
  final Map<String, dynamic> message;
  final String? myId;
  final ColorScheme scheme;
  final ThemeData theme;
  const _MessageBubble({required this.message, required this.myId, required this.scheme, required this.theme});

  @override
  Widget build(BuildContext context) {
    final isMine = message['sender_id'] == myId;
    final time   = message['created_at'] != null ? timeago.format(DateTime.parse(message['created_at'])) : '';

    return Align(
      alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: 8),
        constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.72),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: isMine ? scheme.primary : scheme.surfaceContainerHighest,
          borderRadius: BorderRadius.only(
            topLeft:     const Radius.circular(18),
            topRight:    const Radius.circular(18),
            bottomLeft:  Radius.circular(isMine ? 18 : 4),
            bottomRight: Radius.circular(isMine ? 4  : 18),
          ),
        ),
        child: Column(
          crossAxisAlignment: isMine ? CrossAxisAlignment.end : CrossAxisAlignment.start,
          children: [
            Text(message['content'] ?? '', style: TextStyle(color: isMine ? Colors.white : scheme.onSurface, fontSize: 14, height: 1.4)),
            const SizedBox(height: 4),
            Text(time, style: TextStyle(fontSize: 10, color: (isMine ? Colors.white : scheme.onSurface).withValues(alpha: 0.55))),
          ],
        ),
      ),
    );
  }
}
