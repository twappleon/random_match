<template>
  <main class="screen">
    <section v-if="!clientEntered" class="page page-shell auth-page">
      <div class="auth-card" aria-label="login and age confirmation">
        <div class="auth-brand">
          <span>Findu</span>
          <strong>18+ 真人兴趣配对</strong>
          <p>完成登录与年龄确认后，进入推荐、广场、视频、消息和我的。</p>
        </div>
        <div class="auth-tabs">
          <button type="button" :class="{ active: authMode === 'login' }" @click="authMode = 'login'">登录</button>
          <button type="button" :class="{ active: authMode === 'register' }" @click="authMode = 'register'">注册</button>
        </div>
        <div class="auth-form">
          <label v-if="authMode === 'register'">
            昵称
            <input v-model.trim="profileForm.displayName" maxlength="24" placeholder="Leon" />
          </label>
          <label>
            手机号 / 邮箱
            <input v-model.trim="authAccount" autocomplete="email" placeholder="leon@example.com" />
          </label>
          <label>
            {{ authMode === 'login' ? '密码' : '设置密码' }}
            <input v-model="authPassword" type="password" autocomplete="current-password" placeholder="••••••••" />
          </label>
          <label class="age-check" :class="{ attention: ageCheckAttention }">
            <input v-model="profileForm.ageConfirmed" type="checkbox" />
            <span>我确认已满 18 岁，并同意社区规则与隐私政策。</span>
          </label>
          <button class="auth-submit" type="button" :disabled="authLoading" @click="enterClientApp">
            {{ authLoading ? '进入中' : authMode === 'login' ? '登录并进入' : '创建账号' }}
          </button>
          <div class="auth-alt">
            <button type="button" @click="showClientSheet('验证码登录', '输入手机号或邮箱后，系统会发送一次性验证码。')">验证码登录</button>
            <button type="button" @click="showClientSheet('第三方注册', '第三方注册后仍需完成 18+ 确认，才能进入核心社交功能。')">Google / Apple</button>
          </div>
        </div>
      </div>
    </section>

    <section v-show="clientEntered && activePage === 'recommend'" class="page page-shell">
      <div class="social-page" aria-label="smart recommendations">
        <header class="social-head">
          <div>
            <strong>推荐</strong>
            <span>真人认证、共同兴趣和在线状态优先</span>
          </div>
          <button type="button" @click="showClientSheet('筛选偏好', '可设置地区、距离、语言、共同兴趣和真人状态。精准筛选会使用 Gems 或会员权益。')">筛选</button>
        </header>
        <div class="secondary-menu">
          <button type="button" :class="{ active: recommendTab === 'smart' }" @click="recommendTab = 'smart'">智能推荐</button>
          <button type="button" :class="{ active: recommendTab === 'nearby' }" @click="recommendTab = 'nearby'">附近</button>
          <button type="button" :class="{ active: recommendTab === 'new' }" @click="recommendTab = 'new'">新人</button>
        </div>
        <article v-if="recommendTab === 'smart'" class="recommend-hero">
          <div class="verified-row"><span></span> 真人认证 · 刚刚在线</div>
          <div class="recommend-person">
            <strong>{{ currentRecommendation.name }}, {{ currentRecommendation.age }}</strong>
            <em>{{ currentRecommendation.distance }}</em>
          </div>
          <div class="tags">
            <span v-for="item in currentRecommendation.tags" :key="item">{{ item }}</span>
          </div>
          <div class="action-row">
            <button type="button" @click="skipRecommendation">跳过</button>
            <button class="primary-action" type="button" @click="greetRecommendation">打招呼</button>
          </div>
        </article>
        <section v-else-if="recommendTab === 'nearby'" class="social-list">
          <article v-for="user in nearbyPeople" :key="user.name" class="social-row" @click="showProfileSheet(user.name)">
            <div class="avatar">{{ user.name.slice(0, 1) }}</div>
            <div><strong>{{ user.name }}, {{ user.age }}</strong><span>共同兴趣：{{ user.tags.join('、') }}</span></div>
            <b>{{ user.distance }}</b>
          </article>
        </section>
        <section v-else class="social-list">
          <article class="feed-item">
            <div class="feed-author"><div class="avatar">H</div><div><strong>Hana, 22</strong><span>真人认证中 · 6 分钟前加入</span></div></div>
            <p>想找人练口语，也喜欢城市夜景和独立音乐。</p>
            <div class="feed-actions"><button @click="showClientSheet('打招呼', 'Hi Hana，我也喜欢独立音乐。今晚有空聊 10 分钟吗？')">打招呼</button><button @click="showProfileSheet('Hana')">查看主页</button><button @click="showClientSheet('已跳过', '新人推荐会继续结合兴趣、审核状态和拉黑关系更新。')">跳过</button></div>
          </article>
        </section>
      </div>
    </section>

    <section v-show="clientEntered && activePage === 'square'" class="page page-shell">
      <div class="social-page" aria-label="public square">
        <header class="social-head">
          <div><strong>广场</strong><span>动态、话题、点赞评论和举报</span></div>
          <button type="button" @click="showClientSheet('发布动态', '可以添加兴趣话题、图片或短视频，媒体内容会先进入审核。')">发布</button>
        </header>
        <div class="secondary-menu">
          <button v-for="tab in squareTabs" :key="tab.id" type="button" :class="{ active: squareTab === tab.id }" @click="squareTab = tab.id">{{ tab.label }}</button>
        </div>
        <div class="square-composer" @click="showClientSheet('发布动态', '分享今天想遇见怎样的人。')">
          <div class="avatar">{{ profileInitial }}</div>
          <span>分享今天想遇见怎样的人</span>
          <button type="button">发布</button>
        </div>
        <article v-for="item in visibleFeedItems" :key="item.id" class="feed-item">
          <div class="feed-author"><div class="avatar">{{ item.name.slice(0, 1) }}</div><div><strong>{{ item.name }}</strong><span>{{ item.meta }}</span></div></div>
          <p>{{ item.copy }}</p>
          <div v-if="item.photos" class="photo-grid"><div></div><div></div><div></div></div>
          <div class="feed-actions">
            <button type="button" @click="likeFeedItem(item.id)">喜欢 {{ item.likes }}</button>
            <button type="button" @click="showClientSheet('评论', '写一条友善评论，支持表情和 @ 用户。')">评论 {{ item.comments }}</button>
            <button type="button" @click="showClientSheet('举报内容', '请选择原因：骚扰、裸露、诈骗、未成年人风险或其他。')">举报</button>
          </div>
        </article>
      </div>
    </section>

    <section v-show="clientEntered && activePage === 'video'" ref="stage" class="call-stage page">
      <div class="remote-video">
        <video ref="remoteVideo" autoplay playsinline></video>
        <div v-if="status !== 'matched'" class="state">
          <div class="state-card" :class="{ waiting: status === 'waiting' }">
            <span class="state-badge">{{ stateBadgeText }}</span>
            <strong>{{ stateTitle }}</strong>
            <span>{{ stateSubtitle }}</span>
            <div class="state-chips" aria-label="match hints">
              <span>随机视讯</span>
              <span>快速连接</span>
              <span>安全操作</span>
            </div>
          </div>
        </div>
      </div>

      <div
        ref="localPreview"
        class="local-preview"
        :style="localPreviewStyle"
        @pointerdown="startPreviewDrag"
      >
        <video ref="localVideo" class="local-video" autoplay playsinline muted></video>
      </div>

      <aside class="stats-panel" aria-label="runtime stats">
        <span><b>{{ stats.online }}</b><small>在线</small></span>
        <span><b>{{ stats.waiting }}</b><small>等待</small></span>
        <span><b>{{ stats.chatting }}</b><small>聊天</small></span>
      </aside>

      <section v-if="status === 'matched' && !chatOpen && !peerCardHidden" class="peer-card" aria-label="peer profile">
        <button class="peer-card-close" type="button" aria-label="hide peer profile" @click="peerCardHidden = true">
          ×
        </button>
        <div class="peer-main">
          <div class="profile-head">
            <div class="avatar">{{ peerInitial }}</div>
            <div>
              <strong>{{ peerDisplayName }}</strong>
              <span>{{ peerBio }}</span>
            </div>
          </div>
          <div class="tags">
            <span v-for="item in peerInterests" :key="item">{{ item }}</span>
          </div>
        </div>
        <div class="safety-actions">
          <button class="report" :disabled="safetyLoading || reportedPeerId === activePeerId" @click="reportPeer">
            {{ reportedPeerId === activePeerId ? '已举报' : '举报' }}
          </button>
          <button class="block" :disabled="safetyLoading" @click="blockPeer">
            拉黑
          </button>
        </div>
      </section>

      <section v-if="chatOpen" class="chat-sheet" aria-label="text chat">
        <div class="chat-header">
          <strong>文字聊天</strong>
          <span>{{ chatHeaderText }}</span>
          <button type="button" aria-label="close text chat" @click="chatOpen = false">收起</button>
        </div>
        <div ref="chatList" class="chat-list">
          <p v-if="chatMessages.length === 0" class="chat-empty">{{ chatEmptyText }}</p>
          <div
            v-for="message in chatMessages"
            :key="message.id"
            class="chat-message"
            :class="{ mine: message.sender === 'self' }"
          >
            <span>{{ message.text }}</span>
          </div>
        </div>
        <form class="chat-form" @submit.prevent="sendChatMessage">
          <input
            v-model="chatDraft"
            maxlength="500"
            autocomplete="off"
            :placeholder="chatInputPlaceholder"
            :disabled="!canUseChat"
          />
          <button :disabled="!canSendChat">发送</button>
        </form>
      </section>
    </section>

    <section v-show="clientEntered && activePage === 'messages'" class="page page-shell">
      <div class="social-page" aria-label="messages">
        <header class="social-head">
          <div><strong>消息</strong><span>私聊、未读和系统通知</span></div>
          <button type="button" @click="showClientSheet('消息设置', '可设置只接收互相关注、真人认证用户或系统安全通知。')">设置</button>
        </header>
        <div class="secondary-menu">
          <button v-for="tab in messageTabs" :key="tab.id" type="button" :class="{ active: messageTab === tab.id }" @click="messageTab = tab.id">{{ tab.label }}</button>
        </div>
        <section class="social-list">
          <article v-for="item in visibleMessages" :key="item.name" class="social-row" @click="showClientSheet(item.name, item.text)">
            <div class="avatar">{{ item.name.slice(0, 1) }}</div>
            <div><strong>{{ item.name }}</strong><span>{{ item.text }}</span></div>
            <b>{{ item.time }}</b>
          </article>
        </section>
      </div>
    </section>

    <section v-show="clientEntered && activePage === 'me'" class="page page-shell">
      <div class="social-page" aria-label="my account">
        <section class="me-panel">
          <div class="me-head">
            <div class="me-avatar">{{ profileInitial }}</div>
            <div><strong>{{ profileForm.displayName || 'Leon' }}</strong><span>真人认证 · 兴趣 {{ selectedInterests.length }} 个 · 18+ {{ profileForm.ageConfirmed ? '已确认' : '未确认' }}</span></div>
          </div>
          <div class="stats-grid">
            <button type="button" @click="showClientSheet('关注', '我的关注会优先出现在邀请和消息入口。')"><strong>{{ followedUsers.length || 128 }}</strong><span>关注</span></button>
            <button type="button" @click="showClientSheet('动态', '动态会经过内容审核后展示在广场。')"><strong>36</strong><span>动态</span></button>
            <button type="button" @click="showClientSheet('互动分', '互动分来自友善回复、稳定通话和低举报率。')"><strong>4.8</strong><span>互动分</span></button>
          </div>
        </section>
        <div class="secondary-menu">
          <button type="button" :class="{ active: meTab === 'profile' }" @click="meTab = 'profile'">资料</button>
          <button type="button" :class="{ active: meTab === 'privacy' }" @click="meTab = 'privacy'">隐私</button>
          <button type="button" :class="{ active: meTab === 'safety' }" @click="meTab = 'safety'">安全</button>
        </div>
        <section class="settings-list">
          <button v-for="row in visibleSettings" :key="row.label" type="button" @click="handleSetting(row.action, row.label)">
            {{ row.label }} <span>{{ row.value }}</span>
          </button>
        </section>
        <button class="logout-button" type="button" @click="logoutClient">登出</button>
      </div>
    </section>

    <section v-show="clientEntered && activePage === 'profile'" class="page page-shell">
      <div class="content-card profile-page" aria-label="profile setup">
        <div class="profile-head">
          <div class="avatar">{{ profileInitial }}</div>
          <div>
            <strong>匿名身份</strong>
            <span>用兴趣和状态开始匹配</span>
          </div>
        </div>
        <label>
          昵称
          <input v-model.trim="profileForm.displayName" maxlength="24" placeholder="星球旅人" />
        </label>
        <label>
          简介
          <textarea v-model.trim="profileForm.bio" maxlength="120" rows="3" placeholder="一句话介绍现在的你"></textarea>
        </label>
        <label>
          兴趣标签
          <input v-model="interestsText" placeholder="电影, 音乐, 旅行" />
        </label>
        <div class="interest-picker" aria-label="interest suggestions">
          <button
            v-for="item in interestSuggestions"
            :key="item"
            type="button"
            :class="{ active: selectedInterests.includes(item) }"
            @click="toggleInterest(item)"
          >
            {{ item }}
          </button>
        </div>
        <label>
          常用语言
          <select v-model="profileForm.language">
            <option value="zh">中文</option>
            <option value="en">English</option>
            <option value="ja">日本語</option>
            <option value="ko">한국어</option>
            <option value="es">Español</option>
          </select>
        </label>
        <label ref="ageCheckRef" class="age-check" :class="{ attention: ageCheckAttention }">
          <input v-model="profileForm.ageConfirmed" type="checkbox" />
          <span>我已满 18 岁并同意文明视讯</span>
        </label>
        <button class="save-profile" :disabled="savingProfile" @click="saveProfile">
          {{ savingProfile ? '保存中' : '保存资料' }}
        </button>
        <section class="blocked-section" aria-label="blocked users">
          <div class="section-head">
            <div>
              <strong>拉黑名单</strong>
              <span>解除后未来可能再次匹配到对方</span>
            </div>
            <button type="button" :disabled="loadingBlockedUsers" @click="loadBlockedUsers">
              {{ loadingBlockedUsers ? '读取中' : '刷新' }}
            </button>
          </div>
          <p v-if="!loadingBlockedUsers && blockedUsers.length === 0" class="empty-list">目前没有拉黑对象</p>
          <div v-else class="blocked-list">
            <div v-for="item in blockedUsers" :key="item.user.id" class="blocked-user">
              <div class="profile-head">
                <div class="avatar">{{ item.user.displayName.trim().slice(0, 1).toUpperCase() || '星' }}</div>
                <div>
                  <strong>{{ item.user.displayName || '匿名用户' }}</strong>
                  <span>{{ item.user.bio || `拉黑于 ${formatDate(item.createdAt)}` }}</span>
                </div>
              </div>
              <button type="button" :disabled="unblockingUserId === item.user.id" @click="unblockBlockedUser(item.user.id)">
                {{ unblockingUserId === item.user.id ? '解除中' : '解除' }}
              </button>
            </div>
          </div>
        </section>
      </div>
    </section>

    <section v-show="clientEntered && activePage === 'discover'" class="page page-shell">
      <div class="discover-page" aria-label="lounge discover">
        <header class="discover-head">
          <div>
            <strong>Lounge</strong>
            <span>浏览资料、关注、私信，再进入 1v1 连线</span>
          </div>
          <button type="button" :disabled="discoverLoading" @click="loadDiscoverProfiles">
            {{ discoverLoading ? '刷新中' : '刷新' }}
          </button>
        </header>

        <section class="discover-filters" aria-label="discover filters">
          <div class="mode-switch">
            <button type="button" :class="{ active: mode === 'video' }" @click="mode = 'video'">视讯</button>
            <button type="button" :class="{ active: mode === 'voice' }" @click="mode = 'voice'">语音</button>
          </div>
          <label>
            地区
            <select v-model="selectedRegion" @change="loadDiscoverProfiles">
              <option value="global">全球</option>
              <option value="tw">台湾</option>
              <option value="jp">日本</option>
              <option value="kr">韩国</option>
              <option value="us">美国</option>
            </select>
          </label>
          <label>
            对象
            <select v-model="genderPreference" @change="loadDiscoverProfiles">
              <option value="everyone">不限</option>
              <option value="female">女性</option>
              <option value="male">男性</option>
            </select>
          </label>
          <p>基础随机免费，精准筛选可用 Gems 解锁</p>
        </section>

        <section v-if="followedUsers.length > 0" class="followed-section" aria-label="followed users">
          <div class="section-head">
            <div>
              <strong>我的关注</strong>
              <span>点头像可以用相同偏好进入连线</span>
            </div>
          </div>
          <div class="followed-list">
            <button
              v-for="user in followedUsers"
              :key="user.id"
              class="followed-user"
              type="button"
              @click="startFromProfile(user)"
            >
              <span class="avatar">{{ userInitial(user) }}</span>
              <strong>{{ user.displayName || '星球旅人' }}</strong>
              <small>{{ languageLabel(user.language || 'zh') }}</small>
            </button>
          </div>
        </section>

        <section class="discover-list" aria-label="discover profiles">
          <p v-if="discoverLoading && discoverProfiles.length === 0" class="empty-list">正在读取 Lounge 列表</p>
          <p v-else-if="discoverProfiles.length === 0" class="empty-list">目前没有符合条件的对象，换个地区或对象试试</p>
          <article v-for="user in discoverProfiles" :key="user.id" class="discover-card">
            <div class="discover-user">
              <div class="avatar">{{ userInitial(user) }}</div>
              <div>
                <strong>
                  {{ user.displayName || '星球旅人' }}
                  <span v-if="user.trustBadge" class="trust-badge">✓</span>
                </strong>
                <span>{{ regionLabel(user.region || 'global') }} · {{ user.bio || '愿意认识新朋友' }}</span>
              </div>
            </div>
            <div class="tags">
              <span v-for="item in profileInterests(user)" :key="`${user.id}-${item}`">{{ item }}</span>
            </div>
            <div class="discover-actions">
              <button type="button" @click="toggleFollow(user)">
                {{ isFollowing(user.id) ? '取消关注' : '关注' }}
              </button>
              <button type="button" @click="openDirectMessage(user)">私信</button>
              <button type="button" @click="dismissDiscoverProfile(user.id)">换一批</button>
              <button class="connect" type="button" @click="startFromProfile(user)">以此偏好连接</button>
            </div>
          </article>
        </section>
      </div>
    </section>

    <section v-show="clientEntered && activePage === 'membership'" class="page page-shell">
      <div class="membership-page" aria-label="membership">
        <section class="membership-hero">
          <div class="membership-orb" aria-hidden="true">◆</div>
          <div class="membership-summary">
            <span class="eyebrow">MATCH PASS</span>
            <strong>{{ membershipTitle }}</strong>
            <span>{{ membershipText }}</span>
          </div>
          <button class="membership-cta" :disabled="paymentLoading || commerceStatus?.isMember" @click="buyMembership">
            {{ paymentButtonText }}
          </button>
        </section>

        <section class="membership-status" aria-label="membership status">
          <div class="membership-status-head">
            <div>
              <strong>{{ membershipStatusTitle }}</strong>
              <span>{{ membershipStatusText }}</span>
            </div>
            <button type="button" :disabled="paymentLoading" @click="loadCommerceStatus">
              {{ paymentLoading ? '同步中' : '同步状态' }}
            </button>
          </div>
          <div class="quota-track" aria-hidden="true">
            <span :style="{ width: quotaProgress }"></span>
          </div>
          <div class="status-pills">
            <span><small>匹配额度</small><b>{{ quotaLabel }}</b></span>
            <span><small>Gems</small><b>{{ commerceStatus?.gemsBalance ?? 0 }}</b></span>
            <span><small>优先队列</small><b>{{ commerceStatus?.priorityQueue ? '已开启' : '未开启' }}</b></span>
            <span v-if="commerceStatus?.membershipExpiresAt"><small>到期日</small><b>{{ formatDate(commerceStatus.membershipExpiresAt) }}</b></span>
          </div>
        </section>

        <section class="pass-grid" aria-label="membership packages">
          <button class="pass-card" type="button" :disabled="paymentLoading || commerceStatus?.isMember" @click="buyMembership">
            <span class="pass-gem">◆</span>
            <strong>月度会员</strong>
            <small>30 天会员权益，可续期叠加</small>
            <b>{{ commerceStatus?.isMember ? '已拥有' : '$6.99' }}</b>
          </button>
          <button class="pass-card featured" type="button" :disabled="paymentLoading || commerceStatus?.isMember" @click="buyMembership">
            <span class="pass-ribbon">推荐</span>
            <span class="pass-gem">◆</span>
            <strong>优先通行</strong>
            <small>无限匹配 + 优先队列 + 300 Gems</small>
            <b>{{ commerceStatus?.isMember ? '已开通' : '$6.99/月' }}</b>
          </button>
          <button class="pass-card" type="button" :disabled="paymentLoading || commerceStatus?.isMember" @click="buyMembership">
            <span class="pass-gem pink">◆</span>
            <strong>畅聊模式</strong>
            <small>额度用完后继续使用</small>
            <b>{{ commerceStatus?.isMember ? '已包含' : '会员包含' }}</b>
          </button>
        </section>

        <section class="benefits">
          <div>
            <strong>无限随机匹配</strong>
            <span>不受每日免费次数限制</span>
          </div>
          <div>
            <strong>进入优先队列</strong>
            <span>高峰期减少等待时间</span>
          </div>
          <div>
            <strong>安全加速体验</strong>
            <span>保留举报、拉黑和离开控制</span>
          </div>
        </section>

        <footer class="payment-note">
          <span>安全交易</span>
          <span>加密支付</span>
          <span>可随时确认状态</span>
        </footer>
      </div>
    </section>

    <section v-show="clientEntered && activePage === 'guide'" class="page page-shell">
      <div class="content-card guide-page" aria-label="guide">
        <div class="guide-hero">
          <strong>Random Match 使用指南</strong>
          <span>匿名视讯适合轻松认识新朋友，也需要清楚的安全边界和使用方式。</span>
        </div>
        <article>
          <h2>开始视讯前</h2>
          <p>建议先填写昵称、简介和兴趣标签，让对方知道可以从哪里展开话题。请只使用你愿意公开给陌生人看到的资料，不要在个人简介里留下电话、住址、帐号密码、支付资讯或其他敏感内容。</p>
        </article>
        <article>
          <h2>如何保持安全</h2>
          <p>如果对方让你不舒服，可以立即离开、举报或拉黑。拉黑后系统会记录关系，并尽量避免再次把你们匹配到一起。遇到骚扰、裸露、威胁、诈骗或未成年人相关风险时，请优先使用举报功能。</p>
        </article>
        <article>
          <h2>会员与匹配额度</h2>
          <p>免费用户每天有固定随机匹配次数。会员可以无限匹配，并进入优先队列，适合想减少等待时间的用户。付费前请确认你了解服务内容和当前价格。</p>
        </article>
        <div class="guide-links">
          <a href="/about.html" target="_blank" rel="noreferrer">关于 Random Match</a>
          <a href="/safety.html" target="_blank" rel="noreferrer">安全指南</a>
          <a href="/privacy.html" target="_blank" rel="noreferrer">隐私政策</a>
          <a href="/terms.html" target="_blank" rel="noreferrer">服务条款</a>
        </div>
      </div>
    </section>

    <nav v-if="clientEntered && activePage === 'video'" class="toolbar" aria-label="match controls">
      <button class="chat-toggle" @click="toggleChat">
        {{ chatOpen ? '收起文字' : '文字' }}
      </button>
      <button class="camera" :disabled="loading || switchingCamera || !localStream" @click="switchCamera">
        {{ switchingCamera ? '切换中' : nextCameraText }}
      </button>
      <button class="block-call" :disabled="!canBlockPeer" @click="blockPeer">
        {{ safetyLoading ? '处理中' : '拉黑' }}
      </button>
      <button class="primary" :disabled="loading || status === 'waiting'" @click="startMatch">
        {{ actionText }}
      </button>
      <button class="danger" :disabled="leaving || status === 'idle'" @click="leaveCall">
        {{ leaving ? '退出中' : '退出' }}
      </button>
    </nav>

    <nav v-if="clientEntered" class="app-nav" aria-label="main navigation">
      <button :class="{ active: activePage === 'recommend' }" @click="switchPage('recommend')">推荐</button>
      <button :class="{ active: activePage === 'square' }" @click="switchPage('square')">广场</button>
      <button :class="{ active: activePage === 'video' }" @click="switchPage('video')">视频</button>
      <button :class="{ active: activePage === 'messages' }" @click="switchPage('messages')">消息</button>
      <button :class="{ active: activePage === 'me' }" @click="switchPage('me')">我的</button>
    </nav>

    <div v-if="sheetOpen" class="client-sheet-overlay" @click.self="closeClientSheet">
      <section class="client-sheet" role="dialog" aria-modal="true" :aria-label="sheetTitle">
        <div class="sheet-head">
          <strong>{{ sheetTitle }}</strong>
          <button type="button" aria-label="关闭" @click="closeClientSheet">×</button>
        </div>
        <p>{{ sheetBody }}</p>
        <div class="sheet-actions">
          <button type="button" @click="closeClientSheet">取消</button>
          <button class="primary-action" type="button" @click="confirmClientSheet">确认</button>
        </div>
      </section>
    </div>

    <button v-if="chatToastText" class="chat-toast" type="button" @click="openChatFromToast">
      <strong>新文字讯息</strong>
      <span>{{ chatToastText }}</span>
    </button>

    <p v-if="errorText" class="error" :class="{ 'error-with-toolbar': activePage === 'video' }" role="alert">{{ errorText }}</p>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import { anonymousAuth, blockUser, confirmPaymentOrder, createPaymentOrder, fetchBlockedUsers, fetchCommerceStatus, fetchDiscoverProfiles, fetchFollowedUsers, fetchProfile, fetchStats, followUser, iceServers, joinMatch, leaveMatch, reportUser, savePushSubscription, sendDirectMessage, sendPushTest, unblockUser, unfollowUser, updateProfile, uploadMatchSnapshot, vapidPublicKey, verifySession, type BlockedUser, type CommerceStatus, type MatchMode, type UserProfile, wsURL } from './api'
