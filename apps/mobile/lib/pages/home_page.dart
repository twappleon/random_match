import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';
import 'membership_page.dart';
import 'profile_page.dart';
import 'video_page.dart';

class HomePage extends GetView<MatchController> {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Obx(() {
        return IndexedStack(
          index: controller.page.value.index,
          children: const [
            VideoPage(),
            ProfilePage(),
            MembershipPage(),
          ],
        );
      }),
      bottomNavigationBar: Obx(() {
        return NavigationBar(
          selectedIndex: controller.page.value.index,
          onDestinationSelected: (index) =>
              controller.switchPage(AppPage.values[index]),
          destinations: const [
            NavigationDestination(
                icon: Icon(Icons.videocam_outlined), label: '视讯'),
            NavigationDestination(
                icon: Icon(Icons.person_outline), label: '资料'),
            NavigationDestination(
                icon: Icon(Icons.workspace_premium_outlined), label: '会员'),
          ],
        );
      }),
    );
  }
}
