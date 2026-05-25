import 'package:flutter/material.dart';

class CallScreen extends StatelessWidget {
  final String targetUserId;
  final String targetUsername;
  final String? targetAvatarUrl;
  final bool isCaller;

  const CallScreen({
    super.key,
    required this.targetUserId,
    required this.targetUsername,
    this.targetAvatarUrl,
    required this.isCaller,
  });

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text('Audio calls coming soon'),
      ),
    );
  }
}