import { initAnalytics } from './firebase'

type Status = 'idle' | 'waiting' | 'matched'
type Page = 'recommend' | 'square' | 'video' | 'messages' | 'me' | 'discover' | 'profile' | 'membership' | 'guide'
type LabeledTab = { id: string; label: string }
type MockPerson = { name: string; age: number; distance: string; tags: string[] }
type FeedItem = { id: string; scope: string; name: string; meta: string; copy: string; likes: number; comments: number; photos?: boolean }
type MessagePreview = { scope: string; name: string; text: string; time: string }
type ChatMessage = {
  id: string
  sender: 'self' | 'peer'
  text: string
  createdAt: string
}

const mode = ref<MatchMode>('video')
const clientEntered = ref(localStorage.getItem('clientEntered') === 'true')
const activePage = ref<Page>(clientEntered.value ? 'recommend' : 'video')
const authMode = ref<'login' | 'register'>('login')
const authAccount = ref('')
const authPassword = ref('')
const authLoading = ref(false)
const recommendTab = ref<'smart' | 'nearby' | 'new'>('smart')
const squareTab = ref('recommend')
const messageTab = ref('all')
const meTab = ref<'profile' | 'privacy' | 'safety'>('profile')
const recommendationIndex = ref(0)
const sheetOpen = ref(false)
const sheetTitle = ref('详情')
const sheetBody = ref('')
const status = ref<Status>('idle')
const loading = ref(false)
const leaving = ref(false)
const switchingCamera = ref(false)
const savingProfile = ref(false)
const safetyLoading = ref(false)
const paymentLoading = ref(false)
const loadingBlockedUsers = ref(false)
const discoverLoading = ref(false)
const reportedPeerId = ref<string | null>(null)
const unblockingUserId = ref<string | null>(null)
const pushStatus = ref<'idle' | 'enabled' | 'blocked' | 'unsupported' | 'unconfigured'>('idle')
const errorText = ref('')
const profile = ref<UserProfile | null>(null)
const peerProfile = ref<UserProfile | null>(null)
const blockedUsers = ref<BlockedUser[]>([])
const discoverProfiles = ref<UserProfile[]>([])
const followedUsers = ref<UserProfile[]>([])
const commerceStatus = ref<CommerceStatus | null>(null)
const profileForm = ref({
  displayName: '星球旅人',
  bio: '',
  language: 'zh',
  ageConfirmed: false
})
const interestsText = ref('聊天, 电影, 音乐')
const interestSuggestions = ['聊天', '电影', '音乐', '旅行', '美食', '运动', '游戏', '动漫', '摄影', '宠物', '读书', '咖啡', '健身', '语言交换', '科技', '深夜电台']
const selectedRegion = ref('global')
const genderPreference = ref('everyone')
const stats = ref({ online: 0, waiting: 0, chatting: 0 })
const statsTimer = ref<number | null>(null)
const token = ref(localStorage.getItem('token') ?? '')
const ws = ref<WebSocket | null>(null)
const wsHeartbeatTimer = ref<number | null>(null)
const closingSocket = ref(false)
const activeRoomId = ref<string | null>(null)
const chatOpen = ref(false)
const chatDraft = ref('')
const chatMessages = ref<ChatMessage[]>([])
const chatToastText = ref('')
const chatToastTimer = ref<number | null>(null)
const localStream = ref<MediaStream | null>(null)
const peer = ref<RTCPeerConnection | null>(null)
const peerDisconnectTimer = ref<number | null>(null)
const activePeerId = ref<string | null>(null)
const peerCardHidden = ref(false)
const pendingCandidates = ref<RTCIceCandidateInit[]>([])
const stage = ref<HTMLElement | null>(null)
const chatList = ref<HTMLElement | null>(null)
const localPreview = ref<HTMLElement | null>(null)
const ageCheckRef = ref<HTMLElement | null>(null)
const remoteVideo = ref<HTMLVideoElement | null>(null)
const localVideo = ref<HTMLVideoElement | null>(null)
const ageCheckAttention = ref(false)
const previewPosition = ref({ x: 0, y: 0 })
const previewPositioned = ref(false)
const previewDrag = ref<{
  pointerId: number
  startX: number
  startY: number
  originX: number
  originY: number
} | null>(null)
const capturedSnapshotRooms = new Set<string>()
const cameraFacing = ref<'user' | 'environment'>('user')

