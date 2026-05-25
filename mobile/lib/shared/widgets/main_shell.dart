import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class MainShell extends StatelessWidget {
  final Widget child;
  final String currentLocation;
  const MainShell({
    super.key,
    required this.child,
    required this.currentLocation,
  });

  static const _tabs = [
    (icon: Icons.storefront_outlined,         activeIcon: Icons.storefront,              label: 'Маркет',   path: '/marketplace'),
    (icon: Icons.chat_bubble_outline_rounded, activeIcon: Icons.chat_bubble_rounded,     label: 'Чат',      path: '/chat'),
    (icon: Icons.people_outline_rounded,      activeIcon: Icons.people_rounded,           label: 'Друзья',   path: '/friends'),
    (icon: Icons.map_outlined,               activeIcon: Icons.map_rounded,              label: 'Карта',    path: '/map'),
    (icon: Icons.person_outline_rounded,      activeIcon: Icons.person_rounded,           label: 'Профиль',  path: '/my-profile'),
  ];

  int _currentIndex() {
    for (var i = 0; i < _tabs.length; i++) {
      if (currentLocation.startsWith(_tabs[i].path)) return i;
    }
    return 0;
  }

  @override
  Widget build(BuildContext context) {
    final idx    = _currentIndex();
    final scheme = Theme.of(context).colorScheme;

    return Scaffold(
      body: child,
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          color: Theme.of(context).appBarTheme.backgroundColor ?? scheme.surface,
          border: Border(top: BorderSide(color: scheme.outlineVariant.withValues(alpha: 0.4))),
          boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8, offset: const Offset(0, -2))],
        ),
        child: SafeArea(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 6),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: List.generate(_tabs.length, (i) {
                final tab      = _tabs[i];
                final selected = idx == i;
                return Expanded(
                  child: GestureDetector(
                    onTap: () => context.go(tab.path),
                    behavior: HitTestBehavior.opaque,
                    child: AnimatedContainer(
                      duration: const Duration(milliseconds: 200),
                      padding: const EdgeInsets.symmetric(vertical: 6),
                      decoration: BoxDecoration(
                        color: selected ? scheme.primary.withValues(alpha: 0.1) : Colors.transparent,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            selected ? tab.activeIcon : tab.icon,
                            color: selected ? scheme.primary : scheme.onSurface.withValues(alpha: 0.45),
                            size: 22,
                          ),
                          const SizedBox(height: 3),
                          Text(
                            tab.label,
                            style: TextStyle(
                              fontSize: 10,
                              fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
                              color: selected ? scheme.primary : scheme.onSurface.withValues(alpha: 0.45),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                );
              }),
            ),
          ),
        ),
      ),
    );
  }
}
