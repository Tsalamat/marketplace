import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../../../core/network/api_client.dart';

// WebSocket URL is centralised in api_client.dart → wsBaseUrl

class _FriendLoc {
  final String userId;
  final String username;
  final String avatarUrl;
  double lat;
  double lng;
  _FriendLoc({required this.userId, required this.username, required this.avatarUrl, required this.lat, required this.lng});
}

class MapScreen extends ConsumerStatefulWidget {
  const MapScreen({super.key});
  @override
  ConsumerState<MapScreen> createState() => _MapScreenState();
}

class _MapScreenState extends ConsumerState<MapScreen> {
  final _mapCtrl = MapController();
  LatLng? _myPos;
  final Map<String, _FriendLoc> _friends = {};
  WebSocketChannel? _ws;
  StreamSubscription<Position>? _posSub;
  Timer? _pollTimer;
  bool _sharing = false;

  static const _defaultPos = LatLng(51.1694, 71.4491);

  @override
  void initState() {
    super.initState();
    _initLocation();
    _connectWS();
    _startPolling();
  }

  @override
  void dispose() {
    _posSub?.cancel();
    _ws?.sink.close();
    _pollTimer?.cancel();
    super.dispose();
  }

  Future<void> _initLocation() async {
    if (!await Geolocator.isLocationServiceEnabled()) return;
    var perm = await Geolocator.checkPermission();
    if (perm == LocationPermission.denied) perm = await Geolocator.requestPermission();
    if (perm == LocationPermission.denied || perm == LocationPermission.deniedForever) return;

    final pos = await Geolocator.getCurrentPosition();
    _onMyPos(pos);
    _posSub = Geolocator.getPositionStream(
      locationSettings: const LocationSettings(accuracy: LocationAccuracy.high, distanceFilter: 10),
    ).listen(_onMyPos);
  }

  void _onMyPos(Position pos) {
    if (!mounted) return;
    setState(() { _myPos = LatLng(pos.latitude, pos.longitude); _sharing = true; });
    _sendLocation(pos.latitude, pos.longitude);
  }

  Future<void> _sendLocation(double lat, double lng) async {
    try {
      await ref.read(dioProvider).patch('/api/v1/users/location', data: {'lat': lat, 'lng': lng});
      _ws?.sink.add(json.encode({'type': 'location', 'payload': {'lat': lat, 'lng': lng}}));
    } catch (_) {}
  }

  Future<void> _connectWS() async {
    final token = await readSecureValue('access_token');
    if (token == null) return;
    try {
      _ws = WebSocketChannel.connect(Uri.parse('$wsBaseUrl/ws/chat?token=$token'));
      await _ws!.ready;
      _ws!.stream.listen((raw) {
        try {
          final event = json.decode(raw as String) as Map<String, dynamic>;
          if (event['type'] == 'location') {
            final p = event['payload'] as Map<String, dynamic>;
            final uid = p['user_id'] as String?;
            final lat = (p['lat'] as num?)?.toDouble();
            final lng = (p['lng'] as num?)?.toDouble();
            if (uid != null && lat != null && lng != null && mounted) {
              setState(() {
                _friends[uid] = _FriendLoc(
                  userId: uid,
                  username: p['username'] as String? ?? 'Друг',
                  avatarUrl: p['avatar_url'] as String? ?? '',
                  lat: lat, lng: lng,
                );
              });
            }
          }
        } catch (_) {}
      });
    } catch (_) {}
  }

  void _startPolling() {
    _pollTimer = Timer.periodic(const Duration(seconds: 10), (_) async {
      try {
        final res = await ref.read(dioProvider).get('/api/v1/friends/locations');
        final locs = res.data as List? ?? [];
        if (!mounted) return;
        setState(() {
          for (final loc in locs) {
            final m = loc as Map<String, dynamic>;
            final uid = m['user_id'] as String?;
            final lat = (m['lat'] as num?)?.toDouble();
            final lng = (m['lng'] as num?)?.toDouble();
            if (uid != null && lat != null && lng != null) {
              _friends[uid] = _FriendLoc(
                userId: uid,
                username: m['username'] as String? ?? 'Друг',
                avatarUrl: m['avatar'] as String? ?? '',
                lat: lat, lng: lng,
              );
            }
          }
        });
      } catch (_) {}
    });
  }

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final center = _myPos ?? _defaultPos;

