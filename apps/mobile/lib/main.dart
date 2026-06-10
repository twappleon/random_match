import 'dart:convert';

import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:http/http.dart' as http;
import 'package:web_socket_channel/web_socket_channel.dart';

const apiBase = String.fromEnvironment(
  'API_BASE',
  defaultValue: 'http://localhost:8080',
);

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await Firebase.initializeApp();
  } catch (_) {
    // Native Firebase files are added later by the real Firebase project.
  }
  runApp(const RandomMatchApp());
}

class RandomMatchApp extends StatelessWidget {
  const RandomMatchApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Random Match',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xff26c06f),
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
      ),
      home: const MatchPage(),
    );
  }
}

class MatchPage extends StatefulWidget {
  const MatchPage({super.key});

  @override
  State<MatchPage> createState() => _MatchPageState();
}

class _MatchPageState extends State<MatchPage> {
  final _localRenderer = RTCVideoRenderer();
  final _remoteRenderer = RTCVideoRenderer();
  final _chatInput = TextEditingController();
  final _chatScroll = ScrollController();
  WebSocketChannel? _socket;
  RTCPeerConnection? _peer;
  MediaStream? _localStream;
  String? _token;
  String? _activePeerId;
  String? _activeRoomId;
  String _mode = 'video';
  String _status = 'idle';
  bool _loading = false;
  final List<_ChatMessage> _messages = [];

  @override
  void initState() {
    super.initState();
    _localRenderer.initialize();
    _remoteRenderer.initialize();
  }

  @override
  void dispose() {
    _socket?.sink.close();
    _peer?.close();
    _localStream?.dispose();
    _chatInput.dispose();
    _chatScroll.dispose();
    _localRenderer.dispose();
    _remoteRenderer.dispose();
    super.dispose();
  }