const stateBadgeText = computed(() => status.value === 'waiting' ? 'LIVE MATCH' : 'AURORA READY')
const stateTitle = computed(() => status.value === 'waiting' ? '正在寻找新朋友' : '今晚遇见新朋友')
const stateSubtitle = computed(() => status.value === 'waiting' ? '保持页面开启，匹配成功后会自动进入视讯。' : '更强的氛围感，快速连接在线用户。')

const actionText = computed(() => {
  if (loading.value) return '匹配中'
  if (status.value === 'waiting') return '等待中'
  if (status.value === 'matched') return '已连线'
  return '随机匹配'
})

const localPreviewStyle = computed(() => {
  if (!previewPositioned.value) return {}
  return {
    left: `${previewPosition.value.x}px`,
    top: `${previewPosition.value.y}px`
  }
})

const nextCameraText = computed(() => cameraFacing.value === 'user' ? '后镜头' : '前镜头')
const profileInitial = computed(() => (profileForm.value.displayName || '星').trim().slice(0, 1).toUpperCase())
const peerInitial = computed(() => (peerProfile.value?.displayName || '星').trim().slice(0, 1).toUpperCase())
const peerDisplayName = computed(() => peerProfile.value?.displayName || '对方资料载入中')
const peerBio = computed(() => peerProfile.value?.bio || '对方暂时没有填写简介')
const peerInterests = computed(() => peerProfile.value?.interests?.length ? peerProfile.value.interests : ['随机视讯'])
const membershipTitle = computed(() => commerceStatus.value?.isMember ? '会员已开启' : '免费匹配额度')
const membershipText = computed(() => {
  const status = commerceStatus.value
  if (!status) return '正在读取今日额度'
  if (status.isMember) return `无限匹配 · 优先排队${status.membershipExpiresAt ? ` · 到期 ${formatDate(status.membershipExpiresAt)}` : ''}`
  return `今日剩余 ${status.dailyRemaining}/${status.dailyLimit} 次 · 会员无限匹配并优先排队`
})
const paymentButtonText = computed(() => {
  if (commerceStatus.value?.isMember) return '已是会员'
  return paymentLoading.value ? '开通中' : '$6.99/月 开通'
})
const membershipStatusTitle = computed(() => commerceStatus.value?.isMember ? '当前权益：Match Pass' : '当前权益：免费账户')
const membershipStatusText = computed(() => {
  const status = commerceStatus.value
  if (!status) return '正在同步会员、Gems 和今日额度'
  if (status.isMember) return '无限匹配、优先队列和精准筛选已开启'
  return '免费账户保留基础随机匹配，开通后解除每日次数限制'
})
const quotaLabel = computed(() => {
  const status = commerceStatus.value
  if (!status) return '同步中'
  if (status.isMember) return '无限匹配'
  return `剩余 ${status.dailyRemaining}/${status.dailyLimit} 次`
})
const quotaProgress = computed(() => {
  const status = commerceStatus.value
  if (!status) return '0%'
  if (status.isMember || status.dailyLimit <= 0) return '100%'
  const used = Math.max(0, Math.min(status.dailyUsed, status.dailyLimit))
  return `${(used / status.dailyLimit) * 100}%`
})
const canSendChat = computed(() => status.value === 'matched' && Boolean(activePeerId.value) && chatDraft.value.trim().length > 0)
const canUseChat = computed(() => status.value === 'matched' && Boolean(activePeerId.value))
const canBlockPeer = computed(() => status.value === 'matched' && Boolean(activePeerId.value) && !safetyLoading.value && !leaving.value)
const chatEmptyText = computed(() => canUseChat.value ? '开始文字聊天' : '匹配成功后可文字聊天')
const chatInputPlaceholder = computed(() => canUseChat.value ? '输入消息...' : '等待匹配后开始聊天')
const chatHeaderText = computed(() => canUseChat.value ? peerDisplayName.value : '目前尚未连接对象')
const selectedInterests = computed(() => parsedInterests())
const followedUserIds = computed(() => new Set(followedUsers.value.map((user) => user.id)))
const recommendations = ref<MockPerson[]>([
  { name: 'Mina', age: 24, distance: '2.8 km', tags: ['旅行', '电影', '英语聊天'] },
  { name: 'Ariel', age: 25, distance: '1.6 km', tags: ['独立音乐', '咖啡', '摄影'] },
  { name: 'Ruby', age: 23, distance: '3.6 km', tags: ['健身', '动漫', '夜景'] }
])
const nearbyPeople: MockPerson[] = [
  { name: 'Nana', age: 26, distance: '1.2 km', tags: ['咖啡', '散步', '英语聊天'] },
  { name: 'Ruby', age: 23, distance: '3.6 km', tags: ['城市夜景', '电影', '摄影'] }
]
const squareTabs: LabeledTab[] = [
  { id: 'recommend', label: '推荐' },
  { id: 'nearby', label: '附近' },
  { id: 'following', label: '关注' },
  { id: 'latest', label: '最新' }
]
const messageTabs: LabeledTab[] = [
  { id: 'all', label: '全部' },
  { id: 'unread', label: '未读' },
  { id: 'system', label: '系统' }
]
const feedItems = ref<FeedItem[]>([
  { id: 'feed-1', scope: 'recommend', name: 'Mina', meta: '#电影 · 推荐', copy: '想找一个也喜欢独立电影的人，今晚可以先语音聊聊。', likes: 238, comments: 42, photos: true },
  { id: 'feed-2', scope: 'nearby', name: 'Ken', meta: '#语言交换 · 附近', copy: '中文和英文都可以，想练 20 分钟轻松聊天。', likes: 82, comments: 11 },
  { id: 'feed-3', scope: 'following', name: 'Ariel', meta: '你关注的人 · 2 小时前', copy: '今晚的歌单更新了，想找人一起听 30 分钟。', likes: 419, comments: 58 },
  { id: 'feed-4', scope: 'latest', name: 'Leo', meta: '刚刚发布 · 待审核图已隐藏', copy: '刚到新城市，想知道附近有什么适合聊天的咖啡店。', likes: 9, comments: 2 }
])
const messagePreviews: MessagePreview[] = [
  { scope: 'all unread', name: 'Mina', text: '我也喜欢那部电影，晚上可以聊聊。', time: '09:32' },
  { scope: 'all', name: 'Ariel', text: '视频通话质量很好，下次继续。', time: '昨天' },
  { scope: 'all system', name: 'System', text: '你的动态已通过审核并公开展示。', time: '周一' },
  { scope: 'unread', name: 'Ken', text: '附近咖啡店我有推荐。', time: '08:10' },
  { scope: 'system', name: 'Safety', text: '举报已受理，处理结果会通过系统消息通知。', time: '上周' }
]
const currentRecommendation = computed(() => recommendations.value[recommendationIndex.value % recommendations.value.length])
const visibleFeedItems = computed(() => feedItems.value.filter((item) => squareTab.value === 'recommend' ? item.scope === 'recommend' : item.scope === squareTab.value))
const visibleMessages = computed(() => messagePreviews.filter((item) => item.scope.split(' ').includes(messageTab.value)))
const visibleSettings = computed(() => {
  if (meTab.value === 'privacy') {
    return [
      { label: '谁可以看到我的距离', value: '仅匹配', action: 'sheet' },
      { label: '谁可以邀请视频', value: '互相关注', action: 'sheet' },
      { label: '黑名单', value: String(blockedUsers.value.length || 12), action: 'profile' }
    ]
  }
  if (meTab.value === 'safety') {
    return [
      { label: '18+ 确认', value: profileForm.value.ageConfirmed ? '已完成' : '未完成', action: 'profile' },
      { label: '真人认证', value: '审核通过', action: 'sheet' },
      { label: '举报记录', value: '2', action: 'sheet' },
      { label: '账号安全', value: '›', action: 'guide' }
    ]
  }
  return [
    { label: '资料编辑', value: '›', action: 'profile' },
    { label: '兴趣标签', value: `${selectedInterests.value.length} 个`, action: 'profile' },
    { label: '认证中心', value: '›', action: 'sheet' },
    { label: 'Match Pass', value: commerceStatus.value?.isMember ? '已开通' : '免费账户', action: 'membership' }
  ]
})

