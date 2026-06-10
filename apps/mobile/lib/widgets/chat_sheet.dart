import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';

class ChatSheet extends GetView<MatchController> {
  const ChatSheet({super.key});

  @override
  Widget build(BuildContext context) {
    return Obx(() {
      final enabled = controller.status.value == MatchStatus.matched &&
          controller.activePeerId != null;
      return DecoratedBox(
        decoration: BoxDecoration(
          color: const Color(0xee0d0f13),
          border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
          borderRadius: BorderRadius.circular(8),
          boxShadow: const [BoxShadow(color: Colors.black45, blurRadius: 28)],
        ),
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 8, 8, 4),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      enabled ? '文字聊天' : '匹配成功后可文字聊天',
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
                  ),
                  IconButton(
                    onPressed: () => controller.chatOpen.value = false,
                    icon: const Icon(Icons.close),
                    tooltip: '关闭',
                  ),
                ],
              ),
            ),
            Expanded(
              child: controller.messages.isEmpty
                  ? Center(
                      child: Text(
                        enabled ? '开始文字聊天' : '等待匹配后开始聊天',
                        style: const TextStyle(
                            color: Colors.white70, fontWeight: FontWeight.w700),
                      ),
                    )
                  : ListView.separated(
                      controller: controller.chatScroll,
                      padding: const EdgeInsets.all(12),
                      itemCount: controller.messages.length,
                      separatorBuilder: (_, __) => const SizedBox(height: 8),
                      itemBuilder: (context, index) {
                        final message = controller.messages[index];
                        final mine = message.sender == ChatSender.self;
                        return Align(
                          alignment: mine
                              ? Alignment.centerRight
                              : Alignment.centerLeft,
                          child: ConstrainedBox(
                            constraints: const BoxConstraints(maxWidth: 280),
                            child: DecoratedBox(
                              decoration: BoxDecoration(
                                color: mine
                                    ? const Color(0xff2fd276)
                                    : Colors.white.withValues(alpha: 0.12),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 10, vertical: 8),
                                child: Text(
                                  message.text,
                                  style: TextStyle(
                                    color: mine
                                        ? const Color(0xff06150c)
                                        : Colors.white,
                                    fontWeight: mine
                                        ? FontWeight.w700
                                        : FontWeight.w500,
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
                      controller: controller.chatInput,
                      enabled: enabled,
                      maxLength: 500,
                      minLines: 1,
                      maxLines: 3,
                      decoration: InputDecoration(
                        counterText: '',
                        hintText: enabled ? '输入消息...' : '等待匹配后开始聊天',
                        isDense: true,
                        border: const OutlineInputBorder(),
                      ),
                      onSubmitted: (_) => controller.sendChatMessage(),
                    ),
                  ),
                  const SizedBox(width: 8),
                  FilledButton(
                    onPressed: enabled ? controller.sendChatMessage : null,
                    child: const Text('发送'),
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    });
  }
}
