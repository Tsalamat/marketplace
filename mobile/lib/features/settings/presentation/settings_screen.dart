import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';

const _storage = FlutterSecureStorage(
  aOptions: AndroidOptions(encryptedSharedPreferences: true),
);

final _meProvider = FutureProvider<Map<String, dynamic>>((ref) async {
  final res = await ref.read(dioProvider).get('/api/v1/users/me');
  return res.data as Map<String, dynamic>;
});

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  final _firstNameCtrl   = TextEditingController();
  final _lastNameCtrl    = TextEditingController();
  final _taglineCtrl     = TextEditingController();
  final _bioCtrl         = TextEditingController();
  final _universityCtrl  = TextEditingController();
  final _departmentCtrl  = TextEditingController();
  final _locationCtrl    = TextEditingController();
  final _skillsCtrl      = TextEditingController();
  final _languagesCtrl   = TextEditingController();
  final _githubCtrl      = TextEditingController();
  final _linkedinCtrl    = TextEditingController();
  final _portfolioCtrl   = TextEditingController();
  final _currentPassCtrl = TextEditingController();
  final _newPassCtrl     = TextEditingController();

  String? _avatarUrl;
  bool _uploadingAvatar = false;
  bool _saving = false;
  bool _changingPass = false;
  bool _initialized = false;

  @override
  void dispose() {
    _firstNameCtrl.dispose(); _lastNameCtrl.dispose(); _taglineCtrl.dispose();
    _bioCtrl.dispose(); _universityCtrl.dispose(); _departmentCtrl.dispose();
    _locationCtrl.dispose(); _skillsCtrl.dispose(); _languagesCtrl.dispose();
    _githubCtrl.dispose(); _linkedinCtrl.dispose(); _portfolioCtrl.dispose();
    _currentPassCtrl.dispose(); _newPassCtrl.dispose();
    super.dispose();
  }

  void _populate(Map<String, dynamic> data) {
    if (_initialized) return;
    _initialized = true;
    final p = (data['profile'] as Map?) ?? {};
    _firstNameCtrl.text  = p['first_name']  ?? '';
    _lastNameCtrl.text   = p['last_name']   ?? '';
    _taglineCtrl.text    = p['tagline']     ?? '';
    _bioCtrl.text        = p['bio']         ?? '';
    _universityCtrl.text = p['university']  ?? '';
    _departmentCtrl.text = p['department']  ?? '';
    _locationCtrl.text   = p['location']    ?? '';
    _githubCtrl.text     = p['github_url']  ?? '';
    _linkedinCtrl.text   = p['linkedin_url']  ?? '';
    _portfolioCtrl.text  = p['portfolio_url'] ?? '';
    _skillsCtrl.text     = (p['skills']    as List? ?? []).join(', ');
    _languagesCtrl.text  = (p['languages'] as List? ?? []).join(', ');
    _avatarUrl = p['avatar_url'] as String?;
  }

  Future<void> _pickAvatar() async {
    final picker = ImagePicker();
    final picked = await picker.pickImage(source: ImageSource.gallery, imageQuality: 85);
    if (picked == null) return;

    setState(() => _uploadingAvatar = true);
    try {
      final dio = ref.read(dioProvider);
      final form = FormData.fromMap({
        'file': await MultipartFile.fromFile(picked.path, filename: 'avatar.jpg'),
      });
      final res = await dio.post('/api/v1/upload/avatar', data: form);
      setState(() => _avatarUrl = res.data['url'] as String?);
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Avatar updated!')));
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Upload failed: $e')));
    } finally {
      setState(() => _uploadingAvatar = false);
    }
  }

  Future<void> _save() async {
    setState(() => _saving = true);
    try {
      await ref.read(dioProvider).patch('/api/v1/users/profile', data: {
        'first_name':    _firstNameCtrl.text.trim(),
        'last_name':     _lastNameCtrl.text.trim(),
        'tagline':       _taglineCtrl.text.trim(),
        'bio':           _bioCtrl.text.trim(),
        'university':    _universityCtrl.text.trim(),
        'department':    _departmentCtrl.text.trim(),
        'location':      _locationCtrl.text.trim(),
        'github_url':    _githubCtrl.text.trim(),
        'linkedin_url':  _linkedinCtrl.text.trim(),
        'portfolio_url': _portfolioCtrl.text.trim(),
        'avatar_url':    _avatarUrl ?? '',
        'skills':    _skillsCtrl.text.split(',').map((s) => s.trim()).where((s) => s.isNotEmpty).toList(),
        'languages': _languagesCtrl.text.split(',').map((s) => s.trim()).where((s) => s.isNotEmpty).toList(),
      });
      ref.invalidate(_meProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Changes saved!')));
      }
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
    } finally {
      setState(() => _saving = false);
    }
  }

  Future<void> _changePassword() async {
    if (_currentPassCtrl.text.isEmpty || _newPassCtrl.text.isEmpty) return;
    setState(() => _changingPass = true);
    try {
      await ref.read(dioProvider).post('/api/v1/auth/change-password', data: {
        'current_password': _currentPassCtrl.text,
        'new_password':     _newPassCtrl.text,
      });
      _currentPassCtrl.clear();
      _newPassCtrl.clear();
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Password updated!')));
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
    } finally {
      setState(() => _changingPass = false);
    }
  }

  Future<void> _logout() async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Sign out?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Cancel')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Sign Out')),
        ],
      ),
    );
    if (ok != true) return;
    await _storage.deleteAll();
    if (mounted) context.go('/login');
  }

  @override
  Widget build(BuildContext context) {
    final me     = ref.watch(_meProvider);
    final scheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text('Settings')),
      body: me.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
        data: (data) {
          _populate(data);
          final displayAvatar = _avatarUrl?.isNotEmpty == true
              ? _avatarUrl!
              : 'https://ui-avatars.com/api/?name=${data['username']}&size=80&background=2563eb&color=fff';

          return ListView(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 40),
            children: [
              // Avatar section
              _Section(title: 'Profile Photo', children: [
                Center(child: Stack(clipBehavior: Clip.none, children: [
                  CircleAvatar(
                    radius: 48,
                    backgroundImage: NetworkImage(displayAvatar),
                  ),
                  if (_uploadingAvatar)
                    const Positioned.fill(child: CircleAvatar(
                      radius: 48,
                      backgroundColor: Colors.black45,
                      child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                    )),
                  Positioned(
                    right: 0, bottom: 0,
                    child: GestureDetector(
                      onTap: _pickAvatar,
                      child: Container(
                        width: 32, height: 32,
                        decoration: BoxDecoration(color: scheme.primary, shape: BoxShape.circle),
                        child: const Icon(Icons.camera_alt, color: Colors.white, size: 16),
                      ),
                    ),
                  ),
                ])),
              ]),

              const SizedBox(height: 12),

              // Profile info
              _Section(title: 'Profile Information', children: [
                _row([_field('First Name', _firstNameCtrl), _field('Last Name', _lastNameCtrl)]),
                _field('Tagline', _taglineCtrl, hint: 'Full-stack dev & Math tutor'),
                _field('Bio', _bioCtrl, hint: 'Tell buyers about yourself…', maxLines: 3),
                _row([_field('University', _universityCtrl), _field('Department', _departmentCtrl)]),
                _field('City / Location', _locationCtrl, hint: 'Almaty, Kazakhstan'),
                _field('Skills', _skillsCtrl, hint: 'React, Python, Figma…'),
                _field('Languages', _languagesCtrl, hint: 'English, Russian, Kazakh…'),
              ]),

              const SizedBox(height: 12),

              // Links
              _Section(title: 'Links', children: [
                _field('GitHub', _githubCtrl, hint: 'https://github.com/you'),
                _field('LinkedIn', _linkedinCtrl, hint: 'https://linkedin.com/in/you'),
                _field('Portfolio', _portfolioCtrl, hint: 'https://yoursite.com'),
              ]),

              const SizedBox(height: 12),

              // Save button
              FilledButton(
                onPressed: _saving ? null : _save,
                style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(50)),
                child: _saving ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white)) : const Text('Save Changes', style: TextStyle(fontSize: 16)),
              ),

              const SizedBox(height: 20),

              // Password
              _Section(title: 'Change Password', children: [
                _field('Current Password', _currentPassCtrl, obscure: true),
                _field('New Password', _newPassCtrl, obscure: true, hint: 'Min. 8 characters'),
                Align(
                  alignment: Alignment.centerLeft,
                  child: OutlinedButton(
                    onPressed: _changingPass ? null : _changePassword,
                    child: _changingPass ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2)) : const Text('Update Password'),
                  ),
                ),
              ]),

              const SizedBox(height: 12),

              // Danger zone
              Container(
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: scheme.error.withValues(alpha: 0.4)),
                ),
                child: ListTile(
                  leading: Icon(Icons.logout_rounded, color: scheme.error),
                  title: Text('Sign Out', style: TextStyle(color: scheme.error, fontWeight: FontWeight.w600)),
                  onTap: _logout,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _row(List<Widget> children) => Row(
    children: children.map((c) => Expanded(child: c)).toList(),
  );

  Widget _field(String label, TextEditingController ctrl, {String? hint, bool obscure = false, int maxLines = 1}) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text(label, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600)),
      const SizedBox(height: 4),
      TextField(
        controller: ctrl,
        obscureText: obscure,
        maxLines: maxLines,
        decoration: InputDecoration(
          hintText: hint,
          isDense: true,
          border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
          contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        ),
      ),
    ]),
  );
}

class _Section extends StatelessWidget {
  final String title;
  final List<Widget> children;
  const _Section({required this.title, required this.children});

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      margin: const EdgeInsets.only(bottom: 4),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: scheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: scheme.outlineVariant.withValues(alpha: 0.5)),
      ),
      child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text(title, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15)),
        const SizedBox(height: 14),
        ...children,
      ]),
    );
  }
}
