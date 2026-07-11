import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controllers/match_controller.dart';
import '../data/models.dart';

class DiscoverPage extends GetView<MatchController> {
  const DiscoverPage({super.key});

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: RefreshIndicator(
        onRefresh: controller.loadDiscoverProfiles,
        child: Obx(() {
          final users = controller.discoverProfiles;
          final followedUsers = controller.followedUsers;
          return ListView(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 118),
            children: [
              Row(
                children: [
                  const Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Lounge',
                          style: TextStyle(
                              fontSize: 28, fontWeight: FontWeight.w900),
                        ),
                        SizedBox(height: 3),
                        Text(
                          '浏览资料、关注、私信，再进入 1v1 连线',
                          style: TextStyle(
                            color: Colors.white70,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                      ],
                    ),
                  ),
                  IconButton.filledTonal(
                    tooltip: '刷新',
                    onPressed: controller.discoverLoading.value
                        ? null
                        : controller.loadDiscoverProfiles,
                    icon: controller.discoverLoading.value
                        ? const SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.refresh_rounded),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              const _DiscoverFilters(),
              if (followedUsers.isNotEmpty) ...[
                const SizedBox(height: 16),
                const _FollowedSection(),
              ],
              const SizedBox(height: 16),
              if (users.isEmpty)
                _EmptyDiscover(loading: controller.discoverLoading.value)
              else
                ...users.map((user) => _ProfileCard(user: user)),
            ],
          );
        }),
      ),
    );
  }
}

class _FollowedSection extends GetView<MatchController> {
  const _FollowedSection();

  @override
  Widget build(BuildContext context) {
    return Obx(() {
      final users = controller.followedUsers;
      if (users.isEmpty) return const SizedBox.shrink();
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Expanded(
                child: Text(
                  '我的关注',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.w900),
                ),
              ),
              TextButton(
                onPressed: controller.loadDiscoverProfiles,
                child: const Text('同步'),
              ),
            ],
          ),
          const SizedBox(height: 8),
          SizedBox(
            height: 96,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: users.length,
              separatorBuilder: (_, __) => const SizedBox(width: 10),
              itemBuilder: (context, index) {
                final user = users[index];
                final name = _discoverDisplayName(user);
                final initial = name.characters.first.toUpperCase();
                return InkWell(
                  borderRadius: BorderRadius.circular(8),
                  onTap: () => controller.startFromProfile(user),
                  child: Container(
                    width: 132,
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: const Color(0xcc111419),
                      border: Border.all(
                          color: Colors.white.withValues(alpha: 0.12)),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        CircleAvatar(
                          radius: 18,
                          backgroundColor: const Color(0xff20c8ff),
                          foregroundColor: const Color(0xff041018),
                          child: Text(initial,
                              style:
                                  const TextStyle(fontWeight: FontWeight.w900)),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          name,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(fontWeight: FontWeight.w900),
                        ),
                        Text(
                          user.interests.take(2).join(' · '),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                              color: Colors.white70, fontSize: 12),
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
        ],
      );
    });
  }
}

class _DiscoverFilters extends GetView<MatchController> {
  const _DiscoverFilters();

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: const Color(0xcc111419),
        border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Expanded(
                  child: Obx(() => SegmentedButton<MatchModePreference>(
                        segments: const [
                          ButtonSegment(
                            value: MatchModePreference.video,
                            icon: Icon(Icons.videocam_outlined),
                            label: Text('视讯'),
                          ),
                          ButtonSegment(
                            value: MatchModePreference.voice,
                            icon: Icon(Icons.mic_none_rounded),
                            label: Text('语音'),
                          ),
                        ],
                        selected: {controller.matchMode.value},
                        onSelectionChanged: (value) =>
                            controller.matchMode.value = value.first,
                      )),
                ),
              ],
            ),
            const SizedBox(height: 10),
            Row(
              children: [
                Expanded(
                  child: Obx(() => DropdownButtonFormField<String>(
                        key: ValueKey(
                            'discover-region-${controller.selectedRegion.value}'),
                        initialValue: controller.selectedRegion.value,
                        decoration: const InputDecoration(
                          labelText: '地区',
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
                          if (value == null) return;
                          controller.selectedRegion.value = value;
                          controller.loadDiscoverProfiles();
                        },
                      )),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Obx(() => DropdownButtonFormField<GenderPreference>(
                        key: ValueKey(
                            'discover-gender-${controller.genderPreference.value.name}'),
                        initialValue: controller.genderPreference.value,
                        decoration: const InputDecoration(
                          labelText: '对象',
                          prefixIcon: Icon(Icons.tune_rounded),
                        ),
                        items: const [
                          DropdownMenuItem(
                            value: GenderPreference.everyone,
                            child: Text('不限'),
                          ),
                          DropdownMenuItem(
                            value: GenderPreference.female,
                            child: Text('女生'),
                          ),
                          DropdownMenuItem(
                            value: GenderPreference.male,
                            child: Text('男生'),
                          ),
                        ],
                        onChanged: (value) {
                          if (value == null) return;
                          controller.genderPreference.value = value;
                          controller.loadDiscoverProfiles();
                        },
                      )),
                ),
              ],
            ),
            const SizedBox(height: 10),
            Obx(() => _PremiumFilterNotice(
                  active: controller.premiumFiltersActive,
                  hasAccess: controller.hasPremiumAccess,
                  gems: controller.commerceStatus.value?.gemsBalance ?? 0,
                  onUpgrade: () => controller.switchPage(AppPage.membership),
                )),
          ],
        ),
      ),
    );
  }
}