  Future<void> _startMatch() async {
    setState(() => _loading = true);
    try {
      await _ensureAuth();
      await _openMedia();
      _openSocket();
      setState(() {
        _messages.clear();
        _activePeerId = null;
        _activeRoomId = null;
      });
      final response = await http.post(
        Uri.parse('$apiBase/api/v1/match/join'),
        headers: {
          'Authorization': 'Bearer $_token',
          'Content-Type': 'application/json',
        },
        body: jsonEncode({'mode': _mode, 'region': 'global'}),
      );
      if (response.statusCode != 200 && response.statusCode != 202) {
        throw StateError('match failed');
      }
      final body = jsonDecode(response.body) as Map<String, dynamic>;
      setState(() {
        _status = body['status'] as String? ?? 'waiting';
        _activePeerId = body['peerId'] as String?;
        _activeRoomId = body['roomId'] as String?;
      });
      if (body['initiator'] == true && body['peerId'] is String) {
        await _createPeer(body['peerId'] as String);
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _ensureAuth() async {
    if (_token != null) return;
    final response =
        await http.post(Uri.parse('$apiBase/api/v1/auth/anonymous'));
    if (response.statusCode != 200) throw StateError('login failed');
    final body = jsonDecode(response.body) as Map<String, dynamic>;
    _token = body['token'] as String;
  }

  Future<void> _openMedia() async {
    _localStream?.dispose();
    _localStream = await navigator.mediaDevices.getUserMedia({
      'audio': true,
      'video': _mode == 'video',
    });
    _localRenderer.srcObject = _localStream;
  }

  void _openSocket() {
    if (_socket != null) return;
    final uri = Uri.parse(apiBase).replace(
      scheme: apiBase.startsWith('https') ? 'wss' : 'ws',
      path: '/api/v1/ws',
      queryParameters: {'token': _token},
    );
    _socket = WebSocketChannel.connect(uri);
    _socket!.stream.listen((event) async {
      final msg = jsonDecode(event as String) as Map<String, dynamic>;
      switch (msg['type']) {
        case 'matched':
          setState(() {
            _status = 'matched';
            _activePeerId = msg['peerId'] as String?;
            _activeRoomId = msg['roomId'] as String?;
          });
          if (msg['initiator'] == true) {
            await _createPeer(msg['peerId'] as String);
          }
          break;
        case 'offer':
          await _acceptOffer(msg['peerId'] as String, msg['data']);
          break;
        case 'answer':
          await _peer?.setRemoteDescription(_descriptionFrom(msg['data']));
          break;
        case 'candidate':
          await _peer?.addCandidate(_candidateFrom(msg['data']));
          break;
        case 'chat-message':
          _receiveChatMessage(msg['data']);
          break;
        case 'peer-left':
          setState(() {
            _status = 'idle';
            _activePeerId = null;
            _activeRoomId = null;
            _messages.clear();
          });
          break;
      }
    });
  }

  Future<void> _createPeer(String peerId) async {
    _peer = await _buildPeer(peerId);
    for (final track in _localStream?.getTracks() ?? <MediaStreamTrack>[]) {
      await _peer!.addTrack(track, _localStream!);
    }
    final offer = await _peer!.createOffer();
    await _peer!.setLocalDescription(offer);
    _send({'type': 'offer', 'peerId': peerId, 'data': offer.toMap()});
  }

  Future<void> _acceptOffer(String peerId, dynamic offer) async {
    _peer = await _buildPeer(peerId);
    for (final track in _localStream?.getTracks() ?? <MediaStreamTrack>[]) {
      await _peer!.addTrack(track, _localStream!);
    }
    await _peer!.setRemoteDescription(_descriptionFrom(offer));
    final answer = await _peer!.createAnswer();
    await _peer!.setLocalDescription(answer);
    _send({'type': 'answer', 'peerId': peerId, 'data': answer.toMap()});
  }

  Future<RTCPeerConnection> _buildPeer(String peerId) async {
    final pc = await createPeerConnection({
      'iceServers': [
        {'urls': 'stun:stun.l.google.com:19302'},
      ],
    });
    pc.onIceCandidate = (candidate) {
      _send({'type': 'candidate', 'peerId': peerId, 'data': candidate.toMap()});
    };
    pc.onTrack = (event) {
      if (event.streams.isNotEmpty) {
        _remoteRenderer.srcObject = event.streams.first;
      }
    };
    return pc;
  }

  RTCSessionDescription _descriptionFrom(dynamic data) {
    final map = Map<String, dynamic>.from(data as Map);
    return RTCSessionDescription(map['sdp'] as String?, map['type'] as String?);
  }

  RTCIceCandidate _candidateFrom(dynamic data) {
    final map = Map<String, dynamic>.from(data as Map);
    return RTCIceCandidate(
      map['candidate'] as String?,
      map['sdpMid'] as String?,
      map['sdpMLineIndex'] as int?,
    );
  }

  void _send(Map<String, dynamic> message) {
    _socket?.sink.add(jsonEncode(message));
  }

  void _sendChatMessage() {
    final peerId = _activePeerId;
    final text = _chatInput.text.trim();
    if (_status != 'matched' || peerId == null || text.isEmpty) return;

    final message = _ChatMessage(
      id: DateTime.now().microsecondsSinceEpoch.toString(),
      sender: _ChatSender.self,
      text: text.length > 500 ? text.substring(0, 500) : text,
      createdAt: DateTime.now(),
    );
    setState(() {
      _messages.add(message);
      _chatInput.clear();
    });
    _scrollChatToBottom();
    _send({
      'type': 'chat-message',
      'peerId': peerId,
      if (_activeRoomId != null) 'roomId': _activeRoomId,
      'data': {
        'id': message.id,
        'text': message.text,
        'createdAt': message.createdAt.toIso8601String(),
      },
    });
  }

  void _receiveChatMessage(dynamic data) {
    if (data is! Map) return;
    final rawText = data['text'];
    if (rawText is! String) return;
    final text = rawText.trim();
    if (text.isEmpty) return;
    setState(() {
      _messages.add(_ChatMessage(
        id: data['id'] as String? ??
            DateTime.now().microsecondsSinceEpoch.toString(),
        sender: _ChatSender.peer,
        text: text.length > 500 ? text.substring(0, 500) : text,
        createdAt: DateTime.tryParse(data['createdAt'] as String? ?? '') ??
            DateTime.now(),
      ));
    });
    _scrollChatToBottom();
  }

  void _scrollChatToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_chatScroll.hasClients) return;
      _chatScroll.animateTo(
        _chatScroll.position.maxScrollExtent,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          Positioned.fill(
            child: Container(
              color: const Color(0xff101114),
              child: RTCVideoView(_remoteRenderer,
                  objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitCover),
            ),
          ),
          if (_status != 'matched')
            Center(
              child: Text(
                _status == 'waiting' ? '正在寻找在线用户' : '准备开始随机交友',
                textAlign: TextAlign.center,
                style:
                    const TextStyle(fontSize: 28, fontWeight: FontWeight.w800),
              ),
            ),
          Positioned(
            right: 16,
            bottom: 92,
            width: 118,
            height: 172,
            child: ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: ColoredBox(
                color: const Color(0xff20232a),
                child: RTCVideoView(_localRenderer, mirror: true),
              ),
            ),
          ),
          if (_status == 'matched')
            Positioned(
              left: 12,
              right: 12,
              bottom: 280,
              height: 220,
              child: _ChatPanel(
                messages: _messages,
                controller: _chatInput,
                scrollController: _chatScroll,
                enabled: _activePeerId != null,
                onSend: _sendChatMessage,
              ),
            ),
          Positioned(
            left: 12,
            right: 12,
            bottom: 18,
            child: SafeArea(
              child: Row(
                children: [
                  _ModeButton(
                      label: '视讯',
                      active: _mode == 'video',
                      onTap: () => setState(() => _mode = 'video')),
                  const SizedBox(width: 8),
                  _ModeButton(
                      label: '语音',
                      active: _mode == 'voice',
                      onTap: () => setState(() => _mode = 'voice')),
                  const SizedBox(width: 8),
                  Expanded(
                    child: FilledButton(
                      onPressed: _loading ? null : _startMatch,
                      child: Text(_loading ? '匹配中' : '随机匹配'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

enum _ChatSender { self, peer }

class _ChatMessage {
  const _ChatMessage({
    required this.id,
    required this.sender,
    required this.text,
    required this.createdAt,
  });

  final String id;
  final _ChatSender sender;
  final String text;
  final DateTime createdAt;
}

class _ChatPanel extends StatelessWidget {
  const _ChatPanel({
    required this.messages,
    required this.controller,
    required this.scrollController,
    required this.enabled,
    required this.onSend,
  });

  final List<_ChatMessage> messages;
  final TextEditingController controller;
  final ScrollController scrollController;
  final bool enabled;
  final VoidCallback onSend;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: const Color(0xcc0c0d10),
        border: Border.all(color: Colors.white.withValues(alpha: 0.14)),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          Expanded(
            child: messages.isEmpty
                ? const Center(
                    child: Text(
                      '开始文字聊天',
                      style: TextStyle(
                          color: Colors.white70, fontWeight: FontWeight.w700),
                    ),
                  )
                : ListView.separated(
                    controller: scrollController,
                    padding: const EdgeInsets.all(12),
                    itemCount: messages.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 8),
                    itemBuilder: (context, index) {
                      final message = messages[index];
                      final mine = message.sender == _ChatSender.self;
                      return Align(
                        alignment:
                            mine ? Alignment.centerRight : Alignment.centerLeft,
                        child: ConstrainedBox(
                          constraints: const BoxConstraints(maxWidth: 260),
                          child: DecoratedBox(
                            decoration: BoxDecoration(
                              color: mine
                                  ? const Color(0xff26c06f)
                                  : Colors.white.withValues(alpha: 0.11),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Padding(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 10, vertical: 8),
                              child: Text(
                                message.text,
                                style: TextStyle(
                                  color: mine
                                      ? const Color(0xff07150d)
                                      : Colors.white,
                                  fontWeight:
                                      mine ? FontWeight.w700 : FontWeight.w500,
                                ),
                              ),
                            ),
                          ),
                        ),
                      );
                    },
                  ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(10, 8, 10, 10),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: controller,
                    enabled: enabled,
                    maxLength: 500,
                    minLines: 1,
                    maxLines: 3,
                    decoration: const InputDecoration(
                      counterText: '',
                      hintText: '输入消息...',
                      isDense: true,
                      border: OutlineInputBorder(),
                    ),
                    onSubmitted: (_) => onSend(),
                  ),
                ),
                const SizedBox(width: 8),
                FilledButton(
                  onPressed: enabled ? onSend : null,
                  child: const Text('发送'),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ModeButton extends StatelessWidget {
  const _ModeButton({
    required this.label,
    required this.active,
    required this.onTap,
  });

  final String label;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: OutlinedButton(
        onPressed: onTap,
        style: OutlinedButton.styleFrom(
          backgroundColor:
              active ? const Color(0xffe8f7ef) : const Color(0xff2a2d34),
          foregroundColor: active ? const Color(0xff12231a) : Colors.white,
          side: BorderSide.none,
        ),
        child: Text(label),
      ),
    );
  }
}