initAnalytics()

async function refreshStats() {
  try {
    stats.value = await fetchStats()
  } catch {
    // Keep the last visible values if a short network hiccup happens.
  }
}

function startStatsPolling() {
  void refreshStats()
  statsTimer.value = window.setInterval(refreshStats, 5000)
}

function stopStatsPolling() {
  if (statsTimer.value !== null) {
    window.clearInterval(statsTimer.value)
    statsTimer.value = null
  }
}

function startSocketHeartbeat(socket: WebSocket) {
  stopSocketHeartbeat()
  wsHeartbeatTimer.value = window.setInterval(() => {
    if (ws.value !== socket || socket.readyState !== WebSocket.OPEN) {
      stopSocketHeartbeat()
      return
    }
    socket.send(JSON.stringify({ type: 'ping' }))
  }, 25000)
}

function stopSocketHeartbeat() {
  if (wsHeartbeatTimer.value !== null) {
    window.clearInterval(wsHeartbeatTimer.value)
    wsHeartbeatTimer.value = null
  }
}

async function switchPage(page: Page) {
  activePage.value = page
  if (page !== 'video') chatOpen.value = false
  if (page === 'video') {
    await nextTick()
    ensurePreviewPosition()
  }
  if (page === 'membership') {
    void loadCommerceStatus().catch(() => undefined)
  }
  if (page === 'discover') {
    void loadDiscoverProfiles().catch(() => undefined)
  }
  if (page === 'profile') {
    void loadBlockedUsers().catch(() => undefined)
  }
}