    final markers = <Marker>[
      if (_myPos != null)
        Marker(
          point: _myPos!,
          width: 40, height: 40,
          child: Container(
            decoration: BoxDecoration(
              color: scheme.primary,
              shape: BoxShape.circle,
              border: Border.all(color: Colors.white, width: 3),
              boxShadow: [BoxShadow(color: scheme.primary.withValues(alpha: 0.4), blurRadius: 8, spreadRadius: 2)],
            ),
            child: const Icon(Icons.person, color: Colors.white, size: 20),
          ),
        ),
      ..._friends.values.map((f) => Marker(
        point: LatLng(f.lat, f.lng),
        width: 50, height: 60,
        child: Column(mainAxisSize: MainAxisSize.min, children: [
          Container(
            width: 44, height: 44,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              border: Border.all(color: scheme.primary, width: 2),
              boxShadow: const [BoxShadow(color: Colors.black26, blurRadius: 6)],
            ),
            child: ClipOval(
              child: f.avatarUrl.isNotEmpty
                  ? CachedNetworkImage(imageUrl: f.avatarUrl, fit: BoxFit.cover)
                  : Container(color: scheme.primaryContainer,
                      child: Center(child: Text(f.username[0].toUpperCase(),
                          style: TextStyle(color: scheme.primary, fontWeight: FontWeight.bold)))),
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
            decoration: BoxDecoration(color: scheme.primary, borderRadius: BorderRadius.circular(4)),
            child: Text(f.username, style: const TextStyle(color: Colors.white, fontSize: 9, fontWeight: FontWeight.bold)),
          ),
        ]),
      )),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('Карта друзей'),
        actions: [
          if (_friends.isNotEmpty)
            TextButton.icon(
              icon: const Icon(Icons.people, size: 18),
              label: Text('${_friends.length}'),
              onPressed: _fitAll,
            ),
        ],
      ),
      body: Stack(children: [
        FlutterMap(
          mapController: _mapCtrl,
          options: MapOptions(initialCenter: center, initialZoom: 14),
          children: [
            TileLayer(
              urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
              userAgentPackageName: 'com.example.student_marketplace',
            ),
            MarkerLayer(markers: markers),
          ],
        ),

        // Статус
        Positioned(
          top: 12, left: 16, right: 16,
          child: Card(
            elevation: 3,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
              child: Row(children: [
                Container(
                  width: 10, height: 10,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: _sharing ? Colors.green : scheme.outline,
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  _sharing ? 'Геопозиция включена' : 'Геопозиция выключена',
                  style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500),
                ),
                const Spacer(),
                Text(
                  '${_friends.length} ${_friends.length == 1 ? 'друг' : 'друзей'} онлайн',
                  style: TextStyle(fontSize: 12, color: scheme.onSurface.withValues(alpha: 0.6)),
                ),
              ]),
            ),
          ),
        ),

        // FABs
        Positioned(
          bottom: 24, right: 16,
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            if (_friends.isNotEmpty) ...[
              FloatingActionButton.small(
                heroTag: 'fit',
                onPressed: _fitAll,
                tooltip: 'Показать всех',
                child: const Icon(Icons.people),
              ),
              const SizedBox(height: 8),
            ],
            FloatingActionButton(
              heroTag: 'me',
              onPressed: _sharing ? _centerMe : _initLocation,
              backgroundColor: _sharing ? scheme.primary : scheme.surfaceContainerHighest,
              foregroundColor: _sharing ? scheme.onPrimary : scheme.onSurface,
              tooltip: _sharing ? 'Моя позиция' : 'Включить геопозицию',
              child: Icon(_sharing ? Icons.my_location : Icons.location_off),
            ),
          ]),
        ),
      ]),
    );
  }

  void _centerMe() {
    if (_myPos != null) _mapCtrl.move(_myPos!, 15);
  }

  void _fitAll() {
    final pts = [if (_myPos != null) _myPos!, ..._friends.values.map((f) => LatLng(f.lat, f.lng))];
    if (pts.isEmpty) return;
    if (pts.length == 1) { _mapCtrl.move(pts.first, 14); return; }
    final lats = pts.map((p) => p.latitude);
    final lngs = pts.map((p) => p.longitude);
    final bounds = LatLngBounds(
      LatLng(lats.reduce((a, b) => a < b ? a : b), lngs.reduce((a, b) => a < b ? a : b)),
      LatLng(lats.reduce((a, b) => a > b ? a : b), lngs.reduce((a, b) => a > b ? a : b)),
    );
    _mapCtrl.fitCamera(CameraFit.bounds(bounds: bounds, padding: const EdgeInsets.all(60)));
  }
}
