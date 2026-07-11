import 'package:dio/dio.dart';

import '../core/config.dart';
import 'models.dart';

class RandomMatchApi {
  RandomMatchApi()
      : _dio = Dio(BaseOptions(
          baseUrl: apiBase,
          connectTimeout: const Duration(seconds: 10),
          receiveTimeout: const Duration(seconds: 20),
        ));

  final Dio _dio;

  String? token;

  Options get _authOptions => Options(headers: {
        if (token != null && token!.isNotEmpty)
          'Authorization': 'Bearer $token',
      });

  Future<bool> verifySession() async {
    if (token == null || token!.isEmpty) return false;
    final res =
        await _dio.get<dynamic>('/api/v1/auth/session', options: _authOptions);
    return res.statusCode == 200;
  }

  Future<AuthResponse> anonymousAuth() async {
    final res = await _dio.post<Map<String, dynamic>>('/api/v1/auth/anonymous');
    final auth = AuthResponse.fromJson(res.data ?? {});
    token = auth.token;
    return auth;
  }

  Future<UserProfile> fetchProfile() async {
    final res = await _dio.get<Map<String, dynamic>>('/api/v1/me',
        options: _authOptions);
    return UserProfile.fromJson(
        res.data?['user'] as Map<String, dynamic>? ?? {});
  }

  Future<UserProfile> updateProfile({
    required String displayName,
    required String bio,
    required List<String> interests,
    required bool ageConfirmed,
    required String region,
    required String gender,
    required String language,
  }) async {
    final res = await _dio.put<Map<String, dynamic>>(
      '/api/v1/me',
      data: {
        'displayName': displayName,
        'bio': bio,
        'interests': interests,
        'ageConfirmed': ageConfirmed,
        'region': region,
        'gender': gender,
        'language': language,
      },
      options: _authOptions,
    );
    return UserProfile.fromJson(
        res.data?['user'] as Map<String, dynamic>? ?? {});
  }

  Future<RuntimeStats> fetchStats() async {
    final res = await _dio.get<Map<String, dynamic>>('/api/v1/stats');
    return RuntimeStats.fromJson(res.data ?? {});
  }

  Future<MatchResponse> joinMatch({
    required String mode,
    required String region,
    required String gender,
    required String language,
    required List<String> interests,
  }) async {
    final res = await _dio.post<Map<String, dynamic>>(
      '/api/v1/match/join',
      data: {
        'mode': mode,
        'region': region,
        'gender': gender,
        'language': language,
        'interests': interests,
      },
      options: _authOptions,
    );
    return MatchResponse.fromJson(res.data ?? {});
  }

  Future<List<UserProfile>> fetchDiscoverProfiles({
    required String region,
    required String gender,
  }) async {
    final res = await _dio.get<Map<String, dynamic>>(
      '/api/v1/discover/profiles',
      queryParameters: {'region': region, 'gender': gender},
      options: _authOptions,
    );
    final users = res.data?['users'] as List? ?? const [];
    return users
        .whereType<Map<String, dynamic>>()
        .map(UserProfile.fromJson)
        .toList();
  }

  Future<void> leaveMatch() async {
    await _dio.post<dynamic>('/api/v1/match/leave', options: _authOptions);
  }

  Future<CommerceStatus> fetchCommerceStatus() async {
    final res = await _dio.get<Map<String, dynamic>>(
      '/api/v1/commerce/status',
      options: _authOptions,
    );
    return CommerceStatus.fromJson(res.data ?? {});
  }

  Future<PaymentOrder> createPaymentOrder() async {
    final res = await _dio.post<Map<String, dynamic>>(
      '/api/v1/commerce/orders',
      data: {'plan': 'premium_monthly'},
      options: _authOptions,
    );
    return PaymentOrder.fromJson(
        res.data?['order'] as Map<String, dynamic>? ?? {});
  }

  Future<void> confirmPaymentOrder(String orderId) async {
    await _dio.post<dynamic>(
      '/api/v1/commerce/orders/$orderId/confirm',
      options: _authOptions,
    );
  }

  Future<void> reportUser(String userId) async {
    await _dio.post<dynamic>(
      '/api/v1/users/$userId/report',
      data: {'reason': 'user reported during match'},
      options: _authOptions,
    );
  }

  Future<void> followUser(String userId) async {
    await _dio.post<dynamic>('/api/v1/users/$userId/follow',
        options: _authOptions);
  }

  Future<List<UserProfile>> fetchFollowedUsers() async {
    final res = await _dio.get<Map<String, dynamic>>(
      '/api/v1/users/follows',
      options: _authOptions,
    );
    final users = res.data?['users'] as List? ?? const [];
    return users
        .whereType<Map<String, dynamic>>()
        .map(UserProfile.fromJson)
        .toList();
  }

  Future<void> unfollowUser(String userId) async {
    await _dio.delete<dynamic>('/api/v1/users/$userId/follow',
        options: _authOptions);
  }

  Future<void> sendDirectMessage({
    required String userId,
    required String text,
  }) async {
    await _dio.post<dynamic>(
      '/api/v1/users/$userId/messages',
      data: {'text': text},
      options: _authOptions,
    );
  }

  Future<void> blockUser(String userId) async {
    await _dio.post<dynamic>('/api/v1/users/$userId/block',
        options: _authOptions);
  }

  Future<List<BlockedUser>> fetchBlockedUsers() async {
    final res = await _dio.get<Map<String, dynamic>>(
      '/api/v1/users/blocks',
      options: _authOptions,
    );
    final users = res.data?['users'] as List? ?? const [];
    return users
        .whereType<Map<String, dynamic>>()
        .map(BlockedUser.fromJson)
        .toList();
  }

  Future<void> unblockUser(String userId) async {
    await _dio.delete<dynamic>('/api/v1/users/$userId/block',
        options: _authOptions);
  }

  Future<void> savePushDeviceToken({
    required String token,
    required String platform,
  }) async {
    await _dio.post<dynamic>(
      '/api/v1/push/device-token',
      data: {'token': token, 'platform': platform},
      options: _authOptions,
    );
  }

  Uri wsUri() {
    final uri = Uri.parse(apiBase);
    return uri.replace(
      scheme: uri.scheme == 'https' ? 'wss' : 'ws',
      path: '/api/v1/ws',
      queryParameters: {'token': token ?? ''},
    );
  }

  String userMessage(Object error) {
    if (error is DioException) {
      if (error.response?.statusCode == 404) {
        return '后端尚未支持此功能，请部署最新后端后再试';
      }
      if (error.response?.statusCode == 402) {
        final data = error.response?.data;
        final remaining = data is Map ? data['dailyRemaining'] ?? 0 : 0;
        return '今日免费匹配次数已用完，剩余 $remaining 次。开通会员可无限匹配并优先排队';
      }
      if (error.response?.statusCode == 409) return '已跳过拉黑用户，请再点一次随机匹配';
      if (error.type == DioExceptionType.connectionError) return '无法连接后端服务';
    }
    return error.toString().replaceFirst('Exception: ', '');
  }
}