async function enterClientApp() {
  authLoading.value = true
  errorText.value = ''
  try {
    if (!profileForm.value.ageConfirmed) {
      ageCheckAttention.value = true
      errorText.value = '请先确认已满 18 岁'
      window.setTimeout(() => {
        ageCheckAttention.value = false
      }, 2200)
      return
    }
    try {
      await ensureAuth()
      await persistProfile()
    } catch (error) {
      errorText.value = `${toUserMessage(error)}，已先进入客户端演示模式`
    }
    clientEntered.value = true
    localStorage.setItem('clientEntered', 'true')
    await switchPage('recommend')
    showClientSheet('欢迎来到 Findu', '推荐、广场、视频、消息和我的页面已开启。')
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    authLoading.value = false
  }
}

function showClientSheet(title: string, body: string) {
  sheetTitle.value = title
  sheetBody.value = body
  sheetOpen.value = true
}

function closeClientSheet() {
  sheetOpen.value = false
}

function confirmClientSheet() {
  closeClientSheet()
  errorText.value = '已保存'
  window.setTimeout(() => {
    if (errorText.value === '已保存') errorText.value = ''
  }, 1800)
}

function skipRecommendation() {
  recommendationIndex.value = (recommendationIndex.value + 1) % recommendations.value.length
  showClientSheet('已跳过', '推荐会继续结合兴趣、互动质量、内容审核状态和拉黑关系。')
}

function greetRecommendation() {
  const person = currentRecommendation.value
  showClientSheet('打招呼', `Hi ${person.name}，我也喜欢${person.tags.slice(0, 2).join('和')}。今晚有空聊 10 分钟吗？`)
}

function showProfileSheet(name: string) {
  showClientSheet(name, '真人认证 · 共同兴趣 4 个 · 支持关注、私聊和视频邀请。')
}

function likeFeedItem(id: string) {
  const item = feedItems.value.find((entry) => entry.id === id)
  if (!item) return
  item.likes += 1
}

function handleSetting(action: string, label: string) {
  if (action === 'profile') {
    void switchPage('profile')
    return
  }
  if (action === 'membership') {
    void switchPage('membership')
    return
  }
  if (action === 'guide') {
    void switchPage('guide')
    return
  }
  showClientSheet(label, '此设置会影响推荐、广场、视频邀请和消息权限。')
}

function toggleChat() {
  chatOpen.value = !chatOpen.value
  if (chatOpen.value) {
    clearChatToast()
    scrollChatToBottom()
  }
}

function openChatFromToast() {
  activePage.value = 'video'
  chatOpen.value = true
  clearChatToast()
  scrollChatToBottom()
}

function clearToken() {
  token.value = ''
  localStorage.removeItem('token')
}

function logoutClient() {
  closeSocket()
  stopLocalMedia()
  resetCall()
  clearToken()
  clientEntered.value = false
  localStorage.removeItem('clientEntered')
  activePage.value = 'video'
  authMode.value = 'login'
  chatOpen.value = false
  sheetOpen.value = false
  errorText.value = '已登出'
  window.setTimeout(() => {
    if (errorText.value === '已登出') errorText.value = ''
  }, 1800)
}

async function ensureAuth() {
  if (token.value && (await verifySession(token.value))) return
  clearToken()
  const auth = await anonymousAuth()
  token.value = auth.token
  localStorage.setItem('token', auth.token)
  setProfile(auth.user)
}

function setProfile(nextProfile: UserProfile) {
  profile.value = nextProfile
  profileForm.value = {
    displayName: nextProfile.displayName || '星球旅人',
    bio: nextProfile.bio || '',
    language: nextProfile.language || 'zh',
    ageConfirmed: Boolean(nextProfile.ageConfirmed)
  }
  interestsText.value = (nextProfile.interests?.length ? nextProfile.interests : ['聊天', '电影', '音乐']).join(', ')
}

async function loadProfile() {
  try {
    await ensureAuth()
    setProfile(await fetchProfile(token.value))
    await loadCommerceStatus()
    await loadBlockedUsers()
    await loadFollowedUsers()
  } catch {
    // Profile is refreshed again before matching.
  }
}

async function loadCommerceStatus() {
  await ensureAuth()
  commerceStatus.value = await fetchCommerceStatus(token.value)
}

async function loadBlockedUsers() {
  loadingBlockedUsers.value = true
  try {
    await ensureAuth()
    blockedUsers.value = await fetchBlockedUsers(token.value)
  } finally {
    loadingBlockedUsers.value = false
  }
}

async function loadDiscoverProfiles() {
  discoverLoading.value = true
  try {
    await ensureAuth()
    const users = await fetchDiscoverProfiles(token.value, {
      region: selectedRegion.value,
      gender: genderPreference.value
    })
    discoverProfiles.value = users
    await loadFollowedUsers().catch(() => {
      followedUsers.value = []
    })
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    discoverLoading.value = false
  }
}

async function loadFollowedUsers() {
  await ensureAuth()
  followedUsers.value = await fetchFollowedUsers(token.value)
}

function isFollowing(userId: string) {
  return followedUserIds.value.has(userId)
}

async function toggleFollow(user: UserProfile) {
  if (!user.id) return
  errorText.value = ''
  try {
    await ensureAuth()
    if (isFollowing(user.id)) {
      await unfollowUser(token.value, user.id)
      followedUsers.value = followedUsers.value.filter((item) => item.id !== user.id)
      errorText.value = '已取消关注'
      return
    }
    await followUser(token.value, user.id)
    if (!isFollowing(user.id)) followedUsers.value = [user, ...followedUsers.value]
    errorText.value = `${user.displayName || '星球旅人'} 会出现在我的关注中`
  } catch (error) {
    errorText.value = toUserMessage(error)
  }
}

