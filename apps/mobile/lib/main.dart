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
  WebSocketChannel? _socket;
  RTCPeerConnection? _peer;
  MediaStream? _localStream;
  String? _token;
  String _mode = 'video';
  String _status = 'idle';
  bool _loading = false;

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
      setState(() => _status = body['status'] as String? ?? 'waiting');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _ensureAuth() async {
    if (_token != null) return;
    final response = await http.post(Uri.parse('$apiBase/api/v1/auth/anonymous'));
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
          setState(() => _status = 'matched');
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
      if (event.streams.isNotEmpty) _remoteRenderer.srcObject = event.streams.first;
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          Positioned.fill(
            child: Container(
              color: const Color(0xff101114),
              child: RTCVideoView(_remoteRenderer, objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitCover),
            ),
          ),
          if (_status != 'matched')
            Center(
              child: Text(
                _status == 'waiting' ? '正在寻找在线用户' : '准备开始随机交友',
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 28, fontWeight: FontWeight.w800),
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
          Positioned(
            left: 12,
            right: 12,
            bottom: 18,
            child: SafeArea(
              child: Row(
                children: [
                  _ModeButton(label: '视讯', active: _mode == 'video', onTap: () => setState(() => _mode = 'video')),
                  const SizedBox(width: 8),
                  _ModeButton(label: '语音', active: _mode == 'voice', onTap: () => setState(() => _mode = 'voice')),
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
          backgroundColor: active ? const Color(0xffe8f7ef) : const Color(0xff2a2d34),
          foregroundColor: active ? const Color(0xff12231a) : Colors.white,
          side: BorderSide.none,
        ),
        child: Text(label),
      ),
    );
  }
}
