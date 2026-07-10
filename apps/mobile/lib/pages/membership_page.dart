import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';

class MembershipPage extends GetView<MatchController> {
  const MembershipPage({super.key});

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          DecoratedBox(
            decoration: BoxDecoration(
              color: const Color(0xcc111419),
              border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Obx(() {
                final status = controller.commerceStatus.value;
                final isMember = status?.isMember == true;
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            isMember ? 'Premium 已开启' : 'Gems 与筛选',
                            style: const TextStyle(
                                fontSize: 24, fontWeight: FontWeight.w900),
                          ),
                        ),
                        DecoratedBox(
                          decoration: BoxDecoration(
                            color:
                                const Color(0xff20c8ff).withValues(alpha: 0.18),
                            borderRadius: BorderRadius.circular(999),
                          ),
                          child: Padding(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 10, vertical: 7),
                            child: Text(
                              '${status?.gemsBalance ?? 0} Gems',
                              style:
                                  const TextStyle(fontWeight: FontWeight.w900),
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Text(
                      status == null
                          ? '正在读取今日额度'
                          : isMember
                              ? '无限匹配 · 地区/对象筛选 · 优先排队'
                              : '今日剩余 ${status.dailyRemaining}/${status.dailyLimit} 次 · Gems 可用于筛选与刷新',
                      style: const TextStyle(color: Colors.white70),
                    ),
                    const SizedBox(height: 18),
                    const _Benefit(
                      icon: Icons.public_rounded,
                      title: '地区筛选',
                      text: '全球、附近、亚洲、欧洲、美洲',
                    ),
                    const _Benefit(
                      icon: Icons.tune_rounded,
                      title: '对象筛选',
                      text: '用更明确的偏好进入匹配队列',
                    ),
                    const _Benefit(
                      icon: Icons.bolt_rounded,
                      title: '优先队列',
                      text: '高峰时段优先消耗会员等待队列',
                    ),
                    const _Benefit(
                      icon: Icons.all_inclusive_rounded,
                      title: '无限随机匹配',
                      text: '免费额度用完后继续连接新朋友',
                    ),
                    const SizedBox(height: 18),
                    FilledButton(
                      onPressed: controller.paymentLoading.value || isMember
                          ? null
                          : controller.buyMembership,
                      child: Text(isMember
                          ? '已是会员'
                          : controller.paymentLoading.value
                              ? '开通中'
                              : r'$6.99/月 开通'),
                    ),
                  ],
                );
              }),
            ),
          ),
        ],
      ),
    );
  }
}

class _Benefit extends StatelessWidget {
  const _Benefit({
    required this.icon,
    required this.title,
    required this.text,
  });

  final IconData icon;
  final String title;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Icon(icon, color: const Color(0xff20c8ff)),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(title,
                        style: const TextStyle(fontWeight: FontWeight.w900)),
                    const SizedBox(height: 2),
                    Text(text, style: const TextStyle(color: Colors.white70)),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
