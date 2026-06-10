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
                    Text(
                      isMember ? '会员已开启' : '免费匹配额度',
                      style: const TextStyle(
                          fontSize: 24, fontWeight: FontWeight.w900),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      status == null
                          ? '正在读取今日额度'
                          : isMember
                              ? '无限匹配 · 优先排队'
                              : '今日剩余 ${status.dailyRemaining}/${status.dailyLimit} 次 · 会员无限匹配并优先排队',
                      style: const TextStyle(color: Colors.white70),
                    ),
                    const SizedBox(height: 18),
                    const _Benefit(text: '无限随机匹配'),
                    const _Benefit(text: '进入优先队列'),
                    const _Benefit(text: '免费额度用完后继续使用'),
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
  const _Benefit({required this.text});

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
          child:
              Text(text, style: const TextStyle(fontWeight: FontWeight.w800)),
        ),
      ),
    );
  }
}
