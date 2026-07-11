import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';

const _surface = Color(0xff111723);
const _surfaceSoft = Color(0xff18202f);
const _neon = Color(0xff00ffc3);
const _cyan = Color(0xff00d2ff);
const _violet = Color(0xff7c5cff);
const _pink = Color(0xffff4b91);

class MembershipPage extends GetView<MatchController> {
  const MembershipPage({super.key});

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Obx(() {
        final status = controller.commerceStatus.value;
        final isMember = status?.isMember == true;
        final paymentLoading = controller.paymentLoading.value;
        final canBuy = !paymentLoading && !isMember;
        final title = isMember ? '会员已开启' : '免费匹配额度';
        final subtitle = status == null
            ? '正在读取今日额度'
            : isMember
                ? '无限匹配 · 地区/对象筛选 · 优先排队'
                : '今日剩余 ${status.dailyRemaining}/${status.dailyLimit} 次 · 会员无限匹配并优先排队';
        final actionText = isMember
            ? '已是会员'
            : paymentLoading
                ? '开通中'
                : r'$6.99/月 开通';

        return ListView(
          padding: const EdgeInsets.fromLTRB(14, 16, 14, 20),
          children: [
            _PassHero(
              title: title,
              subtitle: subtitle,
              actionText: actionText,
              onPressed: canBuy ? controller.buyMembership : null,
            ),
            const SizedBox(height: 12),
            _AccountStatusCard(
              status: status,
              loading: controller.paymentLoading.value,
              onRefresh: controller.loadCommerceStatus,
            ),
            const SizedBox(height: 12),
            _PassCard(
              title: '优先通行',
              subtitle: '无限匹配 + 优先队列 + 300 Gems',
              price: isMember ? '已开通' : r'$6.99/月',
              featured: true,
              onPressed: canBuy
                  ? () => controller.buyMembership(plan: 'premium_monthly')
                  : null,
            ),
            const SizedBox(height: 10),
            _PassCard(
              title: '月度会员',
              subtitle: '30 天会员权益，可续期叠加',
              price: isMember ? '已拥有' : r'$6.99',
              onPressed: canBuy
                  ? () => controller.buyMembership(plan: 'premium_monthly')
                  : null,
            ),
            const SizedBox(height: 10),
            _PassCard(
              title: '畅聊模式',
              subtitle: '免费额度用完后仍可继续匹配',
              price: isMember ? '已包含' : '会员包含',
              gemColor: _pink,
              onPressed: canBuy
                  ? () => controller.buyMembership(plan: 'premium_monthly')
                  : null,
            ),
            const SizedBox(height: 12),
            const _BenefitGrid(),
            const SizedBox(height: 12),
            const _PaymentNote(),
          ],
        );
      }),
    );
  }
}

class _PassHero extends StatelessWidget {
  const _PassHero({
    required this.title,
    required this.subtitle,
    required this.actionText,
    required this.onPressed,
  });

  final String title;
  final String subtitle;
  final String actionText;
  final VoidCallback? onPressed;

  @override
  Widget build(BuildContext context) {
    return _GlassPanel(
      padding: const EdgeInsets.fromLTRB(18, 22, 18, 20),
      child: Column(
        children: [
          const _GemOrb(size: 68),
          const SizedBox(height: 18),
          const Text(
            'MATCH PASS',
            style: TextStyle(
              color: _neon,
              fontSize: 13,
              fontWeight: FontWeight.w900,
              letterSpacing: 5,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            title,
            textAlign: TextAlign.center,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 32,
              fontWeight: FontWeight.w900,
              height: 1.08,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            subtitle,
            textAlign: TextAlign.center,
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.68),
              fontSize: 16,
              fontWeight: FontWeight.w700,
              height: 1.42,
            ),
          ),
          const SizedBox(height: 22),
          SizedBox(
            width: 260,
            height: 48,
            child: FilledButton(
              style: FilledButton.styleFrom(
                backgroundColor: _neon,
                disabledBackgroundColor: _neon.withValues(alpha: 0.48),
                foregroundColor: const Color(0xff04110d),
                disabledForegroundColor:
                    const Color(0xff04110d).withValues(alpha: 0.62),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(999),
                ),
                textStyle: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w900,
                ),
              ),
              onPressed: onPressed,
              child: Text(actionText),
            ),
          ),
        ],
      ),
    );
  }
}

