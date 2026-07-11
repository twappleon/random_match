import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';
import '../widgets/chat_sheet.dart';
import '../widgets/peer_bar.dart';

class VideoPage extends GetView<MatchController> {
  const VideoPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Obx(() {
      final status = controller.status.value;
      final matched = status == MatchStatus.matched;
      final waiting = status == MatchStatus.waiting;
      final chatOpen = controller.chatOpen.value;
      final controlsVisible = !matched ||
          controller.controlsVisible.value ||
          controller.controlsExpanded.value;
      final keyboardInset = MediaQuery.viewInsetsOf(context).bottom;
      final keyboardOpen = keyboardInset > 0;
      final chatBottom = keyboardOpen ? 12.0 : 100.0;
      final remoteVideoTick = controller.remoteVideoTick.value;

      return LayoutBuilder(
        builder: (context, constraints) {
          final chatHeight = (keyboardOpen
                  ? constraints.maxHeight - 24
                  : constraints.maxHeight * 0.48)
              .clamp(300.0, 430.0)
              .toDouble();

          return Stack(
            children: [
              Positioned.fill(
                child: _AuroraVideoBackground(
                  child: matched
                      ? RTCVideoView(
                          key: ValueKey(remoteVideoTick),
                          controller.remoteRenderer,
                          objectFit:
                              RTCVideoViewObjectFit.RTCVideoViewObjectFitCover,
                        )
                      : const SizedBox.expand(),
                ),
              ),
              const Positioned.fill(child: _AuroraScrim()),
              Positioned(
                top: 14,
                left: 14,
                right: 14,
                child: SafeArea(
                  bottom: false,
                  child: _StatsBar(stats: controller.stats.value),
                ),
              ),
              if (matched && !chatOpen)
                Positioned(
                  top: 86,
                  left: 14,
                  right: 14,
                  child: SafeArea(
                    bottom: false,
                    child: _TranslationCaption(controller: controller),
                  ),
                ),
              if (!matched)
                Positioned.fill(
                  top: 96,
                  bottom: 212,
                  child: Center(
                    child: SingleChildScrollView(
                      padding: const EdgeInsets.symmetric(horizontal: 20),
                      child: _AuroraStateCard(
                        controller: controller,
                        waiting: waiting,
                        title: waiting ? '正在寻找新朋友' : '今晚遇见新朋友',
                        subtitle: waiting
                            ? '保持页面开启，匹配成功后会自动进入视讯。'
                            : '用视讯、语音、地区和对象偏好快速进入随机连线。',
                      ),
                    ),
                  ),
                ),
              if (matched && !chatOpen)
                Positioned(
                  right: 16,
                  bottom: controlsVisible ? 116 : 82,
                  width: 94,
                  height: 138,
                  child: _LocalPreview(renderer: controller.localRenderer),
                ),
              if (matched && !chatOpen && !controller.peerCardHidden.value)
                Positioned(
                  left: 12,
                  right: 118,
                  bottom: controlsVisible ? 118 : 28,
                  child: const PeerBar(),
                ),
              if (chatOpen)
                Positioned(
                  left: 12,
                  right: 12,
                  bottom: chatBottom,
                  height: chatHeight,
                  child: const ChatSheet(),
                ),
              if (!chatOpen && controlsVisible)
                Positioned(
                  left: 12,
                  right: 12,
                  bottom: 18,
                  child: SafeArea(
                    top: false,
                    child: _ActionDock(
                      controller: controller,
                      status: status,
                    ),
                  ),
                ),
              if (matched && !chatOpen && !controlsVisible)
                Positioned(
                  right: 12,
                  bottom: 18,
                  child: SafeArea(
                    top: false,
                    child: _ControlHandle(
                      onPressed: () {
                        controller.controlsVisible.value = true;
                        controller.controlsExpanded.value = true;
                      },
                    ),
                  ),
                ),
            ],
          );
        },
      );
    });
  }
}

