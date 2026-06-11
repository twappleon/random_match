import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:get/get.dart' hide navigator;
import 'package:web_socket_channel/web_socket_channel.dart';

import '../data/models.dart';
import '../data/random_match_api.dart';

class MatchController extends GetxController {
  MatchController({RandomMatchApi? api}) : api = api ?? RandomMatchApi();

  final RandomMatchApi api;

  final localRenderer = RTCVideoRenderer();
  final remoteRenderer = RTCVideoRenderer();
  final chatInput = TextEditingController();
  final chatScroll = ScrollController();
  final displayNameInput = TextEditingController(text: '星球旅人');
  final bioInput = TextEditingController();
  final interestsInput = TextEditingController(text: '聊天, 电影, 音乐');

  final page = AppPage.video.obs;
  final status = MatchStatus.idle.obs;
  final loading = false.obs;
  final leaving = false.obs;
  final savingProfile = false.obs;
  final paymentLoading = false.obs;
  final safetyLoading = false.obs;
  final blockedLoading = false.obs;
  final chatOpen = false.obs;
  final peerCardHidden = false.obs;
  final ageConfirmed = false.obs;

  final stats = const RuntimeStats(online: 0, waiting: 0, chatting: 0).obs;
  final profile = Rxn<UserProfile>();
  final peerProfile = Rxn<UserProfile>();
  final commerceStatus = Rxn<CommerceStatus>();
  final blockedUsers = <BlockedUser>[].obs;
  final messages = <ChatMessage>[].obs;

  String? activePeerId;
  String? activeRoomId;
  String? reportedPeerId;
  String errorText = '';

  WebSocketChannel? _socket;
  RTCPeerConnection? _peer;
  MediaStream? _localStream;
  Timer? _statsTimer;
  StreamSubscription<String>? _pushTokenSubscription;
  bool _remoteDescriptionReady = false;
  bool _pushInitialized = false;
  final List<RTCIceCandidate> _pendingCandidates = [];

  @override
  void onInit() {
    super.onInit();
    unawaited(_init());
  }

  Future<void> _init() async {
    await localRenderer.initialize();
    await remoteRenderer.initialize();
    _statsTimer =
        Timer.periodic(const Duration(seconds: 5), (_) => refreshStats());
    await refreshStats();
    await loadProfile();
    unawaited(setupPushNotifications());
  }

  @override
  void onClose() {
    _statsTimer?.cancel();
    _pushTokenSubscription?.cancel();
    _socket?.sink.close();
    _peer?.close();
    _localStream?.dispose();
    chatInput.dispose();
    chatScroll.dispose();
    displayNameInput.dispose();
    bioInput.dispose();
    interestsInput.dispose();
    unawaited(localRenderer.dispose());
    unawaited(remoteRenderer.dispose());
    super.onClose();
  }

  Future<void> refreshStats() async {
    try {
      stats.value = await api.fetchStats();
    } catch (_) {
      // Keep the last visible values.
    }
  }

  Future<void> ensureAuth() async {
    if (await api.verifySession()) {
      unawaited(setupPushNotifications());
      return;
    }
    final auth = await api.anonymousAuth();
    setProfile(auth.user);
    unawaited(setupPushNotifications());
  }

  Future<void> setupPushNotifications() async {
    if (_pushInitialized || api.token == null || api.token!.isEmpty) return;
    _pushInitialized = true;
    try {
      final messaging = FirebaseMessaging.instance;
      if (Platform.isIOS || Platform.isMacOS) {
        await messaging.requestPermission(
          alert: true,
          badge: true,
          sound: true,
        );
        await messaging.setForegroundNotificationPresentationOptions(
          alert: true,
          badge: true,
          sound: true,
        );
      } else if (Platform.isAndroid) {
        await messaging.requestPermission();
      }

      final token = await messaging.getToken();
      if (token != null && token.isNotEmpty) {
        await api.savePushDeviceToken(
          token: token,
          platform: Platform.isIOS ? 'ios' : 'android',
        );
      }
      _pushTokenSubscription ??=
          messaging.onTokenRefresh.listen((nextToken) async {
        try {
          if (nextToken.isNotEmpty) {
            await api.savePushDeviceToken(
              token: nextToken,
              platform: Platform.isIOS ? 'ios' : 'android',
            );
          }
        } catch (_) {
          // Token registration is retried when the app starts again.
        }
      });
    } catch (_) {
      _pushInitialized = false;
      // Firebase native config files are supplied per deployment.
    }
  }