class _EmptyDiscover extends StatelessWidget {
  const _EmptyDiscover({required this.loading});

  final bool loading;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 360,
      child: Center(
        child: DecoratedBox(
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.08),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
          ),
          child: Padding(
            padding: const EdgeInsets.all(18),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  loading ? Icons.sync_rounded : Icons.style_outlined,
                  size: 38,
                  color: Colors.white70,
                ),
                const SizedBox(height: 12),
                Text(
                  loading ? '正在整理资料卡' : '还没有可探索的资料',
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 6),
                const Text(
                  '保存资料或稍后刷新，再开始随机连接。',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: Colors.white70),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _ProfileCard extends GetView<MatchController> {
  const _ProfileCard({required this.user});

  final UserProfile user;

  @override
  Widget build(BuildContext context) {
    final followed = controller.followedUserIds.contains(user.id);
    final displayName = _discoverDisplayName(user);
    final initial = displayName.trim().isNotEmpty
        ? displayName.trim().characters.first.toUpperCase()
        : '星';
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: const Color(0xcc111419),
          border: Border.all(color: Colors.white.withValues(alpha: 0.12)),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  CircleAvatar(
                    radius: 28,
                    backgroundColor: const Color(0xff20c8ff),
                    foregroundColor: const Color(0xff041018),
                    child: Text(
                      initial,
                      style: const TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Flexible(
                              child: Text(
                                displayName,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.w900,
                                ),
                              ),
                            ),
                            if (user.trustBadge) ...[
                              const SizedBox(width: 6),
                              const Icon(Icons.verified_rounded,
                                  size: 18, color: Color(0xff2fd276)),
                            ],
                          ],
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '${controller.regionLabel(user.region)} · ${user.bio.isEmpty ? '愿意认识新朋友' : user.bio}',
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: Colors.white70,
                            height: 1.35,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: (user.interests.isEmpty
                        ? const ['聊天', '音乐', '电影']
                        : user.interests)
                    .take(5)
                    .map((item) => _InterestChip(item))
                    .toList(),
              ),
              const SizedBox(height: 14),
              Row(
                children: [
                  Expanded(
                    child: _MiniActionButton(
                      icon: followed
                          ? Icons.favorite_rounded
                          : Icons.favorite_border_rounded,
                      label: followed ? '已关注' : '关注',
                      selected: followed,
                      onTap: () => controller.toggleFollow(user),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: _MiniActionButton(
                      icon: Icons.mail_outline_rounded,
                      label: '私信',
                      onTap: () => controller.openDirectMessage(user),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: () => controller.dismissDiscoverProfile(user),
                      icon: const Icon(Icons.close_rounded),
                      label: const Text('换一批'),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: FilledButton.icon(
                      onPressed: () => controller.startFromProfile(user),
                      icon: const Icon(Icons.videocam_outlined),
                      label: const Text('以此偏好连接'),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

String _discoverDisplayName(UserProfile user) {
  final name = user.displayName.trim();
  if (name.isNotEmpty && name != '星球旅人') return name;
  final id = user.id.trim().toUpperCase();
  final suffix = id.length > 4 ? id.substring(0, 4) : id;
  return suffix.isEmpty ? '星球旅人' : '星球旅人 $suffix';
}

class _PremiumFilterNotice extends StatelessWidget {
  const _PremiumFilterNotice({
    required this.active,
    required this.hasAccess,
    required this.gems,
    required this.onUpgrade,
  });

  final bool active;
  final bool hasAccess;
  final int gems;
  final VoidCallback onUpgrade;

  @override
  Widget build(BuildContext context) {
    final color = hasAccess ? const Color(0xff20c8ff) : const Color(0xffffc857);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color.withValues(alpha: 0.28)),
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 10),
        child: Row(
          children: [
            Icon(
              active ? Icons.diamond_outlined : Icons.lock_open_rounded,
              size: 18,
              color: color,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                active
                    ? hasAccess
                        ? '已启用付费筛选 · $gems Gems 可用'
                        : '地区/对象筛选需要 Gems 或 Premium'
                    : '基础随机免费，精准筛选可用 Gems 解锁',
                style: const TextStyle(fontWeight: FontWeight.w800),
              ),
            ),
            if (!hasAccess)
              TextButton(onPressed: onUpgrade, child: const Text('开通')),
          ],
        ),
      ),
    );
  }
}

class _MiniActionButton extends StatelessWidget {
  const _MiniActionButton({
    required this.icon,
    required this.label,
    required this.onTap,
    this.selected = false,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 42,
      child: OutlinedButton.icon(
        style: OutlinedButton.styleFrom(
          foregroundColor: selected ? const Color(0xff2fd276) : Colors.white,
          side: BorderSide(
            color: selected
                ? const Color(0xff2fd276)
                : Colors.white.withValues(alpha: 0.24),
          ),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
        onPressed: onTap,
        icon: Icon(icon, size: 18),
        label: Text(label),
      ),
    );
  }
}

class _InterestChip extends StatelessWidget {
  const _InterestChip(this.label);

  final String label;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(999),
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
