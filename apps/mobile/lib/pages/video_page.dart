import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';
import '../widgets/chat_sheet.dart';
import '../widgets/peer_bar.dart';

class VideoPage extends GetView<MatchController> {
  const VideoPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Obx(() {
      final status = controller.status.value;
      final matched = status == MatchStatus.matched;
      final waiting = status == MatchStatus.waiting;
      final chatOpen = controller.chatOpen.value;
      final keyboardInset = MediaQuery.viewInsetsOf(context).bottom;
      final keyboardOpen = keyboardInset > 0;
      final chatBottom = keyboardOpen ? 12.0 : 100.0;
      final remoteVideoTick = controller.remoteVideoTick.value;

      return LayoutBuilder(
        builder: (context, constraints) {
          final chatHeight = (keyboardOpen
                  ? constraints.maxHeight - 24
                  : constraints.maxHeight * 0.48)
              .clamp(300.0, 430.0)
              .toDouble();

          return Stack(
            children: [
              Positioned.fill(
                child: _AuroraVideoBackground(
                  child: matched
                      ? RTCVideoView(
                          key: ValueKey(remoteVideoTick),
                          controller.remoteRenderer,
                          objectFit:
                              RTCVideoViewObjectFit.RTCVideoViewObjectFitCover,
                        )
                      : const SizedBox.expand(),
                ),
              ),
              const Positioned.fill(child: _AuroraScrim()),
              Positioned(
                top: 14,
                left: 14,
                right: 14,
                child: SafeArea(
                  bottom: false,
                  child: _StatsBar(stats: controller.stats.value),
                ),
              ),
              if (!matched)
                Positioned.fill(
                  top: 96,
                  bottom: 212,
                  child: Center(
                    child: SingleChildScrollView(
                      padding: const EdgeInsets.symmetric(horizontal: 20),
                      child: _AuroraStateCard(
                        waiting: waiting,
                        title: waiting ? '正在寻找新朋友' : '今晚遇见新朋友',
                        subtitle: waiting
                            ? '保持页面开启，匹配成功后会自动进入视讯。'
                            : '选择视讯或文字，随时开始随机匹配。',
                      ),
                    ),
                  ),
                ),
              if (matched && !chatOpen)
                Positioned(
                  right: 16,
                  bottom: 112,
                  width: 118,
                  height: 172,
                  child: _LocalPreview(renderer: controller.localRenderer),
                ),
              if (matched && !chatOpen && !controller.peerCardHidden.value)
                const Positioned(
                    left: 12, right: 12, bottom: 100, child: PeerBar()),
              if (chatOpen)
                Positioned(
                  left: 12,
                  right: 12,
                  bottom: chatBottom,
                  height: chatHeight,
                  child: const ChatSheet(),
                ),
              if (!chatOpen)
                Positioned(
                  left: 12,
                  right: 12,
                  bottom: 18,
                  child: SafeArea(
                    top: false,
                    child: _ActionDock(
                      controller: controller,
                      status: status,
                    ),
                  ),
                ),
            ],
          );
        },
      );
    });
  }
}

class _AuroraVideoBackground extends StatelessWidget {
  const _AuroraVideoBackground({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: const BoxDecoration(
        gradient: RadialGradient(
          center: Alignment(-0.72, -0.72),
          radius: 1.08,
          colors: [Color(0x677c5cff), Color(0x00000000)],
        ),
      ),
      child: DecoratedBox(
        decoration: const BoxDecoration(
          gradient: RadialGradient(
            center: Alignment(0.8, -0.5),
            radius: 0.92,
            colors: [Color(0x5520c8ff), Color(0x00000000)],
          ),
        ),
        child: DecoratedBox(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                Color(0xff151735),
                Color(0xff080a16),
                Color(0xff050608),
              ],
            ),
          ),
          child: child,
        ),
      ),
    );
  }
}

class _AuroraScrim extends StatelessWidget {
  const _AuroraScrim();

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: DecoratedBox(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [
              Colors.black.withValues(alpha: 0.02),
              Colors.black.withValues(alpha: 0.16),
              Colors.black.withValues(alpha: 0.5),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatsBar extends StatelessWidget {
  const _StatsBar({required this.stats});

  final RuntimeStats stats;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(child: _AuroraStat(value: stats.online, label: '在线')),
        const SizedBox(width: 8),
        Expanded(child: _AuroraStat(value: stats.waiting, label: '等待')),
        const SizedBox(width: 8),
        Expanded(child: _AuroraStat(value: stats.chatting, label: '聊天')),
      ],
    );
  }
}

class _LocalPreview extends StatelessWidget {
  const _LocalPreview({required this.renderer});

  final RTCVideoRenderer renderer;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: DecoratedBox(
        decoration: BoxDecoration(
          border: Border.all(color: const Color(0x66c4aaff)),
          boxShadow: const [
            BoxShadow(color: Color(0x4d7c5cff), blurRadius: 42),
            BoxShadow(color: Colors.black45, blurRadius: 30),
          ],
        ),
        child: ColoredBox(
          color: const Color(0xff202632),
          child: RTCVideoView(renderer, mirror: true),
        ),
      ),
    );
  }
}

class _ActionDock extends StatelessWidget {
  const _ActionDock({
    required this.controller,
    required this.status,
  });

  final MatchController controller;
  final MatchStatus status;

