import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';
import 'discover_page.dart';
import 'membership_page.dart';
import 'profile_page.dart';
import 'video_page.dart';

class HomePage extends GetView<MatchController> {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xff050608),
      resizeToAvoidBottomInset: true,
      body: Obx(() {
        return IndexedStack(
          index: controller.page.value.index,
          children: const [
            VideoPage(),
            DiscoverPage(),
            ProfilePage(),
            MembershipPage(),
          ],
        );
      }),
      bottomNavigationBar: Obx(() {
        return _AuroraBottomNav(
          selectedIndex: controller.page.value.index,
          onSelected: (index) => controller.switchPage(AppPage.values[index]),
        );
      }),
    );
  }
}

class _AuroraBottomNav extends StatelessWidget {
  const _AuroraBottomNav({
    required this.selectedIndex,
    required this.onSelected,
  });

  final int selectedIndex;
  final ValueChanged<int> onSelected;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: const BoxDecoration(
        color: Color(0xff050608),
        border: Border(top: BorderSide(color: Color(0x1fffffff))),
        boxShadow: [
          BoxShadow(color: Colors.black54, blurRadius: 28),
          BoxShadow(color: Color(0x227c5cff), blurRadius: 38),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
          child: Row(
            children: [
              Expanded(
                child: _AuroraNavItem(
                  selected: selectedIndex == 0,
                  icon: Icons.videocam_outlined,
                  label: '视讯',
                  onTap: () => onSelected(0),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: _AuroraNavItem(
                  selected: selectedIndex == 1,
                  icon: Icons.style_outlined,
                  label: '探索',
                  onTap: () => onSelected(1),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: _AuroraNavItem(
                  selected: selectedIndex == 2,
                  icon: Icons.person_outline_rounded,
                  label: '资料',
                  onTap: () => onSelected(2),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: _AuroraNavItem(
                  selected: selectedIndex == 3,
                  icon: Icons.workspace_premium_outlined,
                  label: '会员',
                  onTap: () => onSelected(3),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _AuroraNavItem extends StatelessWidget {
  const _AuroraNavItem({
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
    final foreground = selected ? const Color(0xfff5f0ff) : Colors.white60;

    return InkWell(
      borderRadius: BorderRadius.circular(8),
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 180),
        height: 62,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: selected
                ? const Color(0xffc4aaff).withValues(alpha: 0.42)
                : Colors.white.withValues(alpha: 0.1),
          ),
          gradient: selected
              ? LinearGradient(
                  colors: [
                    const Color(0xff7c5cff).withValues(alpha: 0.34),
                    const Color(0xff20c8ff).withValues(alpha: 0.18),
                  ],
                )
              : null,
          color: selected ? null : const Color(0xff161923),
          boxShadow: selected
              ? const [
                  BoxShadow(color: Color(0x337c5cff), blurRadius: 24),
                ]
              : null,
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, color: foreground, size: 22),
            const SizedBox(height: 3),
            Text(
              label,
              style: TextStyle(
                color: foreground,
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