class _AuroraVideoBackground extends StatelessWidget {
  const _AuroraVideoBackground({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: const BoxDecoration(
        gradient: RadialGradient(
          center: Alignment(-0.72, -0.72),
          radius: 1.08,
          colors: [Color(0x677c5cff), Color(0x00000000)],
        ),
      ),
      child: DecoratedBox(
        decoration: const BoxDecoration(
          gradient: RadialGradient(
            center: Alignment(0.8, -0.5),
            radius: 0.92,
            colors: [Color(0x5520c8ff), Color(0x00000000)],
          ),
        ),
        child: DecoratedBox(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                Color(0xff151735),
                Color(0xff080a16),
                Color(0xff050608),
              ],
            ),
          ),
          child: child,
        ),
      ),
    );
  }
}

class _AuroraScrim extends StatelessWidget {
  const _AuroraScrim();

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: DecoratedBox(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [
              Colors.black.withValues(alpha: 0.02),
              Colors.black.withValues(alpha: 0.16),
              Colors.black.withValues(alpha: 0.5),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatsBar extends StatelessWidget {
  const _StatsBar({required this.stats});

  final RuntimeStats stats;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(child: _AuroraStat(value: stats.online, label: '在线')),
        const SizedBox(width: 8),
        Expanded(child: _AuroraStat(value: stats.waiting, label: '等待')),
        const SizedBox(width: 8),
        Expanded(child: _AuroraStat(value: stats.chatting, label: '聊天')),
      ],
    );
  }
}

class _LocalPreview extends StatelessWidget {
  const _LocalPreview({required this.renderer});

  final RTCVideoRenderer renderer;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: DecoratedBox(
        decoration: BoxDecoration(
          border: Border.all(color: const Color(0x66c4aaff)),
          boxShadow: const [
            BoxShadow(color: Color(0x4d7c5cff), blurRadius: 42),
            BoxShadow(color: Colors.black45, blurRadius: 30),
          ],
        ),
        child: ColoredBox(
          color: const Color(0xff202632),
          child: RTCVideoView(renderer, mirror: true),
        ),
      ),
    );
  }
}

class _ControlHandle extends StatelessWidget {
  const _ControlHandle({required this.onPressed});

  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(999),
        boxShadow: const [
          BoxShadow(color: Colors.black54, blurRadius: 22),
          BoxShadow(color: Color(0x337c5cff), blurRadius: 26),
        ],
      ),
      child: FilledButton.icon(
        style: FilledButton.styleFrom(
          minimumSize: const Size(96, 44),
          padding: const EdgeInsets.symmetric(horizontal: 14),
          backgroundColor: const Color(0xd9080a13),
          foregroundColor: Colors.white,
          shape: RoundedRectangleBorder(
            side: BorderSide(color: Colors.white.withValues(alpha: 0.14)),
            borderRadius: BorderRadius.circular(999),
          ),
        ),
        onPressed: onPressed,
        icon: const Icon(Icons.tune_rounded, size: 18),
        label: const Text(
          '控制',
          style: TextStyle(fontWeight: FontWeight.w900),
        ),
      ),
    );
  }
}

class _ActionDock extends StatelessWidget {
  const _ActionDock({
    required this.controller,
    required this.status,
  });

  final MatchController controller;
  final MatchStatus status;

