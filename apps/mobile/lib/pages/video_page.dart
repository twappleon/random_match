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
    return Stack(
      children: [
        Positioned.fill(
          child: DecoratedBox(
            decoration: const BoxDecoration(
              gradient: RadialGradient(
                center: Alignment(-0.7, -0.72),
                radius: 1.08,
                colors: [Color(0x5c7c5cff), Color(0x00000000)],
              ),
            ),
            child: DecoratedBox(
              decoration: const BoxDecoration(
                gradient: RadialGradient(
                  center: Alignment(0.78, -0.52),
                  radius: 0.92,
                  colors: [Color(0x4520c8ff), Color(0x00000000)],
                ),
              ),
              child: DecoratedBox(
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [
                      Color(0xff111427),
                      Color(0xff070914),
                      Color(0xff07070b),
                    ],
                  ),
                ),
                child: RTCVideoView(
                  controller.remoteRenderer,
                  objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitCover,
                ),
              ),
            ),
          ),
        ),
        Positioned.fill(
          child: IgnorePointer(
            child: DecoratedBox(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    Colors.black.withValues(alpha: 0.02),
                    Colors.black.withValues(alpha: 0.18),
                    Colors.black.withValues(alpha: 0.46),
                  ],
                ),
              ),
            ),
          ),
        ),
        Positioned(
          top: 14,
          left: 14,
          right: 14,
          child: Obx(() {
            final stats = controller.stats.value;
            return Row(
              children: [
                Expanded(child: _AuroraStat(value: stats.online, label: '在线')),
                const SizedBox(width: 8),
                Expanded(child: _AuroraStat(value: stats.waiting, label: '等待')),
                const SizedBox(width: 8),
                Expanded(
                    child: _AuroraStat(value: stats.chatting, label: '聊天')),
              ],
            );
          }),
        ),
        Obx(() {
          if (controller.status.value == MatchStatus.matched) {
            return const SizedBox.shrink();
          }
          final waiting = controller.status.value == MatchStatus.waiting;
          return Positioned.fill(
            top: 74,
            bottom: 244,
            child: Center(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: _AuroraStateCard(
                  waiting: waiting,
                  title: waiting ? '正在寻找新朋友' : '今晚遇见新朋友',
                  subtitle:
                      waiting ? '保持页面开启，匹配成功后会自动进入视讯。' : '更强的氛围感，快速连接在线用户。',
                ),
              ),
            ),
          );
        }),
        Positioned(
          right: 16,
          bottom: 112,
          width: 118,
          height: 172,
          child: ClipRRect(
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
                child: RTCVideoView(controller.localRenderer, mirror: true),
              ),
            ),
          ),
        ),
        Obx(() {
          final show = controller.status.value == MatchStatus.matched &&
              !controller.chatOpen.value &&
              !controller.peerCardHidden.value;
          if (!show) return const SizedBox.shrink();
          return const Positioned(
              left: 12, right: 12, bottom: 100, child: PeerBar());
        }),
        Obx(() {
          if (!controller.chatOpen.value) return const SizedBox.shrink();
          return const Positioned(
              left: 12,
              right: 12,
              bottom: 100,
              height: 320,
              child: ChatSheet());
        }),
        Obx(() {
          if (controller.chatOpen.value) return const SizedBox.shrink();
          return Positioned(
            left: 12,
            bottom: 100,
            child: FilledButton.tonalIcon(
              onPressed: controller.toggleChat,
              icon: const Icon(Icons.chat_bubble_outline),
              label: const Text('文字'),
            ),
          );
        }),
        Positioned(
          left: 12,
          right: 12,
          bottom: 18,
          child: SafeArea(
            top: false,
            child: Obx(() {
              final loading = controller.loading.value;
              final leaving = controller.leaving.value;
              return Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: _AuroraButton(
                          onPressed: controller.toggleChat,
                          label: controller.chatOpen.value ? '收起文字' : '文字',
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: _AuroraButton(
                          onPressed:
                              controller.status.value == MatchStatus.matched &&
                                      controller.activePeerId != null &&
                                      !controller.safetyLoading.value &&
                                      !leaving
                                  ? controller.blockPeer
                                  : null,
                          label: controller.safetyLoading.value ? '处理中' : '拉黑',
                          dangerSoft: true,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  _AuroraButton(
                    onPressed: loading ||
                            controller.status.value == MatchStatus.waiting
                        ? null
                        : controller.startMatch,
                    label: loading ? '匹配中' : '随机匹配',
                    primary: true,
                  ),
                  const SizedBox(height: 8),
                  _AuroraButton(
                    onPressed:
                        leaving || controller.status.value == MatchStatus.idle
                            ? null
                            : controller.leaveCall,
                    label: leaving ? '退出中' : '退出',
                    danger: true,
                  ),
                ],
              );
            }),
          ),
        ),
      ],
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
    this.primary = false,
    this.danger = false,
    this.dangerSoft = false,
  });

  final String label;
  final VoidCallback? onPressed;
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
          child:
              Text(label, style: const TextStyle(fontWeight: FontWeight.w900)),
        ),
      ),
    );
  }
}
