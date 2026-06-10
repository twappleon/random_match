import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';

class PeerBar extends GetView<MatchController> {
  const PeerBar({super.key});

  @override
  Widget build(BuildContext context) {
    return Obx(() {
      final peer = controller.peerProfile.value;
      final name = peer?.displayName ?? '对方资料载入中';
      final bio = peer?.bio.isNotEmpty == true ? peer!.bio : '对方暂时没有填写简介';
      final initial = name.trim().isNotEmpty
          ? name.trim().characters.first.toUpperCase()
          : '星';
      return DecoratedBox(
        decoration: BoxDecoration(
          color: const Color(0xbb0d0f13),
          border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
          borderRadius: BorderRadius.circular(8),
          boxShadow: const [BoxShadow(color: Colors.black38, blurRadius: 24)],
        ),
        child: Padding(
          padding: const EdgeInsets.all(10),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Align(
                alignment: Alignment.centerRight,
                child: IconButton(
                  visualDensity: VisualDensity.compact,
                  onPressed: () => controller.peerCardHidden.value = true,
                  icon: const Icon(Icons.close),
                  tooltip: '隐藏',
                ),
              ),
              Row(
                children: [
                  CircleAvatar(
                    backgroundColor: const Color(0xff2fd276),
                    foregroundColor: const Color(0xff06150c),
                    child: Text(initial,
                        style: const TextStyle(fontWeight: FontWeight.w900)),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(name,
                            style: const TextStyle(
                                fontSize: 16, fontWeight: FontWeight.w900)),
                        Text(
                          bio,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(color: Colors.white70),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: controller.safetyLoading.value
                          ? null
                          : controller.reportPeer,
                      child: Text(
                          controller.reportedPeerId == controller.activePeerId
                              ? '已举报'
                              : '举报'),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: FilledButton(
                      style: FilledButton.styleFrom(
                          backgroundColor: const Color(0xffe55757)),
                      onPressed: controller.safetyLoading.value
                          ? null
                          : controller.blockPeer,
                      child: const Text('拉黑'),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      );
    });
  }
}
