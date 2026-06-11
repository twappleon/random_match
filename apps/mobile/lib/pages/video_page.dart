import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';
import '../widgets/chat_sheet.dart';
import '../widgets/peer_bar.dart';
import '../widgets/status_pill.dart';

class VideoPage extends GetView<MatchController> {
  const VideoPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        Positioned.fill(
          child: ColoredBox(
            color: const Color(0xff101114),
            child: RTCVideoView(
              controller.remoteRenderer,
              objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitCover,
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
                Expanded(child: StatusPill(label: '在线 ${stats.online}')),
                const SizedBox(width: 8),
                Expanded(child: StatusPill(label: '等待 ${stats.waiting}')),
                const SizedBox(width: 8),
                Expanded(child: StatusPill(label: '聊天 ${stats.chatting}')),
              ],
            );
          }),
        ),
        Obx(() {
          if (controller.status.value == MatchStatus.matched) {
            return const SizedBox.shrink();
          }
          final text = controller.status.value == MatchStatus.waiting
              ? '正在寻找视讯用户'
              : '准备开始随机交友';
          return Center(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Text(
                text,
                textAlign: TextAlign.center,
                style:
                    const TextStyle(fontSize: 28, fontWeight: FontWeight.w900),
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
            child: ColoredBox(
              color: const Color(0xff20232a),
              child: RTCVideoView(controller.localRenderer, mirror: true),
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
              return Row(
                children: [
                  Expanded(
                    child: FilledButton.tonal(
                      onPressed: controller.toggleChat,
                      child: Text(controller.chatOpen.value ? '收起文字' : '文字'),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: FilledButton.tonalIcon(
                      onPressed: controller.status.value ==
                                  MatchStatus.matched &&
                              controller.activePeerId != null &&
                              !controller.safetyLoading.value &&
                              !leaving
                          ? controller.blockPeer
                          : null,
                      icon: const Icon(Icons.block, size: 18),
                      label: Text(
                          controller.safetyLoading.value ? '处理中' : '拉黑'),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    flex: 2,
                    child: FilledButton(
                      onPressed: loading ||
                              controller.status.value == MatchStatus.waiting
                          ? null
                          : controller.startMatch,
                      child: Text(loading ? '匹配中' : '随机匹配'),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: FilledButton(
                      style: FilledButton.styleFrom(
                          backgroundColor: const Color(0xffe55757)),
                      onPressed:
                          leaving || controller.status.value == MatchStatus.idle
                              ? null
                              : controller.leaveCall,
                      child: Text(leaving ? '退出中' : '退出'),
                    ),
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
