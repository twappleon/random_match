enum MatchStatus { idle, waiting, matched }

enum AppPage { video, discover, profile, membership }

enum MatchModePreference { video, voice }

enum GenderPreference { everyone, female, male }

enum TranslationLanguage { off, zh, en, ja, ko, es }

enum ChatSender { self, peer }

class UserProfile {
  const UserProfile({
    required this.id,
    required this.displayName,
    required this.avatarUrl,
    required this.ageConfirmed,
    this.bio = '',
    this.interests = const [],
    this.region = 'global',
    this.gender = '',
    this.language = 'zh',
    this.trustBadge = false,
    this.membershipPlan = '',
    this.membershipExpiresAt,
    this.isMember = false,
  });

  final String id;
  final String displayName;
  final String avatarUrl;
  final bool ageConfirmed;
  final String bio;
  final List<String> interests;
  final String region;
  final String gender;
  final String language;
  final bool trustBadge;
  final String membershipPlan;
  final DateTime? membershipExpiresAt;
  final bool isMember;

  factory UserProfile.fromJson(Map<String, dynamic> json) {
    return UserProfile(
      id: json['id'] as String? ?? json['_id'] as String? ?? '',
      displayName: json['displayName'] as String? ?? '星球旅人',
      avatarUrl: json['avatarUrl'] as String? ?? '',
      ageConfirmed: json['ageConfirmed'] == true,
      bio: json['bio'] as String? ?? '',
      interests: (json['interests'] as List? ?? const [])
          .map((item) => item.toString())
          .where((item) => item.trim().isNotEmpty)
          .toList(),
      region: json['region'] as String? ?? 'global',
      gender: json['gender'] as String? ?? '',
      language: json['language'] as String? ?? 'zh',
      trustBadge: json['trustBadge'] == true,
      membershipPlan: json['membershipPlan'] as String? ?? '',
      membershipExpiresAt: DateTime.tryParse(
        json['membershipExpiresAt'] as String? ?? '',
      ),
      isMember: json['isMember'] == true,
    );
  }
}

class AuthResponse {
  const AuthResponse({required this.token, required this.user});

  final String token;
  final UserProfile user;

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      token: json['token'] as String? ?? '',
      user: UserProfile.fromJson(json['user'] as Map<String, dynamic>? ?? {}),
    );
  }
}

class MatchResponse {
  const MatchResponse({
    required this.status,
    this.roomId,
    this.peerId,
    this.peerProfile,
    this.initiator = false,
  });

  final MatchStatus status;
  final String? roomId;
  final String? peerId;
  final UserProfile? peerProfile;
  final bool initiator;

  factory MatchResponse.fromJson(Map<String, dynamic> json) {
    return MatchResponse(
      status: json['status'] == 'matched'
          ? MatchStatus.matched
          : MatchStatus.waiting,
      roomId: json['roomId'] as String?,
      peerId: json['peerId'] as String?,
      peerProfile: json['peerProfile'] is Map<String, dynamic>
          ? UserProfile.fromJson(json['peerProfile'] as Map<String, dynamic>)
          : null,
      initiator: json['initiator'] == true,
    );
  }
}

class RuntimeStats {
  const RuntimeStats({
    required this.online,
    required this.waiting,
    required this.chatting,
  });

  final int online;
  final int waiting;
  final int chatting;

  factory RuntimeStats.fromJson(Map<String, dynamic> json) {
    return RuntimeStats(
      online: json['online'] as int? ?? 0,
      waiting: json['waiting'] as int? ?? 0,
      chatting: json['chatting'] as int? ?? 0,
    );
  }
}

class CommerceStatus {
  const CommerceStatus({
    required this.isMember,
    required this.dailyLimit,
    required this.dailyUsed,
    required this.dailyRemaining,
    required this.priorityQueue,
    required this.gemsBalance,
    this.membershipPlan = '',
    this.membershipExpiresAt,
  });

  final bool isMember;
  final int dailyLimit;
  final int dailyUsed;
  final int dailyRemaining;
  final bool priorityQueue;
  final int gemsBalance;
  final String membershipPlan;
  final DateTime? membershipExpiresAt;

  factory CommerceStatus.fromJson(Map<String, dynamic> json) {
    return CommerceStatus(
      isMember: json['isMember'] == true,
      dailyLimit: json['dailyLimit'] as int? ?? 0,
      dailyUsed: json['dailyUsed'] as int? ?? 0,
      dailyRemaining: json['dailyRemaining'] as int? ?? 0,
      priorityQueue: json['priorityQueue'] == true,
      gemsBalance: json['gemsBalance'] as int? ?? 0,
      membershipPlan: json['membershipPlan'] as String? ?? '',
      membershipExpiresAt: DateTime.tryParse(
        json['membershipExpiresAt'] as String? ?? '',
      ),
    );
  }
}

class PaymentOrder {
  const PaymentOrder({required this.id});

  final String id;

  factory PaymentOrder.fromJson(Map<String, dynamic> json) {
    return PaymentOrder(id: json['id'] as String? ?? '');
  }
}

class BlockedUser {
  const BlockedUser({required this.user, required this.createdAt});

  final UserProfile user;
  final DateTime createdAt;

  factory BlockedUser.fromJson(Map<String, dynamic> json) {
    return BlockedUser(
      user: UserProfile.fromJson(json['user'] as Map<String, dynamic>? ?? {}),
      createdAt: DateTime.tryParse(json['createdAt'] as String? ?? '') ??
          DateTime.now(),
    );
  }
}

class ChatMessage {
  const ChatMessage({
    required this.id,
    required this.sender,
    required this.text,
    required this.createdAt,
  });

  final String id;
  final ChatSender sender;
  final String text;
  final DateTime createdAt;
}
