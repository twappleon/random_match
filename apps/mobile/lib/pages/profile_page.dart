import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';

class ProfilePage extends GetView<MatchController> {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _Panel(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Text('匿名身份',
                    style:
                        TextStyle(fontSize: 22, fontWeight: FontWeight.w900)),
                const SizedBox(height: 14),
                TextField(
                  controller: controller.displayNameInput,
                  maxLength: 24,
                  decoration:
                      const InputDecoration(labelText: '昵称', counterText: ''),
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: controller.bioInput,
                  maxLength: 120,
                  maxLines: 3,
                  decoration:
                      const InputDecoration(labelText: '简介', counterText: ''),
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: controller.interestsInput,
                  decoration: const InputDecoration(labelText: '兴趣标签'),
                ),
                const SizedBox(height: 8),
                Obx(() => CheckboxListTile(
                      value: controller.ageConfirmed.value,
                      onChanged: (value) =>
                          controller.ageConfirmed.value = value ?? false,
                      contentPadding: EdgeInsets.zero,
                      title: const Text('我已满 18 岁并同意文明视讯'),
                    )),
                const SizedBox(height: 8),
                Obx(() => FilledButton(
                      onPressed: controller.savingProfile.value
                          ? null
                          : controller.saveProfile,
                      child:
                          Text(controller.savingProfile.value ? '保存中' : '保存资料'),
                    )),
              ],
            ),
          ),
          const SizedBox(height: 16),
          _Panel(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Row(
                  children: [
                    const Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('拉黑名单',
                              style: TextStyle(
                                  fontSize: 18, fontWeight: FontWeight.w900)),
                          Text('解除后未来可能再次匹配到对方',
                              style: TextStyle(color: Colors.white70)),
                        ],
                      ),
                    ),
                    Obx(() => TextButton(
                          onPressed: controller.blockedLoading.value
                              ? null
                              : controller.loadBlockedUsers,
                          child: Text(
                              controller.blockedLoading.value ? '读取中' : '刷新'),
                        )),
                  ],
                ),
                const SizedBox(height: 10),
                Obx(() {
                  if (controller.blockedUsers.isEmpty) {
                    return const Text('目前没有拉黑对象',
                        style: TextStyle(color: Colors.white70));
                  }
                  return Column(
                    children: controller.blockedUsers.map((item) {
                      final user = item.user;
                      final initial = user.displayName.trim().isNotEmpty
                          ? user.displayName
                              .trim()
                              .characters
                              .first
                              .toUpperCase()
                          : '星';
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 8),
                        child: DecoratedBox(
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.06),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: ListTile(
                            leading: CircleAvatar(
                              backgroundColor: const Color(0xff2fd276),
                              foregroundColor: const Color(0xff06150c),
                              child: Text(initial,
                                  style: const TextStyle(
                                      fontWeight: FontWeight.w900)),
                            ),
                            title: Text(user.displayName),
                            subtitle:
                                Text(user.bio.isEmpty ? '已拉黑用户' : user.bio),
                            trailing: TextButton(
                              onPressed: () => controller.unblockUser(user.id),
                              child: const Text('解除'),
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  );
                }),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _Panel extends StatelessWidget {
  const _Panel({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: const Color(0xcc111419),
        border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Material(
        color: Colors.transparent,
        child: Padding(padding: const EdgeInsets.all(16), child: child),
      ),
    );
  }
}