async function openDirectMessage(user: UserProfile) {
  if (!user.id) return
  const text = window.prompt(`发送私信给 ${user.displayName || '星球旅人'}`, '你好，想认识你')
  if (!text?.trim()) return
  errorText.value = ''
  try {
    await ensureAuth()
    await sendDirectMessage(token.value, user.id, text.trim())
    errorText.value = '私信已送出'
  } catch (error) {
    errorText.value = toUserMessage(error)
  }
}

function dismissDiscoverProfile(userId: string) {
  discoverProfiles.value = discoverProfiles.value.filter((user) => user.id !== userId)
  if (discoverProfiles.value.length <= 2) void loadDiscoverProfiles().catch(() => undefined)
}

async function startFromProfile(user: UserProfile) {
  if (user.region) selectedRegion.value = user.region
  if (user.gender === 'female' || user.gender === 'male') genderPreference.value = user.gender
  if (user.language) profileForm.value.language = user.language
  if (user.interests?.length) interestsText.value = user.interests.slice(0, 6).join(', ')
  peerProfile.value = user
  await startMatch()
}

function userInitial(user: UserProfile) {
  return (user.displayName || '星').trim().slice(0, 1).toUpperCase()
}

function profileInterests(user: UserProfile) {
  return user.interests?.length ? user.interests.slice(0, 4) : ['聊天', '电影', '音乐']
}

function regionLabel(region: string) {
  const labels: Record<string, string> = {
    global: '全球',
    tw: '台湾',
    jp: '日本',
    kr: '韩国',
    us: '美国'
  }
  return labels[region] || region
}

function languageLabel(language: string) {
  const labels: Record<string, string> = {
    zh: '中文',
    en: 'English',
    ja: '日本語',
    ko: '한국어',
    es: 'Español'
  }
  return labels[language] || language
}

async function unblockBlockedUser(userId: string) {
  if (unblockingUserId.value) return
  unblockingUserId.value = userId
  errorText.value = ''
  try {
    await ensureAuth()
    await unblockUser(token.value, userId)
    blockedUsers.value = blockedUsers.value.filter((item) => item.user.id !== userId)
    errorText.value = '已解除拉黑'
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    unblockingUserId.value = null
  }
}

function parsedInterests() {
  const items = interestsText.value
    .split(/[,，]/)
    .map((item) => item.trim())
    .filter(Boolean)
  return Array.from(new Set(items)).slice(0, 6)
}

function toggleInterest(item: string) {
  const items = parsedInterests()
  const existingIndex = items.indexOf(item)
  if (existingIndex >= 0) {
    items.splice(existingIndex, 1)
  } else if (items.length < 6) {
    items.push(item)
  } else {
    errorText.value = '最多选择 6 个兴趣标签'
    return
  }
  interestsText.value = items.join(', ')
}

async function saveProfile() {
  savingProfile.value = true
  errorText.value = ''
  try {
    await persistProfile()
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    savingProfile.value = false
  }
}

async function persistProfile() {
  await ensureAuth()
  const updated = await updateProfile(token.value, {
    displayName: profileForm.value.displayName,
    bio: profileForm.value.bio,
    interests: parsedInterests(),
    language: profileForm.value.language,
    ageConfirmed: profileForm.value.ageConfirmed
  })
  setProfile(updated)
}

async function setupPushNotifications(promptUser = false) {
  const publicKey = vapidPublicKey()
  if (!publicKey) {
    pushStatus.value = 'unconfigured'
    return
  }
  if (!('serviceWorker' in navigator) || !('PushManager' in window) || !('Notification' in window)) {
    pushStatus.value = 'unsupported'
    if (promptUser) errorText.value = iosPushHelpText()
    return
  }

  try {
    let permission = Notification.permission
    if (permission === 'default' && promptUser) {
      permission = await Notification.requestPermission()
    }
    if (permission === 'default') return
    if (permission !== 'granted') {
      pushStatus.value = 'blocked'
      if (promptUser) errorText.value = '通知权限未开启，请在浏览器网址列左侧设置里允许通知'
      return
    }

    await ensureAuth()
    const registration = await navigator.serviceWorker.register('/sw.js')
    const existing = await registration.pushManager.getSubscription()
    const subscription = existing ?? await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(publicKey)
    })
    const payload = subscription.toJSON()
    if (!payload.endpoint || !payload.keys?.auth || !payload.keys?.p256dh) return
    await savePushSubscription(token.value, {
      endpoint: payload.endpoint,
      keys: {
        auth: payload.keys.auth,
        p256dh: payload.keys.p256dh
      }
    })
    pushStatus.value = 'enabled'
    if (promptUser) {
      await sendPushTest(token.value)
    }
  } catch {
    if (promptUser && !errorText.value) errorText.value = '通知开启失败，请确认使用 HTTPS 并重新整理页面后再试'
  }
}

function iosPushHelpText() {
  if (!/iPad|iPhone|iPod/.test(navigator.userAgent)) return '当前浏览器不支持网页推播通知'
  return 'iPhone 需先用 Safari 分享按钮加入主画面，再从主画面打开后才能开启通知'
}

function urlBase64ToUint8Array(value: string) {
  const padding = '='.repeat((4 - value.length % 4) % 4)
  const base64 = (value + padding).replace(/-/g, '+').replace(/_/g, '/')
  const raw = window.atob(base64)
  const output = new Uint8Array(raw.length)
  for (let i = 0; i < raw.length; i += 1) {
    output[i] = raw.charCodeAt(i)
  }
  return output
}

async function startMatch() {
  loading.value = true
  errorText.value = ''
  try {
    await switchPage('video')
    resetCall()
    if (!profileForm.value.ageConfirmed) {
      await showAgeConfirmationRequired()
      return
    }
    await persistProfile()
    await setupPushNotifications(pushStatus.value === 'idle')
    await ensureAuth()
    await openMedia()
    await openSocketWithAuth()
    const result = await joinMatch(token.value, {
      mode: mode.value,
      region: selectedRegion.value,
      gender: genderPreference.value,
      language: profileForm.value.language,
      interests: parsedInterests()
    })
    await loadCommerceStatus()
    status.value = result.status === 'matched' ? 'matched' : 'waiting'
    if (result.status === 'matched' && result.roomId) {
      activeRoomId.value = result.roomId
      activePeerId.value = result.peerId || null
      peerCardHidden.value = false
      peerProfile.value = result.peerProfile || null
      if (mode.value === 'video') void captureAndUploadSnapshot(result.roomId, result.peerId)
    }
    if (result.status === 'matched' && result.initiator && result.peerId) {
      await createPeer(result.peerId)
    }
    void refreshStats()
  } catch (error) {
    errorText.value = toUserMessage(error)
    void loadCommerceStatus().catch(() => undefined)
  } finally {
    loading.value = false
  }
}

async function showAgeConfirmationRequired() {
  loading.value = false
  leaving.value = false
  status.value = 'idle'
  chatOpen.value = false
  await switchPage('profile')
  errorText.value = '请先确认已满 18 岁并保存资料'
  ageCheckAttention.value = true
  await nextTick()
  ageCheckRef.value?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  window.setTimeout(() => {
    ageCheckAttention.value = false
  }, 2600)
}

async function buyMembership() {
  if (paymentLoading.value || commerceStatus.value?.isMember) return
  paymentLoading.value = true
  errorText.value = ''
  try {
    await ensureAuth()
    const order = await createPaymentOrder(token.value)
    await confirmPaymentOrder(token.value, order.id)
    await loadCommerceStatus()
    errorText.value = '会员已开通，可无限匹配并优先排队'
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    paymentLoading.value = false
  }
}

async function leaveCall() {
  if (leaving.value || status.value === 'idle') return
  leaving.value = true
  errorText.value = ''
  try {
    if (token.value) await leaveMatch(token.value)
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    closeSocket()
    stopLocalMedia()
    resetCall()
    leaving.value = false
    void refreshStats()
  }
}

async function switchCamera() {
  if (switchingCamera.value || !localStream.value) return
  switchingCamera.value = true
  errorText.value = ''
  const nextFacing = cameraFacing.value === 'user' ? 'environment' : 'user'
  const currentFacing = cameraFacing.value
  const oldVideoTracks = localStream.value.getVideoTracks()

  try {
    const nextStream = await openCameraForSwitch(nextFacing, oldVideoTracks)
    const [nextVideoTrack] = nextStream.getVideoTracks()
    if (!nextVideoTrack) throw new Error('没有找到可用的摄像头')

    const currentStream = localStream.value
    const audioTracks = currentStream.getAudioTracks()
    localStream.value = new MediaStream([...audioTracks, nextVideoTrack])
    if (localVideo.value) localVideo.value.srcObject = localStream.value
    cameraFacing.value = nextFacing

    const sender = peer.value?.getSenders().find((item) => item.track?.kind === 'video')
    if (sender) {
      await sender.replaceTrack(nextVideoTrack)
      await setVideoSenderLimits(sender)
    }
    oldVideoTracks.forEach((track) => {
      if (track !== nextVideoTrack) track.stop()
    })
  } catch (error) {
    await restoreCamera(currentFacing)
    errorText.value = toUserMessage(error)
  } finally {
    switchingCamera.value = false
  }
}

