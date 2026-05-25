import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../features/auth/presentation/login_screen.dart';
import '../../features/auth/presentation/register_screen.dart';
import '../../features/marketplace/presentation/marketplace_screen.dart';
import '../../features/marketplace/presentation/gig_detail_screen.dart';
import '../../features/orders/presentation/orders_screen.dart';
import '../../features/orders/presentation/order_detail_screen.dart';
import '../../features/chat/presentation/chat_screen.dart';
import '../../features/community/presentation/community_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/profile/presentation/my_profile_screen.dart';
import '../../features/notifications/presentation/notifications_screen.dart';
import '../../features/friends/presentation/friends_screen.dart';
import '../../features/map/presentation/map_screen.dart';
import '../../features/settings/presentation/settings_screen.dart';
import '../../shared/widgets/main_shell.dart';

const _storage = FlutterSecureStorage(
  aOptions: AndroidOptions(encryptedSharedPreferences: true),
);

Future<String?> _readAccessToken() async {
  try {
    return await _storage.read(key: 'access_token');
  } catch (_) {
    try {
      await _storage.deleteAll();
    } catch (_) {}
    return null;
  }
}

final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/marketplace',
    redirect: (context, state) async {
      final token = await _readAccessToken();
      final isAuth = token != null;
      final protectedPrefixes = ['/orders', '/chat', '/my-profile', '/friends', '/map', '/settings', '/notifications'];
      final isProtected = protectedPrefixes.any((r) => state.matchedLocation.startsWith(r));
      if (!isAuth && isProtected) return '/login?redirect=${state.uri}';
      if (isAuth && (state.matchedLocation == '/login' || state.matchedLocation == '/register')) {
        return '/marketplace';
      }
      return null;
    },
    routes: [
      GoRoute(path: '/login',    name: 'login',    builder: (_, __) => const LoginScreen()),
      GoRoute(path: '/register', name: 'register', builder: (_, __) => const RegisterScreen()),

      // OAuth deep-link callback: studentmarketplace:///oauth-callback?at=TOKEN&rt=TOKEN
      GoRoute(
        path: '/oauth-callback',
        name: 'oauth-callback',
        builder: (_, state) => _OAuthCallbackScreen(
          accessToken:  state.uri.queryParameters['at'],
          refreshToken: state.uri.queryParameters['rt'],
        ),
      ),

      ShellRoute(
        builder: (context, state, child) => MainShell(
          currentLocation: state.matchedLocation,
          child: child,
        ),
        routes: [
          GoRoute(path: '/marketplace', name: 'marketplace', builder: (_, __) => const MarketplaceScreen(),
            routes: [
              GoRoute(path: 'gig/:slug', name: 'gig-detail',
                builder: (_, state) => GigDetailScreen(slug: state.pathParameters['slug']!)),
            ]),
          GoRoute(path: '/orders', name: 'orders', builder: (_, __) => const OrdersScreen(),
            routes: [
              GoRoute(path: ':id', name: 'order-detail',
                builder: (_, state) => OrderDetailScreen(orderId: state.pathParameters['id']!)),
            ]),
          GoRoute(
            path: '/chat',
            name: 'chat',
            builder: (_, __) => const ChatScreen(),
            routes: [
              GoRoute(
                path: 'direct/:userId',
                name: 'direct-chat',
                builder: (_, state) => ChatScreen(
                  initialDirectUserId: state.pathParameters['userId'],
                ),
              ),
            ],
          ),
          GoRoute(path: '/community',     name: 'community',     builder: (_, __) => const CommunityScreen()),
          GoRoute(path: '/friends',       name: 'friends',       builder: (_, __) => const FriendsScreen()),
          GoRoute(path: '/map',           name: 'map',           builder: (_, __) => const MapScreen()),
          GoRoute(path: '/my-profile',    name: 'my-profile',    builder: (_, __) => const MyProfileScreen()),
          GoRoute(path: '/settings',      name: 'settings',      builder: (_, __) => const SettingsScreen()),
          GoRoute(path: '/notifications', name: 'notifications', builder: (_, __) => const NotificationsScreen()),
          GoRoute(path: '/profile/:username', name: 'profile',
            builder: (_, state) => ProfileScreen(username: state.pathParameters['username']!)),
        ],
      ),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
        const Text('404', style: TextStyle(fontSize: 48, fontWeight: FontWeight.bold)),
        const SizedBox(height: 8),
        Text('Страница не найдена: ${state.uri}'),
        const SizedBox(height: 16),
        ElevatedButton(onPressed: () => context.go('/marketplace'), child: const Text('На главную')),
      ])),
    ),
  );
});

// ─── OAuth Callback ───────────────────────────────────────────────────────────
// Opened when Android receives the deep link:
//   studentmarketplace:///oauth-callback?at=ACCESS_TOKEN&rt=REFRESH_TOKEN
class _OAuthCallbackScreen extends StatefulWidget {
  final String? accessToken;
  final String? refreshToken;
  const _OAuthCallbackScreen({this.accessToken, this.refreshToken});

  @override
  State<_OAuthCallbackScreen> createState() => _OAuthCallbackScreenState();
}

class _OAuthCallbackScreenState extends State<_OAuthCallbackScreen> {
  static const _sec = FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
  );

  @override
  void initState() {
    super.initState();
    _handleCallback();
  }

  Future<void> _handleCallback() async {
    final at = widget.accessToken;
    final rt = widget.refreshToken;
    if (at != null && at.isNotEmpty) {
      await _sec.write(key: 'access_token', value: at);
      if (rt != null && rt.isNotEmpty) {
        await _sec.write(key: 'refresh_token', value: rt);
      }
    }
    if (mounted) context.go('/marketplace');
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(child: CircularProgressIndicator()),
    );
  }
}