class _PassCard extends StatelessWidget {
  const _PassCard({
    required this.title,
    required this.subtitle,
    required this.price,
    required this.onPressed,
    this.featured = false,
    this.gemColor = _cyan,
  });

  final String title;
  final String subtitle;
  final String price;
  final VoidCallback? onPressed;
  final bool featured;
  final Color gemColor;

  @override
  Widget build(BuildContext context) {
    final borderColor = featured
        ? _neon.withValues(alpha: 0.48)
        : Colors.white.withValues(alpha: 0.1);
    return Material(
      color: Colors.transparent,
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onPressed,
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: featured
                ? _neon.withValues(alpha: 0.08)
                : Colors.white.withValues(alpha: 0.035),
            border: Border.all(color: borderColor),
            borderRadius: BorderRadius.circular(8),
            boxShadow: [
              if (featured)
                BoxShadow(
                  color: _neon.withValues(alpha: 0.14),
                  blurRadius: 28,
                  offset: const Offset(0, 10),
                ),
            ],
          ),
          child: Stack(
            children: [
              if (featured)
                const Positioned(
                  left: 0,
                  right: 0,
                  top: 0,
                  child: ColoredBox(
                    color: _neon,
                    child: SizedBox(height: 3),
                  ),
                ),
              if (featured)
                Positioned(
                  right: 10,
                  top: 9,
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      color: _neon,
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: const Padding(
                      padding: EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                      child: Text(
                        '推荐',
                        style: TextStyle(
                          color: Color(0xff06150c),
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                  ),
                ),
              Padding(
                padding: const EdgeInsets.all(14),
                child: Row(
                  children: [
                    _GemOrb(size: 48, color: gemColor),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            title,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 17,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            subtitle,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              color: Colors.white.withValues(alpha: 0.58),
                              fontSize: 13,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 10),
                    DecoratedBox(
                      decoration: BoxDecoration(
                        color: featured ? _neon : Colors.white,
                        borderRadius: BorderRadius.circular(999),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 13,
                          vertical: 10,
                        ),
                        child: Text(
                          price,
                          style: const TextStyle(
                            color: Color(0xff111318),
                            fontSize: 14,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ),
                    ),
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

class _AccountStatusCard extends StatelessWidget {
  const _AccountStatusCard({
    required this.status,
    required this.loading,
    required this.onRefresh,
  });

  final CommerceStatus? status;
  final bool loading;
  final Future<void> Function() onRefresh;

  @override
  Widget build(BuildContext context) {
    final isMember = status?.isMember == true;
    final dailyLimit = status?.dailyLimit ?? 10;
    final dailyUsed = status?.dailyUsed ?? 0;
    final dailyRemaining = status?.dailyRemaining ?? dailyLimit;
    final gemsBalance = status?.gemsBalance ?? 0;
    final expiresAt = status?.membershipExpiresAt;
    final progress = isMember || dailyLimit <= 0
        ? 1.0
        : (dailyUsed / dailyLimit).clamp(0.0, 1.0);
    final quotaText = isMember
        ? '无限匹配'
        : '今日剩余 $dailyRemaining/$dailyLimit 次';

    return _GlassPanel(
      padding: const EdgeInsets.all(14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Icon(
                isMember
                    ? Icons.verified_rounded
                    : Icons.local_activity_rounded,
                color: isMember ? _neon : _cyan,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  isMember ? '当前权益：Match Pass' : '当前权益：免费账户',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              TextButton(
                onPressed: loading ? null : onRefresh,
                child: Text(loading ? '同步中' : '同步状态'),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ClipRRect(
            borderRadius: BorderRadius.circular(999),
            child: LinearProgressIndicator(
              value: progress,
              minHeight: 8,
              backgroundColor: Colors.white.withValues(alpha: 0.08),
              valueColor: AlwaysStoppedAnimation<Color>(
                isMember ? _neon : _cyan,
              ),
            ),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              _StatusPill(label: '匹配额度', value: quotaText),
              _StatusPill(label: 'Gems', value: '$gemsBalance'),
              _StatusPill(
                label: '优先队列',
                value: isMember ? '已开启' : '未开启',
              ),
              if (expiresAt != null)
                _StatusPill(label: '到期日', value: _formatDate(expiresAt)),
            ],
          ),
        ],
      ),
    );
  }
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.06),
        border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              '$label ',
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.55),
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
            Text(
              value,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _BenefitGrid extends StatelessWidget {
  const _BenefitGrid();

  @override
  Widget build(BuildContext context) {
    return _GlassPanel(
      padding: EdgeInsets.zero,
      child: Column(
        children: const [
          _BenefitTile(
            title: '无限随机匹配',
            text: '不受每日免费次数限制',
          ),
          _DividerLine(),
          _BenefitTile(
            title: '进入优先队列',
            text: '高峰期减少等待时间',
          ),
          _DividerLine(),
          _BenefitTile(
            title: '安全加速体验',
            text: '保留举报、拉黑和离开控制',
          ),
        ],
      ),
    );
  }
}

String _formatDate(DateTime value) {
  final local = value.toLocal();
  return '${local.year}/${local.month.toString().padLeft(2, '0')}/${local.day.toString().padLeft(2, '0')}';
}

class _BenefitTile extends StatelessWidget {
  const _BenefitTile({
    required this.title,
    required this.text,
  });

  final String title;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(14),
      child: Row(
        children: [
          const Icon(Icons.check_circle_rounded, color: _neon, size: 20),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  text,
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: 0.58),
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _PaymentNote extends StatelessWidget {
  const _PaymentNote();

  @override
  Widget build(BuildContext context) {
    return _GlassPanel(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 13),
      child: Wrap(
        alignment: WrapAlignment.center,
        spacing: 14,
        runSpacing: 8,
        children: const [
          _NoteItem('安全交易'),
          _NoteItem('加密支付'),
          _NoteItem('可随时确认状态'),
        ],
      ),
    );
  }
}

class _NoteItem extends StatelessWidget {
  const _NoteItem(this.text);

  final String text;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Text(
          '•',
          style: TextStyle(color: _neon, fontWeight: FontWeight.w900),
        ),
        const SizedBox(width: 6),
        Text(
          text,
          style: TextStyle(
            color: Colors.white.withValues(alpha: 0.58),
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }
}

class _GemOrb extends StatelessWidget {
  const _GemOrb({
    required this.size,
    this.color = _cyan,
  });

  final double size;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors:
              color == _pink ? const [_pink, _violet] : const [_neon, _cyan],
        ),
        boxShadow: [
          BoxShadow(
            color: color.withValues(alpha: 0.28),
            blurRadius: 28,
            offset: const Offset(0, 12),
          ),
        ],
      ),
      child: SizedBox(
        width: size,
        height: size,
        child: Center(
          child: Transform.rotate(
            angle: 0.785398,
            child: Container(
              width: size * 0.34,
              height: size * 0.34,
              color: color == _pink ? Colors.white : const Color(0xff050608),
            ),
          ),
        ),
      ),
    );
  }
}

class _GlassPanel extends StatelessWidget {
  const _GlassPanel({
    required this.child,
    required this.padding,
  });

  final Widget child;
  final EdgeInsetsGeometry padding;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: _surface.withValues(alpha: 0.82),
        border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
        borderRadius: BorderRadius.circular(8),
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            _neon.withValues(alpha: 0.07),
            _cyan.withValues(alpha: 0.04),
            _surfaceSoft.withValues(alpha: 0.66),
          ],
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.34),
            blurRadius: 34,
            offset: const Offset(0, 18),
          ),
        ],
      ),
      child: Padding(
        padding: padding,
        child: child,
      ),
    );
  }
}

class _DividerLine extends StatelessWidget {
  const _DividerLine();

  @override
  Widget build(BuildContext context) {
    return Divider(
      height: 1,
      thickness: 1,
      color: Colors.white.withValues(alpha: 0.09),
    );
  }
}
