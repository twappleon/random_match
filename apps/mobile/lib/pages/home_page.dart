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
      backgroundColor: const Color(0xff070b16),
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
        color: Color(0xff0a1020),
        border: Border(top: BorderSide(color: Color(0xff1b2435))),
      ),
      child: NavigationBar(
        selectedIndex: selectedIndex,
        onDestinationSelected: onSelected,
        destinations: [
          NavigationDestination(
            icon: const Icon(Icons.videocam_outlined),
            selectedIcon: const Icon(Icons.videocam_rounded),
            label: '连线'.tr,
          ),
          NavigationDestination(
            icon: const Icon(Icons.explore_outlined),
            selectedIcon: const Icon(Icons.explore_rounded),
            label: '探索'.tr,
          ),
          NavigationDestination(
            icon: const Icon(Icons.person_outline_rounded),
            selectedIcon: const Icon(Icons.person_rounded),
            label: '资料'.tr,
          ),
          NavigationDestination(
            icon: const Icon(Icons.workspace_premium_outlined),
            selectedIcon: const Icon(Icons.workspace_premium_rounded),
            label: '会员'.tr,
          ),
        ],
      ),
    );
  }
}
