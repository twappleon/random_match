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
          color: const Color(0xa80d0f13),
          border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
          borderRadius: BorderRadius.circular(8),
          boxShadow: const [BoxShadow(color: Colors.black38, blurRadius: 24)],
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 8),
          child: Row(
            children: [
              CircleAvatar(
                radius: 17,
                backgroundColor: const Color(0xff2fd276),
                foregroundColor: const Color(0xff06150c),
                child: Text(
                  initial,
                  style: const TextStyle(fontWeight: FontWeight.w900),
                ),
              ),
              const SizedBox(width: 9),
              Expanded(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    Text(
                      bio,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: Colors.white.withValues(alpha: 0.68),
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ),
              ),
              if (peer != null) ...[
                const SizedBox(width: 4),
                IconButton(
                  visualDensity: VisualDensity.compact,
                  tooltip: '关注',
                  onPressed: () => controller.toggleFollow(peer),
                  icon: Icon(
                    controller.followedUserIds.contains(peer.id)
                        ? Icons.favorite_rounded
                        : Icons.favorite_border_rounded,
                    size: 20,
                  ),
                ),
              ],
              IconButton(
                visualDensity: VisualDensity.compact,
                onPressed: () => controller.peerCardHidden.value = true,
                icon: const Icon(Icons.close, size: 20),
                tooltip: '隐藏',
              ),
            ],
          ),
        ),
      );
    });
  }
}