async function openCameraForSwitch(facing: 'user' | 'environment', oldVideoTracks: MediaStreamTrack[]) {
  try {
    return await navigator.mediaDevices.getUserMedia({
      video: videoConstraints(facing),
      audio: false
    })
  } catch (error) {
    oldVideoTracks.forEach((track) => track.stop())
    try {
      return await navigator.mediaDevices.getUserMedia({
        video: videoConstraints(facing),
        audio: false
      })
    } catch {
      throw error
    }
  }
}

async function restoreCamera(facing: 'user' | 'environment') {
  if (!localStream.value) return
  try {
    const fallbackStream = await navigator.mediaDevices.getUserMedia({
      video: videoConstraints(facing),
      audio: false
    })
    const [fallbackVideoTrack] = fallbackStream.getVideoTracks()
    if (!fallbackVideoTrack) return

    const audioTracks = localStream.value.getAudioTracks()
    localStream.value = new MediaStream([...audioTracks, fallbackVideoTrack])
    if (localVideo.value) localVideo.value.srcObject = localStream.value

    const sender = peer.value?.getSenders().find((item) => item.track?.kind === 'video')
    if (sender) {
      await sender.replaceTrack(fallbackVideoTrack)
      await setVideoSenderLimits(sender)
    }
  } catch {
    // Keep the original switch error visible to the user.
  }
}

async function openMedia() {
  if (!navigator.mediaDevices?.getUserMedia) {
    throw new Error('当前浏览器不支持摄像头/麦克风访问，请使用 HTTPS 或 localhost 打开页面')
  }
  stopLocalMedia()
  localStream.value = await navigator.mediaDevices.getUserMedia({
    video: mode.value === 'video' ? videoConstraints(cameraFacing.value) : false,
    audio: {
      echoCancellation: true,
      noiseSuppression: true,
      autoGainControl: true
    }
  })
  if (localVideo.value) localVideo.value.srcObject = localStream.value
  await nextTick()
  ensurePreviewPosition()
}

function videoConstraints(facingMode: 'user' | 'environment') {
  return {
    width: { ideal: 480, max: 640 },
    height: { ideal: 640, max: 720 },
    frameRate: { ideal: 15, max: 20 },
    facingMode: { ideal: facingMode }
  }
}

async function openSocketWithAuth() {
  try {
    await openSocket()
  } catch {
    clearToken()
    await ensureAuth()
    await openSocket()
  }
}

function openSocket() {
  if (ws.value?.readyState === WebSocket.OPEN) return Promise.resolve()
  ws.value?.close()

  return new Promise<void>((resolve, reject) => {
    const socket = new WebSocket(wsURL(token.value))
    ws.value = socket

    const failTimer = window.setTimeout(() => {
      reject(new Error('连接信令服务超时，请确认后端服务可访问'))
      socket.close()
    }, 8000)

    socket.onopen = () => {
      window.clearTimeout(failTimer)
      startSocketHeartbeat(socket)
      resolve()
    }

    socket.onerror = () => {
      window.clearTimeout(failTimer)
      reject(new Error('连接信令服务失败，登录可能已过期，请重试'))
    }

    socket.onclose = () => {
      stopSocketHeartbeat()
      if (ws.value === socket) ws.value = null
      if (closingSocket.value) {
        closingSocket.value = false
        return
      }
      if (status.value === 'matched') resetCall('信令连接已断开，请重新匹配')
    }

    socket.onmessage = async (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'pong') return
        if (msg.type === 'matched') {
          status.value = 'matched'
          activeRoomId.value = msg.roomId || null
          activePeerId.value = msg.peerId || null
          peerCardHidden.value = false
          peerProfile.value = msg.peerProfile || null
          if (msg.roomId) void captureAndUploadSnapshot(msg.roomId, msg.peerId)
          if (msg.initiator && msg.peerId) await createPeer(msg.peerId)
          return
        }
        if (msg.type === 'offer' && msg.peerId) {
          await acceptOffer(msg.peerId, msg.data)
          return
        }
        if (msg.type === 'answer' && peer.value?.signalingState === 'have-local-offer') {
          await peer.value.setRemoteDescription(msg.data)
          await flushCandidates()
          return
        }
        if (msg.type === 'candidate') {
          await addRemoteCandidate(msg.data)
          return
        }
        if (msg.type === 'chat-message') {
          receiveChatMessage(msg.data)
          return
        }
        if (msg.type === 'peer-left') {
          resetCall('对方已离开，请重新匹配')
        }
      } catch (error) {
        errorText.value = toUserMessage(error)
      }
    }
  })
}

function closeSocket() {
  const socket = ws.value
  ws.value = null
  stopSocketHeartbeat()
  closingSocket.value = Boolean(socket)
  socket?.close()
}

function stopLocalMedia() {
  localStream.value?.getTracks().forEach((track) => track.stop())
  localStream.value = null
  if (localVideo.value) localVideo.value.srcObject = null
}

async function captureAndUploadSnapshot(roomId: string, peerId = '') {
  if (capturedSnapshotRooms.has(roomId)) return
  if (!token.value || !localStream.value?.getVideoTracks().length) return
  capturedSnapshotRooms.add(roomId)

  try {
    await waitForLocalVideoFrame()
    const video = localVideo.value
    if (!video?.videoWidth || !video.videoHeight) return

    const canvas = document.createElement('canvas')
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    const context = canvas.getContext('2d')
    if (!context) return
    context.drawImage(video, 0, 0, canvas.width, canvas.height)

    await uploadMatchSnapshot(token.value, {
      roomId,
      peerId,
      mode: 'video',
      image: canvas.toDataURL('image/jpeg', 0.82),
      width: canvas.width,
      height: canvas.height
    })
  } catch {
    capturedSnapshotRooms.delete(roomId)
  }
}

function waitForLocalVideoFrame() {
  const startedAt = performance.now()
  return new Promise<void>((resolve, reject) => {
    const check = () => {
      const video = localVideo.value
      if (video?.videoWidth && video.videoHeight) {
        resolve()
        return
      }
      if (performance.now() - startedAt > 1800) {
        reject(new Error('local video frame timeout'))
        return
      }
      window.requestAnimationFrame(check)
    }
    check()
  })
}

function previewBounds() {
  const stageRect = stage.value?.getBoundingClientRect()
  const previewRect = localPreview.value?.getBoundingClientRect()
  if (!stageRect || !previewRect) return null
  return {
    maxX: Math.max(0, stageRect.width - previewRect.width - 16),
    maxY: Math.max(0, stageRect.height - previewRect.height - 16)
  }
}

function clampPreviewPosition(x: number, y: number) {
  const bounds = previewBounds()
  if (!bounds) return { x, y }
  return {
    x: Math.min(Math.max(16, x), bounds.maxX),
    y: Math.min(Math.max(16, y), bounds.maxY)
  }
}

function ensurePreviewPosition() {
  const bounds = previewBounds()
  if (!bounds) return
  if (!previewPositioned.value) {
    previewPosition.value = { x: bounds.maxX, y: bounds.maxY }
    previewPositioned.value = true
    return
  }
  previewPosition.value = clampPreviewPosition(previewPosition.value.x, previewPosition.value.y)
}

function startPreviewDrag(event: PointerEvent) {
  if (!localPreview.value || !previewPositioned.value) ensurePreviewPosition()
  previewDrag.value = {
    pointerId: event.pointerId,
    startX: event.clientX,
    startY: event.clientY,
    originX: previewPosition.value.x,
    originY: previewPosition.value.y
  }
  localPreview.value?.setPointerCapture(event.pointerId)
  window.addEventListener('pointermove', dragPreview)
  window.addEventListener('pointerup', stopPreviewDrag)
  window.addEventListener('pointercancel', stopPreviewDrag)
}

function dragPreview(event: PointerEvent) {
  const drag = previewDrag.value
  if (!drag || event.pointerId !== drag.pointerId) return
  previewPosition.value = clampPreviewPosition(
    drag.originX + event.clientX - drag.startX,
    drag.originY + event.clientY - drag.startY
  )
}

function stopPreviewDrag(event: PointerEvent) {
  const drag = previewDrag.value
  if (!drag || event.pointerId !== drag.pointerId) return
  localPreview.value?.releasePointerCapture(event.pointerId)
  previewDrag.value = null
  window.removeEventListener('pointermove', dragPreview)
  window.removeEventListener('pointerup', stopPreviewDrag)
  window.removeEventListener('pointercancel', stopPreviewDrag)
}

function teardownPeer(clearSession = true) {
  clearPeerDisconnectTimer()
  const currentPeer = peer.value
  peer.value = null
  currentPeer?.close()
  activePeerId.value = null
  peerCardHidden.value = false
  pendingCandidates.value = []
  if (clearSession) {
    activeRoomId.value = null
    peerProfile.value = null
    chatOpen.value = false
    clearChatToast()
    chatDraft.value = ''
    chatMessages.value = []
  }
  if (remoteVideo.value) remoteVideo.value.srcObject = null
}

function clearPeerDisconnectTimer() {
  if (peerDisconnectTimer.value !== null) {
    window.clearTimeout(peerDisconnectTimer.value)
    peerDisconnectTimer.value = null
  }
}

function schedulePeerDisconnect(message: string) {
  if (peerDisconnectTimer.value !== null) return
  peerDisconnectTimer.value = window.setTimeout(() => {
    peerDisconnectTimer.value = null
    resetCall(message)
  }, 5000)
}

function resetCall(message = '') {
  teardownPeer()
  status.value = 'idle'
  if (message) errorText.value = message
}