  @override
  Widget build(BuildContext context) {
    final loading = controller.loading.value;
    final leaving = controller.leaving.value;
    final matched = status == MatchStatus.matched;
    final waiting = status == MatchStatus.waiting;
    final expanded = controller.controlsExpanded.value;
    final canBlock = matched &&
        controller.activePeerId != null &&
        !controller.safetyLoading.value &&
        !leaving;
    final canStart = !loading && !waiting && !matched;
    final canLeave = !leaving && status != MatchStatus.idle;

    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
        borderRadius: BorderRadius.circular(10),
        color: const Color(0xd6080a13),
        boxShadow: const [
          BoxShadow(color: Colors.black54, blurRadius: 34),
          BoxShadow(color: Color(0x267c5cff), blurRadius: 44),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (expanded) ...[
              _CompactSettings(controller: controller, matched: matched),
              const SizedBox(height: 8),
            ],
            Row(
              children: [
                _DockIconButton(
                  selected: expanded,
                  icon: expanded
                      ? Icons.keyboard_arrow_down_rounded
                      : Icons.tune_rounded,
                  label: expanded ? '收起' : '设置',
                  onPressed: () {
                    if (expanded && matched) {
                      controller.controlsExpanded.value = false;
                      controller.controlsVisible.value = false;
                      return;
                    }
                    controller.controlsExpanded.value = !expanded;
                  },
                ),
                const SizedBox(width: 7),
                Expanded(
                  child: _DockMainButton(
                    onPressed: controller.toggleChat,
                    icon: controller.chatOpen.value
                        ? Icons.keyboard_arrow_down_rounded
                        : Icons.chat_bubble_outline_rounded,
                    label: controller.chatOpen.value ? '收起文字' : '文字',
                  ),
                ),
                const SizedBox(width: 7),
                Expanded(
                  child: _DockMainButton(
                    onPressed: canBlock ? controller.blockPeer : null,
                    icon: Icons.block_rounded,
                    label: controller.safetyLoading.value ? '处理中' : '拉黑',
                    dangerSoft: true,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(
                  child: _DockMainButton(
                    onPressed: canStart ? controller.startMatch : null,
                    icon: matched
                        ? Icons.check_rounded
                        : waiting
                            ? Icons.hourglass_top_rounded
                            : Icons.auto_awesome_rounded,
                    label: matched
                        ? '已连线'
                        : loading || waiting
                            ? '匹配中'
                            : '随机匹配',
                    primary: true,
                  ),
                ),
                const SizedBox(width: 7),
                Expanded(
                  child: _DockMainButton(
                    onPressed: canLeave ? controller.leaveCall : null,
                    icon: Icons.logout_rounded,
                    label: leaving ? '退出中' : '退出',
                    danger: true,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _CompactSettings extends StatelessWidget {
  const _CompactSettings({
    required this.controller,
    required this.matched,
  });

  final MatchController controller;
  final bool matched;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.055),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.white.withValues(alpha: 0.08)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            _PreferencePill(
              icon: matched ? Icons.lock_open_rounded : Icons.public_rounded,
              label:
                  '${controller.matchMode.value == MatchModePreference.voice ? '语音' : '视讯'} · ${controller.regionLabel(controller.selectedRegion.value)} · ${controller.genderPreferenceLabel}',
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(child: _TranslationSelector(controller: controller)),
                const SizedBox(width: 8),
                _DockIconButton(
                  icon: Icons.style_outlined,
                  label: '探索',
                  onPressed: () => controller.switchPage(AppPage.discover),
                ),
                const SizedBox(width: 8),
                _DockIconButton(
                  icon: Icons.workspace_premium_outlined,
                  label: '会员',
                  onPressed: () => controller.switchPage(AppPage.membership),
                ),
              ],
            ),
            if (!matched && controller.status.value != MatchStatus.waiting) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                    child: _ModeButton(
                      selected: controller.matchMode.value ==
                          MatchModePreference.video,
                      icon: Icons.videocam_outlined,
                      label: '视讯',
                      onTap: () => controller.matchMode.value =
                          MatchModePreference.video,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: _ModeButton(
                      selected: controller.matchMode.value ==
                          MatchModePreference.voice,
                      icon: Icons.mic_none_rounded,
                      label: '语音',
                      onTap: () => controller.matchMode.value =
                          MatchModePreference.voice,
                    ),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _TranslationCaption extends StatelessWidget {
  const _TranslationCaption({required this.controller});

  final MatchController controller;

  @override
  Widget build(BuildContext context) {
    final enabled =
        controller.translationLanguage.value != TranslationLanguage.off;
    return AnimatedOpacity(
      opacity: enabled ? 1 : 0,
      duration: const Duration(milliseconds: 180),
      child: IgnorePointer(
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: const Color(0xcc050608),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
          ),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            child: Row(
              children: [
                const Icon(Icons.translate_rounded,
                    size: 18, color: Color(0xff20c8ff)),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    '实时字幕 · ${controller.translationLanguageLabel}',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontWeight: FontWeight.w900),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _DockIconButton extends StatelessWidget {
  const _DockIconButton({
    required this.icon,
    required this.label,
    required this.onPressed,
    this.selected = false,
  });

  final IconData icon;
  final String label;
  final VoidCallback? onPressed;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 58,
      height: 52,
      child: FilledButton(
        style: FilledButton.styleFrom(
          padding: EdgeInsets.zero,
          backgroundColor:
              selected ? const Color(0xff25305a) : const Color(0xff1c2130),
          disabledBackgroundColor: const Color(0xff171b27),
          foregroundColor: Colors.white,
          disabledForegroundColor: Colors.white54,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
        onPressed: onPressed,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 19),
            const SizedBox(height: 3),
            Text(
              label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w900),
            ),
          ],
        ),
      ),
    );
  }
}

class _DockMainButton extends StatelessWidget {
  const _DockMainButton({
    required this.label,
    this.onPressed,
    this.icon,
    this.primary = false,
    this.danger = false,
    this.dangerSoft = false,
  });

  final String label;
  final VoidCallback? onPressed;
  final IconData? icon;
  final bool primary;
  final bool danger;
  final bool dangerSoft;

  @override
  Widget build(BuildContext context) {
    final background = primary
        ? null
        : danger
            ? const Color(0xffb84d5b)
            : dangerSoft
                ? const Color(0x3dff5c74)
                : const Color(0xff1c2130);
    return SizedBox(
      height: 52,
      child: DecoratedBox(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(8),
          gradient: primary
              ? const LinearGradient(
                  colors: [Color(0xff7c5cff), Color(0xff20c8ff)],
                )
              : null,
          boxShadow: primary
              ? const [BoxShadow(color: Color(0x477c5cff), blurRadius: 22)]
              : null,
        ),
        child: FilledButton(
          style: FilledButton.styleFrom(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            backgroundColor: primary ? Colors.transparent : background,
            disabledBackgroundColor: const Color(0xff171b27),
            shadowColor: Colors.transparent,
            foregroundColor: Colors.white,
            disabledForegroundColor: Colors.white54,
            shape:
                RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          ),
          onPressed: onPressed,
          child: FittedBox(
            fit: BoxFit.scaleDown,
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (icon != null) ...[
                  Icon(icon, size: 18),
                  const SizedBox(width: 5),
                ],
                Text(
                  label,
                  style: const TextStyle(fontWeight: FontWeight.w900),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _TranslationSelector extends StatelessWidget {
  const _TranslationSelector({required this.controller});

  final MatchController controller;

  @override
  Widget build(BuildContext context) {
    return DropdownButtonFormField<TranslationLanguage>(
      key: ValueKey('translation-${controller.translationLanguage.value.name}'),
      initialValue: controller.translationLanguage.value,
      decoration: const InputDecoration(
        labelText: '实时翻译',
        prefixIcon: Icon(Icons.translate_rounded),
        isDense: true,
      ),
      items: const [
        DropdownMenuItem(value: TranslationLanguage.off, child: Text('关闭')),
        DropdownMenuItem(value: TranslationLanguage.zh, child: Text('中文')),
        DropdownMenuItem(value: TranslationLanguage.en, child: Text('English')),
        DropdownMenuItem(value: TranslationLanguage.ja, child: Text('日本語')),
        DropdownMenuItem(value: TranslationLanguage.ko, child: Text('한국어')),
        DropdownMenuItem(value: TranslationLanguage.es, child: Text('Español')),
      ],
      onChanged: (value) {
        if (value != null) controller.translationLanguage.value = value;
      },
    );
  }
}

class _ModeButton extends StatelessWidget {
  const _ModeButton({
    required this.selected,
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final bool selected;
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 42,
      child: FilledButton.icon(
        style: FilledButton.styleFrom(
          backgroundColor:
              selected ? const Color(0xff25305a) : const Color(0xff1c2130),
          foregroundColor: Colors.white,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
        onPressed: onTap,
        icon: Icon(icon, size: 18),
        label: Text(label, style: const TextStyle(fontWeight: FontWeight.w900)),
      ),
    );
  }
}

class _PreferencePill extends StatelessWidget {
  const _PreferencePill({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 10),
        child: Row(
          children: [
            Icon(icon, size: 18, color: Colors.white70),
            const SizedBox(width: 7),
            Expanded(
              child: Text(
                label,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(fontWeight: FontWeight.w900),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AuroraStat extends StatelessWidget {
  const _AuroraStat({required this.value, required this.label});

  final int value;
  final String label;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
        borderRadius: BorderRadius.circular(8),
        color: Colors.white.withValues(alpha: 0.09),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              '$value',
              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w900),
            ),
            Text(
              label,
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.72),
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AuroraStateCard extends StatelessWidget {
  const _AuroraStateCard({
    required this.controller,
    required this.waiting,
    required this.title,
    required this.subtitle,
  });

  final MatchController controller;
  final bool waiting;
  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.white.withValues(alpha: 0.15)),
        borderRadius: BorderRadius.circular(8),
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            const Color(0xff7c5cff).withValues(alpha: 0.2),
            const Color(0xff20c8ff).withValues(alpha: 0.1),
            Colors.white.withValues(alpha: 0.05),
          ],
        ),
        color: const Color(0xbb141827),
        boxShadow: const [
          BoxShadow(color: Colors.black54, blurRadius: 52),
          BoxShadow(color: Color(0x267c5cff), blurRadius: 70),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            DecoratedBox(
              decoration: BoxDecoration(
                border: Border.all(
                    color: (waiting
                            ? const Color(0xff20c8ff)
                            : const Color(0xffc4aaff))
                        .withValues(alpha: 0.42)),
                borderRadius: BorderRadius.circular(999),
                color: (waiting
                        ? const Color(0xff20c8ff)
                        : const Color(0xff7c5cff))
                    .withValues(alpha: 0.18),
              ),
              child: Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 11, vertical: 6),
                child: Text(
                  waiting ? 'LIVE MATCH' : 'RANDOM LIVE',
                  style: const TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                    letterSpacing: 1.4,
                  ),
                ),
              ),
            ),
            const SizedBox(height: 14),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 30, fontWeight: FontWeight.w900),
            ),
            const SizedBox(height: 10),
            Text(
              subtitle,
              textAlign: TextAlign.center,
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.7),
                fontSize: 14,
                fontWeight: FontWeight.w700,
                height: 1.45,
              ),
            ),
            const SizedBox(height: 14),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              alignment: WrapAlignment.center,
              children: [
                _AuroraChip(
                    controller.matchMode.value == MatchModePreference.voice
                        ? '语音模式'
                        : '视讯模式'),
                _AuroraChip(
                    controller.regionLabel(controller.selectedRegion.value)),
                _AuroraChip(controller.genderPreferenceLabel),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: _HeroShortcut(
                    icon: Icons.style_outlined,
                    label: '探索',
                    onTap: () => controller.switchPage(AppPage.discover),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _HeroShortcut(
                    icon: Icons.workspace_premium_outlined,
                    label:
                        '${controller.commerceStatus.value?.gemsBalance ?? 0} Gems',
                    onTap: () => controller.switchPage(AppPage.membership),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _HeroShortcut(
                    icon: Icons.shield_outlined,
                    label: '安全',
                    onTap: () => controller.switchPage(AppPage.profile),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _HeroShortcut extends StatelessWidget {
  const _HeroShortcut({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(8),
      onTap: onTap,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 11, horizontal: 8),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(icon, size: 20, color: Colors.white70),
              const SizedBox(height: 5),
              Text(
                label,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _AuroraChip extends StatelessWidget {
  const _AuroraChip(this.label);

  final String label;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(999),
        color: Colors.white.withValues(alpha: 0.11),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
        child: Text(
          label,
          style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w800),
        ),
      ),
    );
  }
}
