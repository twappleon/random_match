import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({super.key});

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  final _scrollController = ScrollController();
  final _ageKey = GlobalKey();
  late final MatchController controller;
  Worker? _agePromptWorker;

  @override
  void initState() {
    super.initState();
    controller = Get.find<MatchController>();
    _agePromptWorker = ever<int>(controller.agePromptRevision, (_) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _scrollToAgeCheck());
    });
  }

  @override
  void dispose() {
    _agePromptWorker?.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _scrollToAgeCheck() {
    final context = _ageKey.currentContext;
    if (context == null) return;
    Scrollable.ensureVisible(
      context,
      duration: const Duration(milliseconds: 420),
      curve: Curves.easeOutCubic,
      alignment: 0.36,
    );
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: ListView(
        controller: _scrollController,
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
                const SizedBox(height: 10),
                const _InterestPicker(),
                const SizedBox(height: 10),
                Obx(() => DropdownButtonFormField<String>(
                      key: ValueKey(
                          'profile-region-${controller.selectedRegion.value}'),
                      initialValue: controller.selectedRegion.value,
                      decoration: const InputDecoration(
                        labelText: '常用地区',
                        prefixIcon: Icon(Icons.public_rounded),
                      ),
                      items: const [
                        DropdownMenuItem(value: 'global', child: Text('全球')),
                        DropdownMenuItem(value: 'nearby', child: Text('附近')),
                        DropdownMenuItem(value: 'asia', child: Text('亚洲')),
                        DropdownMenuItem(value: 'europe', child: Text('欧洲')),
                        DropdownMenuItem(value: 'america', child: Text('美洲')),
                      ],
                      onChanged: (value) {
                        if (value != null) {
                          controller.selectedRegion.value = value;
                        }
                      },
                    )),
                const SizedBox(height: 10),
                Obx(() => DropdownButtonFormField<String>(
                      key: ValueKey(
                          'profile-gender-${controller.profileGender.value}'),
                      initialValue: controller.profileGender.value,
                      decoration: const InputDecoration(
                        labelText: '性别显示',
                        prefixIcon: Icon(Icons.badge_outlined),
                      ),
                      items: const [
                        DropdownMenuItem(value: 'private', child: Text('不公开')),
                        DropdownMenuItem(value: 'female', child: Text('女生')),
                        DropdownMenuItem(value: 'male', child: Text('男生')),
                      ],
                      onChanged: (value) {
                        if (value != null) {
                          controller.profileGender.value = value;
                        }
                      },
                    )),
                const SizedBox(height: 10),
                Obx(() => DropdownButtonFormField<String>(
                      key: ValueKey(
                          'profile-language-${controller.profileLanguage.value}'),
                      initialValue: controller.profileLanguage.value,
                      decoration: const InputDecoration(
                        labelText: '常用语言',
                        prefixIcon: Icon(Icons.translate_rounded),
                      ),
                      items: const [
                        DropdownMenuItem(value: 'zh', child: Text('中文')),
                        DropdownMenuItem(value: 'en', child: Text('English')),
                        DropdownMenuItem(value: 'ja', child: Text('日本語')),
                        DropdownMenuItem(value: 'ko', child: Text('한국어')),
                        DropdownMenuItem(value: 'es', child: Text('Español')),
                      ],
                      onChanged: (value) {
                        if (value != null) {
                          controller.profileLanguage.value = value;
                        }
                      },
                    )),
                const SizedBox(height: 8),
                Obx(() => AnimatedContainer(
                      key: _ageKey,
                      duration: const Duration(milliseconds: 220),
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                          color: controller.ageConfirmed.value
                              ? Colors.transparent
                              : const Color(0xffffc857).withValues(alpha: 0.45),
                        ),
                        color: controller.ageConfirmed.value
                            ? Colors.transparent
                            : const Color(0xffffc857).withValues(alpha: 0.08),
                      ),
                      child: Material(
                        color: Colors.transparent,
                        child: CheckboxListTile(
                          value: controller.ageConfirmed.value,
                          onChanged: (value) =>
                              controller.ageConfirmed.value = value ?? false,
                          contentPadding:
                              const EdgeInsets.symmetric(horizontal: 8),
                          title: const Text('我已满 18 岁并同意文明视讯'),
                        ),
                      ),
                    )),
                const SizedBox(height: 8),
                Obx(() {
                  final saving = controller.savingProfile.value;
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      FilledButton(
                        onPressed: saving ? null : controller.saveProfile,
                        child: Text(saving ? '保存中' : '保存资料'),
                      ),
                      if (controller.ageConfirmed.value) ...[
                        const SizedBox(height: 8),
                        OutlinedButton.icon(
                          onPressed: saving
                              ? null
                              : () async {
                                  await controller.saveProfile();
                                  controller.switchPage(AppPage.video);
                                },
                          icon: const Icon(Icons.videocam_outlined),
                          label: const Text('保存并返回视讯'),
                        ),
                      ],
                    ],
                  );
                }),
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
                          child: Material(
                            color: Colors.transparent,
                            child: ListTile(
                              leading: CircleAvatar(
                                backgroundColor: const Color(0xff2fd276),
                                foregroundColor: const Color(0xff06150c),
                                child: Text(initial,
                                    style: const TextStyle(
                                        fontWeight: FontWeight.w900)),
                              ),
                              title: Text(user.displayName),
                              subtitle: Text(
                                  user.bio.isEmpty ? '已拉黑用户' : user.bio),
                              trailing: TextButton(
                                onPressed: () =>
                                    controller.unblockUser(user.id),
                                child: const Text('解除'),
                              ),
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

class _InterestPicker extends GetView<MatchController> {
  const _InterestPicker();

  @override
  Widget build(BuildContext context) {
    return Obx(() {
      controller.interestsRevision.value;
      final selected = controller.parsedInterests();
      return Wrap(
        spacing: 8,
        runSpacing: 8,
        children: interestSuggestions.map((item) {
          final active = selected.contains(item);
          return FilterChip(
            selected: active,
            label: Text(item),
            onSelected: (_) => controller.toggleInterest(item),
            selectedColor: const Color(0xff20c8ff).withValues(alpha: 0.22),
            checkmarkColor: const Color(0xff20c8ff),
            side: BorderSide(
              color: active
                  ? const Color(0xff20c8ff).withValues(alpha: 0.55)
                  : Colors.white.withValues(alpha: 0.14),
            ),
            labelStyle: TextStyle(
              color:
                  active ? Colors.white : Colors.white.withValues(alpha: 0.78),
              fontWeight: FontWeight.w800,
            ),
          );
        }).toList(),
      );
    });
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
