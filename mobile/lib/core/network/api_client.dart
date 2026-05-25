import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

// Override at compile time:
//   --dart-define=API_URL=https://your-server
//   --dart-define=WS_URL=wss://your-server
const _customUrl = String.fromEnvironment('API_URL', defaultValue: '');
const _customWs  = String.fromEnvironment('WS_URL',  defaultValue: '');
const _productionApiUrl = 'https://www.404tears.kz';
const _productionWsUrl = 'wss://www.404tears.kz';

/// Platform-aware API base URL.
/// Release builds use the production domain automatically.
/// Override at build time: --dart-define=API_URL=https://your-server.com
String get apiBaseUrl {
  if (_customUrl.isNotEmpty) return _customUrl;
  if (kReleaseMode) return _productionApiUrl;
  if (kIsWeb) return 'http://localhost:8080';
  if (defaultTargetPlatform == TargetPlatform.android) return 'http://10.0.2.2:8080';
  return 'http://localhost:8080';
}

/// Platform-aware WebSocket base URL.
String get wsBaseUrl {
  if (_customWs.isNotEmpty) return _customWs;
  if (kReleaseMode) return _productionWsUrl;
  if (kIsWeb) return 'ws://localhost:8080';
  if (defaultTargetPlatform == TargetPlatform.android) return 'ws://10.0.2.2:8080';
  return 'ws://localhost:8080';
}

String get _baseUrl => apiBaseUrl;

const _storage = FlutterSecureStorage(
  aOptions: AndroidOptions(encryptedSharedPreferences: true),
);

Future<String?> readSecureValue(String key) async {
  try {
    return await _storage.read(key: key);
  } catch (_) {
    try {
      await _storage.deleteAll();
    } catch (_) {}
    return null;
  }
}

final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(BaseOptions(
    baseUrl: _baseUrl,
    connectTimeout: const Duration(seconds: 15),
    receiveTimeout: const Duration(seconds: 30),
    headers: {'Content-Type': 'application/json'},
  ));

  dio.interceptors.add(_AuthInterceptor(dio));
  if (kDebugMode) {
    dio.interceptors.add(LogInterceptor(
      requestBody: false,
      responseBody: false,
      logPrint: (o) => debugPrint('[API] $o'),
    ));
  }

  return dio;
});

class _AuthInterceptor extends Interceptor {
  final Dio _dio;
  _AuthInterceptor(this._dio);

  @override
  Future<void> onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    final token = await readSecureValue('access_token');
    if (token != null) options.headers['Authorization'] = 'Bearer $token';
    handler.next(options);
  }

  @override
  Future<void> onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode != 401) return handler.next(err);

    try {
      final refresh = await readSecureValue('refresh_token');
      if (refresh == null) return handler.next(err);

      final res = await Dio(BaseOptions(baseUrl: _baseUrl)).post(
        '/api/v1/auth/refresh',
        data: {'refresh_token': refresh},
      );
      final newAccess  = res.data['access_token'] as String;
      final newRefresh = res.data['refresh_token'] as String;
      await _storage.write(key: 'access_token',  value: newAccess);
      await _storage.write(key: 'refresh_token', value: newRefresh);

      err.requestOptions.headers['Authorization'] = 'Bearer $newAccess';
      final retried = await _dio.fetch(err.requestOptions);
      handler.resolve(retried);
    } catch (_) {
      await _storage.deleteAll();
      handler.next(err);
    }
  }
}

class ApiException implements Exception {
  final String message;
  final int? statusCode;
  ApiException(this.message, [this.statusCode]);

  @override
  String toString() => message;
}
