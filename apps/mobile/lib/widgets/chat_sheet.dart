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
          color: const Color(0xf20a1020),
          border: Border.all(color: const Color(0xff263247)),
          borderRadius: BorderRadius.circular(24),
          boxShadow: const [
            BoxShadow(
              color: Colors.black45,
              blurRadius: 30,
              offset: Offset(0, 16),
            )
          ],
        ),
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 8, 6),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      enabled ? '文字聊天'.tr : '匹配成功后可文字聊天'.tr,
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
                  ),
                  IconButton(
                    onPressed: () => controller.chatOpen.value = false,
                    icon: const Icon(Icons.close),
                    tooltip: '关闭'.tr,
                  ),
                ],
              ),
            ),
            Expanded(
              child: controller.messages.isEmpty
                  ? Center(
                      child: Text(
                        enabled ? '开始文字聊天'.tr : '等待匹配后开始聊天'.tr,
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
                        final translated =
                            controller.translatedPreview(message.text);
                        return Align(
                          alignment: mine
                              ? Alignment.centerRight
                              : Alignment.centerLeft,
                          child: ConstrainedBox(
                            constraints: const BoxConstraints(maxWidth: 280),
                            child: DecoratedBox(
                              decoration: BoxDecoration(
                                color: mine
                                    ? const Color(0xff63e0dc)
                                    : const Color(0xff182132),
                                borderRadius: BorderRadius.only(
                                  topLeft: const Radius.circular(18),
                                  topRight: const Radius.circular(18),
                                  bottomLeft: Radius.circular(mine ? 18 : 6),
                                  bottomRight: Radius.circular(mine ? 6 : 18),
                                ),
                              ),
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 12, vertical: 10),
                                child: Column(
                                  crossAxisAlignment: mine
                                      ? CrossAxisAlignment.end
                                      : CrossAxisAlignment.start,
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    Text(
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
                                    if (translated.isNotEmpty) ...[
                                      const SizedBox(height: 5),
                                      Text(
                                        translated,
                                        style: TextStyle(
                                          color: mine
                                              ? const Color(0xaa06150c)
                                              : Colors.white70,
                                          fontSize: 12,
                                          fontWeight: FontWeight.w700,
                                        ),
                                      ),
                                    ],
                                  ],
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
                        hintText: enabled ? '输入消息...'.tr : '等待匹配后开始聊天'.tr,
                        isDense: true,
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(18),
                        ),
                      ),
                      onSubmitted: (_) => controller.sendChatMessage(),
                    ),
                  ),
                  const SizedBox(width: 8),
                  FilledButton(
                    onPressed: enabled ? controller.sendChatMessage : null,
                    style: FilledButton.styleFrom(
                      minimumSize: const Size(74, 48),
                      backgroundColor: const Color(0xff63e0dc),
                      foregroundColor: const Color(0xff07111f),
                    ),
                    child: Text(
                      '发送'.tr,
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
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