  @override
  Widget build(BuildContext context) {
    final loading = controller.loading.value;
    final leaving = controller.leaving.value;
    final matched = status == MatchStatus.matched;
    final waiting = status == MatchStatus.waiting;

    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
        borderRadius: BorderRadius.circular(12),
        color: const Color(0xd9080a13),
        boxShadow: const [
          BoxShadow(color: Colors.black54, blurRadius: 34),
          BoxShadow(color: Color(0x267c5cff), blurRadius: 44),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(10),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                Expanded(
                  child: _AuroraButton(
                    onPressed: controller.toggleChat,
                    icon: controller.chatOpen.value
                        ? Icons.keyboard_arrow_down_rounded
                        : Icons.chat_bubble_outline_rounded,
                    label: controller.chatOpen.value ? '收起文字' : '文字',
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _AuroraButton(
                    onPressed: matched &&
                            controller.activePeerId != null &&
                            !controller.safetyLoading.value &&
                            !leaving
                        ? controller.blockPeer
                        : null,
                    icon: Icons.block_rounded,
                    label: controller.safetyLoading.value ? '处理中' : '拉黑',
                    dangerSoft: true,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            _AuroraButton(
              onPressed:
                  loading || waiting || matched ? null : controller.startMatch,
              icon: Icons.auto_awesome_rounded,
              label: matched
                  ? '已连线'
                  : loading || waiting
                      ? '匹配中'
                      : '随机匹配',
              primary: true,
            ),
            const SizedBox(height: 8),
            _AuroraButton(
              onPressed: leaving || status == MatchStatus.idle
                  ? null
                  : controller.leaveCall,
              icon: Icons.logout_rounded,
              label: leaving ? '退出中' : '退出',
              danger: true,
            ),
          ],
        ),
      ),
    );
  }
}

class _AuroraStat extends StatelessWidget {
  const _AuroraStat({required this.value, required this.label});

  final int value;
  final String label;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
        borderRadius: BorderRadius.circular(8),
        color: Colors.white.withValues(alpha: 0.09),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              '$value',
              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w900),
            ),
            Text(
              label,
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.72),
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AuroraStateCard extends StatelessWidget {
  const _AuroraStateCard({
    required this.waiting,
    required this.title,
    required this.subtitle,
  });

  final bool waiting;
  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.white.withValues(alpha: 0.15)),
        borderRadius: BorderRadius.circular(8),
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            const Color(0xff7c5cff).withValues(alpha: 0.2),
            const Color(0xff20c8ff).withValues(alpha: 0.1),
            Colors.white.withValues(alpha: 0.05),
          ],
        ),
        color: const Color(0xbb141827),
        boxShadow: const [
          BoxShadow(color: Colors.black54, blurRadius: 52),
          BoxShadow(color: Color(0x267c5cff), blurRadius: 70),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            DecoratedBox(
              decoration: BoxDecoration(
                border: Border.all(
                    color: (waiting
                            ? const Color(0xff20c8ff)
                            : const Color(0xffc4aaff))
                        .withValues(alpha: 0.42)),
                borderRadius: BorderRadius.circular(999),
                color: (waiting
                        ? const Color(0xff20c8ff)
                        : const Color(0xff7c5cff))
                    .withValues(alpha: 0.18),
              ),
              child: Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 11, vertical: 6),
                child: Text(
                  waiting ? 'LIVE MATCH' : 'AURORA READY',
                  style: const TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                    letterSpacing: 1.4,
                  ),
                ),
              ),
            ),
            const SizedBox(height: 14),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 30, fontWeight: FontWeight.w900),
            ),
            const SizedBox(height: 10),
            Text(
              subtitle,
              textAlign: TextAlign.center,
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.7),
                fontSize: 14,
                fontWeight: FontWeight.w700,
                height: 1.45,
              ),
            ),
            const SizedBox(height: 14),
            const Wrap(
              spacing: 8,
              runSpacing: 8,
              alignment: WrapAlignment.center,
              children: [
                _AuroraChip('随机视讯'),
                _AuroraChip('快速连接'),
                _AuroraChip('安全操作'),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _AuroraChip extends StatelessWidget {
  const _AuroraChip(this.label);

  final String label;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(999),
        color: Colors.white.withValues(alpha: 0.11),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
        child: Text(
          label,
          style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w800),
        ),
      ),
    );
  }
}

class _AuroraButton extends StatelessWidget {
  const _AuroraButton({
    required this.label,
    this.onPressed,
    this.icon,
    this.primary = false,
    this.danger = false,
    this.dangerSoft = false,
  });

  final String label;
  final VoidCallback? onPressed;
  final IconData? icon;
  final bool primary;
  final bool danger;
  final bool dangerSoft;

  @override
  Widget build(BuildContext context) {
    final background = primary
        ? null
        : danger
            ? const Color(0xffb84d5b)
            : dangerSoft
                ? const Color(0x3dff5c74)
                : const Color(0xff1c2130);
    return SizedBox(
      width: double.infinity,
      height: 48,
      child: DecoratedBox(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(8),
          gradient: primary
              ? const LinearGradient(
                  colors: [Color(0xff7c5cff), Color(0xff20c8ff)],
                )
              : null,
          boxShadow: primary
              ? const [BoxShadow(color: Color(0x477c5cff), blurRadius: 28)]
              : null,
        ),
        child: FilledButton(
          style: FilledButton.styleFrom(
            backgroundColor: primary ? Colors.transparent : background,
            disabledBackgroundColor: const Color(0xff1c2130),
            shadowColor: Colors.transparent,
            foregroundColor: Colors.white,
            disabledForegroundColor: Colors.white54,
            shape:
                RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          ),
          onPressed: onPressed,
          child: FittedBox(
            fit: BoxFit.scaleDown,
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (icon != null) ...[
                  Icon(icon, size: 18),
                  const SizedBox(width: 6),
                ],
                Text(
                  label,
                  style: const TextStyle(fontWeight: FontWeight.w900),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