  Future<void> loadProfile() async {
    try {
      await ensureAuth();
      setProfile(await api.fetchProfile());
      await loadCommerceStatus();
      await loadBlockedUsers();
    } catch (_) {
      // Profile is refreshed again before matching.
    }
  }

  void setProfile(UserProfile next) {
    profile.value = next;
    displayNameInput.text = next.displayName;
    bioInput.text = next.bio;
    ageConfirmed.value = next.ageConfirmed;
    interestsInput.text =
        (next.interests.isEmpty ? ['聊天', '电影', '音乐'] : next.interests)
            .join(', ');
  }

  List<String> parsedInterests() {
    final seen = <String>{};
    return interestsInput.text
        .split(RegExp('[,，]'))
        .map((item) => item.trim())
        .where((item) => item.isNotEmpty && seen.add(item))
        .take(6)
        .toList();
  }

  Future<void> saveProfile() async {
    savingProfile.value = true;
    try {
      await ensureAuth();
      final next = await api.updateProfile(
        displayName: displayNameInput.text.trim().isEmpty
            ? '星球旅人'
            : displayNameInput.text.trim(),
        bio: bioInput.text.trim(),
        interests: parsedInterests(),
        ageConfirmed: ageConfirmed.value,
      );
      setProfile(next);
      Get.snackbar('资料已保存', '新的匿名身份已更新', snackPosition: SnackPosition.BOTTOM);
    } catch (error) {
      showError(error);
    } finally {
      savingProfile.value = false;
    }
  }

  Future<void> startMatch() async {
    if (loading.value) return;
    loading.value = true;
    try {
      page.value = AppPage.video;
      resetCall();
      if (!ageConfirmed.value) throw Exception('请先确认已满 18 岁并保存资料');
      await saveProfile();
      await ensureAuth();
      await openMedia();
      openSocket();
      final result = await api.joinMatch();
      status.value = result.status;
      if (result.status == MatchStatus.matched) {
        activeRoomId = result.roomId;
        activePeerId = result.peerId;
        peerProfile.value = result.peerProfile;
        peerCardHidden.value = false;
        if (result.initiator && result.peerId != null) {
          await createPeer(result.peerId!);
        }
      }
      await loadCommerceStatus();
      await refreshStats();
    } catch (error) {
      showError(error);
      await loadCommerceStatus();
    } finally {
      loading.value = false;
    }
  }

  Future<void> leaveCall() async {
    if (leaving.value || status.value == MatchStatus.idle) return;
    leaving.value = true;
    try {
      await api.leaveMatch();
    } catch (error) {
      showError(error);
    } finally {
      closeSocket();
      stopMedia();
      resetCall();
      leaving.value = false;
      await refreshStats();
    }
  }

  Future<void> openMedia() async {
    stopMedia();
    _localStream = await navigator.mediaDevices.getUserMedia({
      'audio': {
        'echoCancellation': true,
        'noiseSuppression': true,
        'autoGainControl': true,
      },
      'video': {
        'width': {'ideal': 480, 'max': 640},
        'height': {'ideal': 640, 'max': 720},
        'frameRate': {'ideal': 15, 'max': 20},
        'facingMode': {'ideal': 'user'},
      },
    });
    localRenderer.srcObject = _localStream;
  }

  void stopMedia() {
    _localStream?.dispose();
    _localStream = null;
    localRenderer.srcObject = null;
  }

  void openSocket() {
    if (_socket != null) return;
    _socket = WebSocketChannel.connect(api.wsUri());
    _socket!.stream.listen(
      (event) => unawaited(
          handleSignal(jsonDecode(event as String) as Map<String, dynamic>)),
      onDone: () {
        _socket = null;
        if (status.value == MatchStatus.matched) resetCall('信令连接已断开，请重新匹配');
      },
      onError: (_) => showError(Exception('信令连接失败')),
    );
  }

  void closeSocket() {
    _socket?.sink.close();
    _socket = null;
  }

  Future<void> handleSignal(Map<String, dynamic> msg) async {
    switch (msg['type']) {
      case 'matched':
        status.value = MatchStatus.matched;
        activeRoomId = msg['roomId'] as String?;
        activePeerId = msg['peerId'] as String?;
        peerProfile.value = msg['peerProfile'] is Map<String, dynamic>
            ? UserProfile.fromJson(msg['peerProfile'] as Map<String, dynamic>)
            : null;
        peerCardHidden.value = false;
        if (msg['initiator'] == true && activePeerId != null) {
          await createPeer(activePeerId!);
        }
      case 'offer':
        await acceptOffer(msg['peerId'] as String, msg['data']);
      case 'answer':
        await _peer?.setRemoteDescription(descriptionFrom(msg['data']));
        _remoteDescriptionReady = true;
        await flushCandidates();
      case 'candidate':
        await addRemoteCandidate(candidateFrom(msg['data']));
      case 'chat-message':
        receiveChatMessage(msg['data']);
      case 'peer-left':
        resetCall('对方已离开，请重新匹配');
    }
  }