async function reportPeer() {
  if (!activePeerId.value || safetyLoading.value) return
  safetyLoading.value = true
  errorText.value = ''
  try {
    await reportUser(token.value, activePeerId.value, 'user reported during match')
    reportedPeerId.value = activePeerId.value
    errorText.value = '已收到举报'
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    safetyLoading.value = false
  }
}

async function blockPeer() {
  if (!activePeerId.value || safetyLoading.value) return
  const confirmed = window.confirm('确定要拉黑并结束当前视讯吗？之后不会再匹配到这个用户。')
  if (!confirmed) return
  safetyLoading.value = true
  errorText.value = ''
  try {
    await blockUser(token.value, activePeerId.value)
    closeSocket()
    stopLocalMedia()
    resetCall('已拉黑对方，不会再匹配到此用户')
  } catch (error) {
    errorText.value = toUserMessage(error)
  } finally {
    safetyLoading.value = false
  }
}

async function addRemoteCandidate(candidate: RTCIceCandidateInit) {
  if (!peer.value?.remoteDescription) {
    pendingCandidates.value.push(candidate)
    return
  }
  await peer.value.addIceCandidate(candidate)
}

async function flushCandidates() {
  if (!peer.value?.remoteDescription) return
  for (const candidate of pendingCandidates.value) {
    await peer.value.addIceCandidate(candidate)
  }
  pendingCandidates.value = []
}

async function createPeer(peerId: string) {
  if (activePeerId.value === peerId && peer.value?.localDescription?.type === 'offer') return

  teardownPeer(false)
  activePeerId.value = peerId

  const pc = buildPeer(peerId)
  peer.value = pc
  await addLocalTracks(pc)
  if (!isCurrentPeer(pc)) return

  const offer = await pc.createOffer()
  if (!isCurrentPeer(pc)) return
  await pc.setLocalDescription(offer)
  if (!isCurrentPeer(pc)) return
  send({ type: 'offer', peerId, data: pc.localDescription })
}

async function acceptOffer(peerId: string, offer: RTCSessionDescriptionInit) {
  if (activePeerId.value === peerId && peer.value?.localDescription?.type === 'answer') return

  teardownPeer(false)
  activePeerId.value = peerId

  const pc = buildPeer(peerId)
  peer.value = pc
  await addLocalTracks(pc)
  if (!isCurrentPeer(pc)) return

  await pc.setRemoteDescription(offer)
  if (!isCurrentPeer(pc)) return
  await flushCandidates()
  if (!isCurrentPeer(pc)) return
  const answer = await pc.createAnswer()
  if (!isCurrentPeer(pc)) return
  await pc.setLocalDescription(answer)
  if (!isCurrentPeer(pc)) return
  send({ type: 'answer', peerId, data: pc.localDescription })
}

function isCurrentPeer(pc: RTCPeerConnection) {
  return peer.value === pc && pc.signalingState !== 'closed'
}

function buildPeer(peerId: string) {
  const pc = new RTCPeerConnection({
    iceServers: iceServers(),
    iceTransportPolicy: import.meta.env.VITE_FORCE_TURN === 'true' ? 'relay' : 'all'
  })
  pc.onicecandidate = (event) => {
    if (!isCurrentPeer(pc)) return
    if (event.candidate) send({ type: 'candidate', peerId, data: event.candidate })
  }
  pc.onconnectionstatechange = () => {
    if (!isCurrentPeer(pc)) return
    if (pc.connectionState === 'connected') {
      clearPeerDisconnectTimer()
      return
    }
    if (pc.connectionState === 'disconnected') {
      schedulePeerDisconnect('对方连接不稳定，请重新匹配')
      return
    }
    if (['failed', 'closed'].includes(pc.connectionState)) {
      resetCall('对方已断线，请重新匹配')
    }
  }
  pc.oniceconnectionstatechange = () => {
    if (!isCurrentPeer(pc)) return
    if (['connected', 'completed'].includes(pc.iceConnectionState)) {
      clearPeerDisconnectTimer()
      return
    }
    if (pc.iceConnectionState === 'disconnected') {
      schedulePeerDisconnect('对方连接不稳定，请重新匹配')
      return
    }
    if (['failed', 'closed'].includes(pc.iceConnectionState)) {
      resetCall('对方连接已中断，请重新匹配')
    }
  }
  pc.ontrack = (event) => {
    if (!isCurrentPeer(pc)) return
    const [stream] = event.streams
    if (remoteVideo.value) remoteVideo.value.srcObject = stream
    event.track.onended = () => {
      if (isCurrentPeer(pc)) resetCall('对方已离开，请重新匹配')
    }
    event.track.onmute = () => {
      if (!isCurrentPeer(pc)) return
      if (pc.connectionState !== 'connected') resetCall('对方媒体已中断，请重新匹配')
    }
  }
  return pc
}

async function addLocalTracks(pc: RTCPeerConnection) {
  const stream = localStream.value
  if (!stream) return
  for (const track of stream.getTracks()) {
    const sender = pc.addTrack(track, stream)
    if (track.kind !== 'video') continue
    await setVideoSenderLimits(sender)
  }
}

async function setVideoSenderLimits(sender: RTCRtpSender) {
  const params = sender.getParameters()
  params.encodings = params.encodings?.length ? params.encodings : [{}]
  params.encodings[0] = {
    ...params.encodings[0],
    maxBitrate: 420_000,
    maxFramerate: 20
  }
  try {
    await sender.setParameters(params)
  } catch {
    // Some browsers reject sender parameter changes before negotiation.
  }
}

function send(message: unknown) {
  if (ws.value?.readyState !== WebSocket.OPEN) {
    errorText.value = '信令连接已断开，请重新匹配'
    return
  }
  ws.value.send(JSON.stringify(message))
}

function sendChatMessage() {
  if (!canSendChat.value || !activePeerId.value) return
  const text = chatDraft.value.trim()
  const message: ChatMessage = {
    id: newMessageId(),
    sender: 'self',
    text,
    createdAt: new Date().toISOString()
  }
  chatDraft.value = ''
  chatMessages.value.push(message)
  scrollChatToBottom()
  send({
    type: 'chat-message',
    peerId: activePeerId.value,
    roomId: activeRoomId.value || undefined,
    data: {
      id: message.id,
      text: message.text,
      createdAt: message.createdAt
    }
  })
}

function receiveChatMessage(data: unknown) {
  if (!data || typeof data !== 'object') return
  const payload = data as { id?: unknown; text?: unknown; createdAt?: unknown }
  if (typeof payload.text !== 'string') return
  const text = payload.text.trim()
  if (!text) return
  chatMessages.value.push({
    id: typeof payload.id === 'string' ? payload.id : newMessageId(),
    sender: 'peer',
    text: truncateText(text, 500),
    createdAt: typeof payload.createdAt === 'string' ? payload.createdAt : new Date().toISOString()
  })
  if (!chatOpen.value) showChatToast(text)
  scrollChatToBottom()
}

function showChatToast(text: string) {
  chatToastText.value = truncateText(text, 42)
  if (chatToastText.value !== text) chatToastText.value += '...'
  if (chatToastTimer.value !== null) window.clearTimeout(chatToastTimer.value)
  chatToastTimer.value = window.setTimeout(clearChatToast, 3500)
}

function clearChatToast() {
  chatToastText.value = ''
  if (chatToastTimer.value !== null) {
    window.clearTimeout(chatToastTimer.value)
    chatToastTimer.value = null
  }
}

function scrollChatToBottom() {
  void nextTick(() => {
    const list = chatList.value
    if (list) list.scrollTop = list.scrollHeight
  })
}

function newMessageId() {
  if (crypto.randomUUID) return crypto.randomUUID()
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function truncateText(value: string, maxLength: number) {
  const segmenter = typeof Intl !== 'undefined' && 'Segmenter' in Intl
    ? new Intl.Segmenter(undefined, { granularity: 'grapheme' })
    : null
  const chars = segmenter
    ? Array.from(segmenter.segment(value), (item) => item.segment)
    : Array.from(value)
  return chars.length > maxLength ? chars.slice(0, maxLength).join('') : value
}

function toUserMessage(error: unknown) {
  if (error instanceof DOMException && error.name === 'NotAllowedError') {
    return '请允许摄像头和麦克风权限后再开始匹配'
  }
  if (error instanceof DOMException && error.name === 'NotFoundError') {
    return '没有找到可用的摄像头或麦克风，或当前设备没有另一个镜头'
  }
  if (error instanceof TypeError && error.message === 'Failed to fetch') {
    return '无法连接后端服务，请确认 API 服务已启动且允许当前页面域名'
  }
  if (error instanceof Error) return error.message
  return '操作失败，请稍后重试'
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit'
  }).format(new Date(value))
}

onBeforeUnmount(() => {
  stopStatsPolling()
  stopSocketHeartbeat()
  clearChatToast()
  window.removeEventListener('resize', ensurePreviewPosition)
  window.removeEventListener('pointermove', dragPreview)
  window.removeEventListener('pointerup', stopPreviewDrag)
  window.removeEventListener('pointercancel', stopPreviewDrag)
  closeSocket()
  teardownPeer()
  stopLocalMedia()
})

onMounted(() => {
  window.addEventListener('resize', ensurePreviewPosition)
  startStatsPolling()
  void loadProfile()
  void setupPushNotifications(false)
})
</script>