  Future<void> createPeer(String peerId) async {
    await teardownPeer(clearSession: false);
    activePeerId = peerId;
    final pc = await buildPeer(peerId);
    _peer = pc;
    await addLocalTracks(pc);
    final offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    send({'type': 'offer', 'peerId': peerId, 'data': offer.toMap()});
  }

  Future<void> acceptOffer(String peerId, dynamic offer) async {
    await teardownPeer(clearSession: false);
    activePeerId = peerId;
    final pc = await buildPeer(peerId);
    _peer = pc;
    await addLocalTracks(pc);
    await pc.setRemoteDescription(descriptionFrom(offer));
    _remoteDescriptionReady = true;
    await flushCandidates();
    final answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);
    send({'type': 'answer', 'peerId': peerId, 'data': answer.toMap()});
  }

  Future<RTCPeerConnection> buildPeer(String peerId) async {
    final pc = await createPeerConnection({
      'iceServers': [
        {'urls': 'stun:stun.l.google.com:19302'},
      ],
    });
    pc.onIceCandidate = (candidate) {
      send({'type': 'candidate', 'peerId': peerId, 'data': candidate.toMap()});
    };
    pc.onTrack = (event) {
      if (event.streams.isNotEmpty) {
        remoteRenderer.srcObject = event.streams.first;
      }
    };
    return pc;
  }

  Future<void> addLocalTracks(RTCPeerConnection pc) async {
    final stream = _localStream;
    if (stream == null) return;
    for (final track in stream.getTracks()) {
      await pc.addTrack(track, stream);
    }
  }

  RTCSessionDescription descriptionFrom(dynamic data) {
    final map = Map<String, dynamic>.from(data as Map);
    return RTCSessionDescription(map['sdp'] as String?, map['type'] as String?);
  }

  RTCIceCandidate candidateFrom(dynamic data) {
    final map = Map<String, dynamic>.from(data as Map);
    return RTCIceCandidate(
      map['candidate'] as String?,
      map['sdpMid'] as String?,
      map['sdpMLineIndex'] as int?,
    );
  }

  Future<void> addRemoteCandidate(RTCIceCandidate candidate) async {
    if (!_remoteDescriptionReady) {
      _pendingCandidates.add(candidate);
      return;
    }
    await _peer?.addCandidate(candidate);
  }

  Future<void> flushCandidates() async {
    if (!_remoteDescriptionReady) return;
    for (final candidate in _pendingCandidates) {
      await _peer?.addCandidate(candidate);
    }
    _pendingCandidates.clear();
  }

  void send(Map<String, dynamic> message) {
    _socket?.sink.add(jsonEncode(message));
  }

  Future<void> teardownPeer({bool clearSession = true}) async {
    await _peer?.close();
    _peer = null;
    activePeerId = null;
    _remoteDescriptionReady = false;
    _pendingCandidates.clear();
    remoteRenderer.srcObject = null;
    if (clearSession) {
      activeRoomId = null;
      peerProfile.value = null;
      peerCardHidden.value = false;
      chatOpen.value = false;
      messages.clear();
      chatInput.clear();
    }
  }

  void resetCall([String message = '']) {
    unawaited(teardownPeer());
    status.value = MatchStatus.idle;
    if (message.isNotEmpty) {
      Get.snackbar('通话已结束', message, snackPosition: SnackPosition.BOTTOM);
    }
  }

  void toggleChat() {
    chatOpen.toggle();
    if (chatOpen.value) scrollChatToBottom();
  }

  void sendChatMessage() {
    final peerId = activePeerId;
    final text = chatInput.text.trim();
    if (status.value != MatchStatus.matched || peerId == null || text.isEmpty) {
      return;
    }
    final message = ChatMessage(
      id: DateTime.now().microsecondsSinceEpoch.toString(),
      sender: ChatSender.self,
      text: truncateText(text, 500),
      createdAt: DateTime.now(),
    );
    messages.add(message);
    chatInput.clear();
    scrollChatToBottom();
    send({
      'type': 'chat-message',
      'peerId': peerId,
      if (activeRoomId != null) 'roomId': activeRoomId,
      'data': {
        'id': message.id,
        'text': message.text,
        'createdAt': message.createdAt.toIso8601String(),
      },
    });
  }

  void receiveChatMessage(dynamic data) {
    if (data is! Map) return;
    final rawText = data['text'];
    if (rawText is! String) return;
    final text = rawText.trim();
    if (text.isEmpty) return;
    messages.add(ChatMessage(
      id: data['id'] as String? ??
          DateTime.now().microsecondsSinceEpoch.toString(),
      sender: ChatSender.peer,
      text: truncateText(text, 500),
      createdAt: DateTime.tryParse(data['createdAt'] as String? ?? '') ??
          DateTime.now(),
    ));
    if (!chatOpen.value) {
      Get.snackbar(
        '新文字讯息',
        toastPreview(text),
        snackPosition: SnackPosition.BOTTOM,
        mainButton: TextButton(
          onPressed: () {
            chatOpen.value = true;
            Get.closeCurrentSnackbar();
          },
          child: const Text('查看'),
        ),
      );
    }
    scrollChatToBottom();
  }

  void scrollChatToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!chatScroll.hasClients) return;
      chatScroll.animateTo(
        chatScroll.position.maxScrollExtent,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
      );
    });
  }

  String truncateText(String value, int maxLength) {
    final chars = value.characters;
    return chars.length > maxLength ? chars.take(maxLength).toString() : value;
  }

  String toastPreview(String value) {
    final truncated = truncateText(value, 42);
    return truncated == value ? value : '$truncated...';
  }

  Future<void> loadCommerceStatus() async {
    try {
      await ensureAuth();
      commerceStatus.value = await api.fetchCommerceStatus();
    } catch (error) {
      showError(error);
    }
  }

  Future<void> buyMembership() async {
    if (paymentLoading.value || commerceStatus.value?.isMember == true) return;
    paymentLoading.value = true;
    try {
      await ensureAuth();
      final order = await api.createPaymentOrder();
      await api.confirmPaymentOrder(order.id);
      await loadCommerceStatus();
      Get.snackbar('会员已开通', '可无限匹配并优先排队', snackPosition: SnackPosition.BOTTOM);
    } catch (error) {
      showError(error);
    } finally {
      paymentLoading.value = false;
    }
  }

  Future<void> reportPeer() async {
    final peerId = activePeerId;
    if (peerId == null || safetyLoading.value) return;
    safetyLoading.value = true;
    try {
      await api.reportUser(peerId);
      reportedPeerId = peerId;
      Get.snackbar('已收到举报', '我们会保留记录供后续处理',
          snackPosition: SnackPosition.BOTTOM);
    } catch (error) {
      showError(error);
    } finally {
      safetyLoading.value = false;
    }
  }

  Future<void> blockPeer() async {
    final peerId = activePeerId;
    if (peerId == null || safetyLoading.value) return;
    final confirmed = await Get.dialog<bool>(
      AlertDialog(
        title: const Text('拉黑并结束'),
        content: const Text('之后不会再匹配到这个用户。'),
        actions: [
          TextButton(
              onPressed: () => Get.back(result: false),
              child: const Text('取消')),
          FilledButton(
              onPressed: () => Get.back(result: true), child: const Text('拉黑')),
        ],
      ),
    );
    if (confirmed != true) return;
    safetyLoading.value = true;
    try {
      await api.blockUser(peerId);
      closeSocket();
      stopMedia();
      resetCall('已拉黑对方，不会再匹配到此用户');
      await loadBlockedUsers();
    } catch (error) {
      showError(error);
    } finally {
      safetyLoading.value = false;
    }
  }

  Future<void> loadBlockedUsers() async {
    blockedLoading.value = true;
    try {
      await ensureAuth();
      blockedUsers.assignAll(await api.fetchBlockedUsers());
    } catch (error) {
      showError(error);
    } finally {
      blockedLoading.value = false;
    }
  }

  Future<void> unblockUser(String userId) async {
    try {
      await api.unblockUser(userId);
      blockedUsers.removeWhere((item) => item.user.id == userId);
      Get.snackbar('已解除拉黑', '未来可能再次匹配到对方', snackPosition: SnackPosition.BOTTOM);
    } catch (error) {
      showError(error);
    }
  }

  void switchPage(AppPage next) {
    page.value = next;
    if (next != AppPage.video) chatOpen.value = false;
    if (next == AppPage.profile) unawaited(loadBlockedUsers());
    if (next == AppPage.membership) unawaited(loadCommerceStatus());
  }

  void showError(Object error) {
    errorText = api.userMessage(error);
    Get.snackbar('操作失败', errorText, snackPosition: SnackPosition.BOTTOM);
  }
}
