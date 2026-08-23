# Random Match

线上随机视讯交友项目脚手架。

## 技术栈

- Backend: Go, MongoDB, Redis
- H5: Vue 3 + Vite
- App: Flutter
- Realtime: WebSocket signaling + WebRTC client
- Push/Auth/Analytics: Firebase 占位集成

## 项目结构

```text
apps/
  h5/          Vue3 H5
  mobile/      Flutter App
backend/       Go API 和实时匹配服务
deploy/        Docker Compose 本地依赖
firebase/      Firebase 配置占位
```

## 本地开发

1. 安装 Go、Node.js/npm、Flutter、Docker。
2. 复制环境变量：

```bash
cp .env.example .env
```

3. 启动 MongoDB 和 Redis：

```bash
docker-compose -f deploy/docker-compose.yml up -d
```

4. 启动 Go 后端：

```bash
cd backend
go mod tidy
go run ./cmd/api
```

5. 启动 H5：

```bash
cd apps/h5
npm install
npm run dev
```

6. 启动 Flutter：

```bash
cd apps/mobile
flutter create .
flutter pub get
flutter run
```

## 核心接口

- `GET /health`: 健康检查
- `POST /api/v1/auth/anonymous`: 匿名登录
- `POST /api/v1/auth/firebase`: 使用 Firebase ID token 登录/注册，并换取后端 JWT
- `GET /api/v1/me`: 读取匿名身份资料
- `PUT /api/v1/me`: 更新昵称、简介、兴趣标签和年龄确认
- `PUT /api/v1/me/location`: 更新当前用户定位；其他用户只会看到计算后的距离
- `POST /api/v1/match/join`: 加入随机视讯匹配队列
- `POST /api/v1/match/leave`: 离开匹配队列
- `POST /api/v1/match/snapshot`: 保存配对成功后的截图
- `POST /api/v1/users/{id}/report`: 举报用户
- `POST /api/v1/users/{id}/block`: 拉黑用户
- `GET /api/v1/commerce/status`: 读取会员状态和每日免费匹配额度
- `POST /api/v1/commerce/orders`: 创建会员订单
- `POST /api/v1/commerce/orders/{id}/confirm`: 确认会员订单并开通会员
- `POST /api/v1/push/subscription`: 保存浏览器推送订阅
- `POST /api/v1/push/device-token`: 保存 Native App FCM 推送 token
- `GET /api/v1/ws?token=...`: WebRTC signaling WebSocket

## 商业化能力

当前版本内置基础商业化闭环：

- 免费用户每天可发起 10 次随机匹配，额度按 UTC 日期重置。
- 会员用户不限制每日匹配次数。
- 会员用户进入优先队列，撮合时优先消化会员等待队列，再消化普通队列。
- H5 空闲页展示今日剩余额度和会员开通入口。
- 订单接口目前是模拟支付确认：`/commerce/orders/{id}/confirm` 会直接把订单标记为 paid，并给用户开通一个月会员。生产收款前应替换为 Stripe、Apple/Google IAP 或本地支付网关的服务端回调确认。

## 匿名社交资料

H5 空闲状态会显示匿名身份卡，用户可填写昵称、简介、兴趣标签，并确认已满 18 岁。开始随机匹配时会自动保存资料；配对成功后双方会看到对方的匿名资料卡，可直接举报或拉黑。

拉黑后当前通话会结束，后端会记录拉黑关系并尽量跳过后续匹配中的该用户。

## Firebase Auth 登录

H5 登录页已接入 Firebase Authentication，不再使用 mock 登录流程。当前支持：

- Email / Password 登录
- Email / Password 注册
- 手机验证码登录，使用 Firebase Phone Auth 和 invisible reCAPTCHA
- Google popup 登录
- Apple popup 登录

### 1. Firebase Console 配置

1. 在 Firebase Console 创建或打开项目。
2. 到 Authentication > Sign-in method 启用以下 provider：
   - Email/Password
   - Phone
   - Google
   - Apple
3. Phone Auth 需要把测试/生产域名加入 Firebase Authentication 的 Authorized domains。开发环境通常需要：

```text
localhost
127.0.0.1
```

4. Google 登录需要配置 OAuth consent screen。
5. Apple 登录需要在 Apple Developer 后台配置 Sign in with Apple，并把 Service ID / redirect domain 按 Firebase Console 指示填好。

当前 Firebase 项目为 `random-match-7370c`，Web app 名称为 `random-match-web`。Firebase Console 已启用：

- Email/Password
- Phone
- Google
- Apple

Authorized domains 已包含：

```text
localhost
random-match-7370c.firebaseapp.com
random-match-7370c.web.app
```

Apple provider 已使用以下 Apple Developer 设置启用：

- Team ID: `PWNBD8U4DL`
- App ID: `com.leon456.randommatch`
- Service ID: `com.leon456.randommatch.web`
- Key ID: `JZ7B4L4R36`
- Firebase redirect URL:

```text
https://random-match-7370c.firebaseapp.com/__/auth/handler
```

Apple private key 已保存在本机专案外部路径，不要提交到 Git：

```text
/Users/liuleon/Documents/random_match_secrets/apple/AuthKey_JZ7B4L4R36.p8
```

### 2. H5 环境变量

H5 通过 Vite 环境变量读取 Firebase Web config。先复制示例：

```bash
cd apps/h5
cp .env.example .env
```

然后把 Firebase Console > Project settings > General > Web app config 填入：

```env
VITE_FIREBASE_API_KEY=AIzaSyAwXlt3ZTo7GTSPPEEpYz0rZfU4hWgSVEo
VITE_FIREBASE_AUTH_DOMAIN=random-match-7370c.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=random-match-7370c
VITE_FIREBASE_STORAGE_BUCKET=random-match-7370c.firebasestorage.app
VITE_FIREBASE_MESSAGING_SENDER_ID=79262345196
VITE_FIREBASE_APP_ID=1:79262345196:web:2a4161e46668f0882642c9
VITE_FIREBASE_MEASUREMENT_ID=G-1W9LWYBGCT
```

`VITE_API_BASE` 可不填；开发和生产默认使用当前 origin。若 H5 和 API 不同域，才设置：

```env
VITE_API_BASE=https://api.example.com
```

生产服务器使用 `deploy/docker-compose.prod.yml` 时，也必须把同一组 `VITE_FIREBASE_*` 写进根目录 `.env`。这些变量会在 Docker build 阶段传入 Vite；如果只在容器启动后才设置，前端静态档已经打包完成，页面仍会显示「Firebase 尚未设置」。

更新服务器 `.env` 后，至少重建 H5：

```bash
docker-compose -f deploy/docker-compose.prod.yml --env-file .env build --no-cache h5
docker-compose -f deploy/docker-compose.prod.yml --env-file .env up -d --force-recreate h5
```

### 3. 后端 Firebase Admin 配置

H5 使用 Firebase 登录成功后，会拿 Firebase ID token 调用：

```http
POST /api/v1/auth/firebase
```

后端会用 Firebase Admin SDK 验证 ID token，按 `firebaseUid` 创建或更新本地用户，然后签发本项目自己的 JWT。后续 `/api/v1/me`、匹配、聊天、定位、会员等 API 仍使用这个后端 JWT。

生产服务器不要提交 service account JSON 到 Git。建议放到服务器安全路径，例如：

```text
/opt/random_match/secrets/firebase-service-account.json
```

后端环境变量：

```env
FIREBASE_PROJECT_ID=random-match-7370c
GOOGLE_APPLICATION_CREDENTIALS=/app/firebase-service-account.json
```

如果使用 `deploy/docker-compose.prod.yml`，可以用 override 文件挂载：

```yaml
# deploy/docker-compose.firebase.yml
services:
  backend:
    environment:
      FIREBASE_PROJECT_ID: ${FIREBASE_PROJECT_ID}
      GOOGLE_APPLICATION_CREDENTIALS: /app/firebase-service-account.json
    volumes:
      - /opt/random_match/secrets/firebase-service-account.json:/app/firebase-service-account.json:ro
```

启动或重建：

```bash
docker-compose \
  -f deploy/docker-compose.prod.yml \
  -f deploy/docker-compose.firebase.yml \
  --env-file .env \
  up -d --build --force-recreate backend h5
```

### 4. 登录流程

```text
H5 Firebase Auth 登录
  -> firebaseUser.getIdToken()
  -> POST /api/v1/auth/firebase
  -> 后端 VerifyIDToken
  -> users.firebaseUid upsert
  -> 后端 JWT
  -> H5 localStorage token
```

注意事项：

- 未设置 `VITE_FIREBASE_*` 时，H5 会提示 Firebase 尚未设置。
- 未设置后端 `FIREBASE_PROJECT_ID` 或 Admin credentials 时，`/api/v1/auth/firebase` 会失败。
- 手机验证码必须使用 Firebase 认可的授权域名；生产 HTTPS 域名上线后也要加入 Authorized domains。
- Apple 登录已完成 Firebase Console provider 配置；若重新生成 Apple key，需要同步更新 Firebase Apple provider。

## 离线上线通知

H5 会在浏览器允许通知后注册 Web Push 订阅。之后只要有人建立 WebSocket 上线，后端会推送一则「有人上线了」通知给目前未在线、且曾经授权通知的用户。

注意事项：

- 浏览器 Push Notification 需要 HTTPS，`localhost` 例外；直接用 `http://64.177.113.148` 通常不会生效。
- iPhone 普通 Safari 页面不支持 Web Push；需要 Safari 打开网站后点分享按钮，选择「加入主画面」，再从主画面图标打开，才能开启通知。
- 用户必须曾经打开过 H5 并允许通知，后端才有订阅可推送。
- 后端对每个离线接收者有 10 分钟冷却，避免频繁刷新造成通知轰炸。

在服务器产生 VAPID 密钥：

```bash
docker run --rm node:22-alpine sh -c "npm -g install web-push >/dev/null && web-push generate-vapid-keys"
```

执行后会输出 `Public Key` 和 `Private Key`，把结果写入服务器 `.env`：

```env
VAPID_PUBLIC_KEY=上面输出的PublicKey
VAPID_PRIVATE_KEY=上面输出的PrivateKey
VAPID_SUBJECT=mailto:admin@random-match.online
```

更新并重建 H5 和后端：

```bash
cd ~/random_match
git pull
docker-compose -f deploy/docker-compose.prod.yml --env-file .env up -d --build --force-recreate backend h5
```

测试方式：

1. 用 `https://h5.random-match.online` 打开 H5。
2. iPhone 先用 Safari 分享按钮「加入主画面」，再从主画面打开；电脑 Chrome/Android Chrome 可直接打开网页测试。
3. 点击「随机匹配」。
4. 浏览器跳出通知权限时选择允许；如果也跳出摄像头/麦克风权限，也选择允许。
5. 如果通知订阅成功，后端会立刻发出一则「服务器推送测试」通知。
6. 关闭这个浏览器页面，再用另一台设备或另一个浏览器打开 H5 上线，离线设备应收到「有人上线了」通知。

排查订阅是否写入后端：

```bash
docker exec random-match-mongo mongosh random_match --quiet --eval 'db.push_subscriptions.find({}, {userId:1, endpoint:1, updatedAt:1}).pretty()'
```

排查后端是否尝试发送通知：

```bash
docker-compose -f deploy/docker-compose.prod.yml --env-file .env logs backend | grep push
```

## Native App 推送通知

Flutter App 已接入 Firebase Cloud Messaging token 注册流程。App 启动并完成匿名登录后，会请求通知权限、取得 FCM token，并写入后端 `push_device_tokens` 集合；后端在用户离线时会通过 Firebase Admin SDK 发送 App Push。

### 1. Firebase Console 配置

1. 在 Firebase Console 创建或打开项目。
2. 新增 Android app，Package name 填：

```text
com.leon456.randommatch
```

3. 下载 `google-services.json`，放到：

```text
apps/mobile/android/app/google-services.json
```

4. 新增 iOS app，Bundle ID 填：

```text
com.leon456.randommatch
```

5. 下载 `GoogleService-Info.plist`，放到：

```text
apps/mobile/ios/Runner/GoogleService-Info.plist
```

6. 在 Firebase Console > Project settings > Cloud Messaging，为 iOS app 上传 APNs Auth Key，并填写 Apple Team ID 和 Key ID。

### 2. App 本地打包配置

`google-services.json` 和 `GoogleService-Info.plist` 是每个 Firebase 项目的配置文件，不提交到 Git。拉代码后需要把文件放到上述路径，再执行：

```bash
cd apps/mobile
flutter pub get
flutter build apk --release
PATH="$HOME/.gem/ruby/2.6.0/bin:$PATH" flutter build ios --release --no-codesign
```

如果要产出 signed IPA，还需要在 Xcode 的 `Runner` target 里选择 Apple Developer Team，并确保 `Push Notifications` 能力可用。

### 3. 后端服务账号配置

后端通过 Firebase Admin SDK 发送 FCM。不要把 service account JSON 提交到 Git，生产服务器上放到安全路径，例如：

```text
/opt/random_match/secrets/firebase-service-account.json
```

服务器 `.env` 需要设置：

```env
FIREBASE_PROJECT_ID=你的 Firebase project id
GOOGLE_APPLICATION_CREDENTIALS=/app/firebase-service-account.json
```

如果使用 `deploy/docker-compose.prod.yml`，可以用一个本地 override 文件挂载 service account：

```yaml
# deploy/docker-compose.firebase.yml
services:
  backend:
    environment:
      FIREBASE_PROJECT_ID: ${FIREBASE_PROJECT_ID}
      GOOGLE_APPLICATION_CREDENTIALS: /app/firebase-service-account.json
    volumes:
      - /opt/random_match/secrets/firebase-service-account.json:/app/firebase-service-account.json:ro
```

启动或重建后端时同时带上 override：

```bash
docker-compose \
  -f deploy/docker-compose.prod.yml \
  -f deploy/docker-compose.firebase.yml \
  --env-file .env \
  up -d --build --force-recreate backend
```

### 4. 文件和环境变量清单

- Android Firebase app package name: `com.leon456.randommatch`
- iOS Bundle ID: `com.leon456.randommatch`
- Android 客户端文件：`apps/mobile/android/app/google-services.json`
- iOS 客户端文件：`apps/mobile/ios/Runner/GoogleService-Info.plist`
- 后端 service account 文件：通过服务器文件或 secret 挂载到容器内
- 后端环境变量：

```env
FIREBASE_PROJECT_ID=你的 Firebase project id
GOOGLE_APPLICATION_CREDENTIALS=/app/firebase-service-account.json
```

注意：`google-services.json`、`GoogleService-Info.plist` 和 service account JSON 不提交到 Git。生产部署时通过服务器文件或 secret 挂载。

### 5. 验证和排查

排查 App token 是否写入：

```bash
docker exec random-match-mongo mongosh random_match --quiet --eval 'db.push_device_tokens.find({}, {userId:1, platform:1, updatedAt:1}).pretty()'
```

测试推送：

```bash
curl -X POST https://你的后端域名/api/v1/push/test \
  -H "Authorization: Bearer 用户JWT"
```

查看后端 FCM 日志：

```bash
docker-compose \
  -f deploy/docker-compose.prod.yml \
  -f deploy/docker-compose.firebase.yml \
  --env-file .env \
  logs backend | grep "push native"
```

## 线上截图文件

配对成功后，H5 会上传本地摄像头截图到后端容器的 `/app/snapshots`，生产环境通过 Docker volume 持久化。

查看截图文件：

```bash
docker exec random-match-api find /app/snapshots -type f | sort
```

复制到服务器当前目录查看：

```bash
docker cp random-match-api:/app/snapshots ./snapshots
```

## WebRTC 延迟排查

如果两端已经配对成功，但视频严重 lag，优先检查 TURN 和视频码率。

常见原因：

- TURN 只开放了 `3478`，没有开放 coturn relay 媒体端口。
- `VITE_FORCE_TURN=true` 会强制所有媒体走 TURN，延迟通常高于 P2P。
- 摄像头采集分辨率或帧率过高，会增加上传带宽和 CPU 压力。

更新并重建 H5 和 TURN：

```bash
cd ~/random_match
git pull
docker-compose -f deploy/docker-compose.prod.yml --env-file .env up -d --build --force-recreate h5 turn
```

服务器防火墙放行 TURN 端口：

```bash
ufw allow 3478/tcp
ufw allow 3478/udp
ufw allow 49160:49200/tcp
ufw allow 49160:49200/udp
ufw reload
```

Vultr Firewall 也要放行：

```text
TCP 3478
UDP 3478
TCP 49160-49200
UDP 49160-49200
```

确认 TURN 稳定后，建议把 `.env` 里的：

```env
VITE_FORCE_TURN=true
```

改成：

```env
VITE_FORCE_TURN=false
```

这样能直连 P2P 时会优先 P2P，不能直连才走 TURN，通常延迟更低。改完后重建 H5：

```bash
docker-compose -f deploy/docker-compose.prod.yml --env-file .env up -d --build --force-recreate h5
```

## 下一步

- 接入真实 Firebase 项目配置。
- 增加短信/Apple/Google 登录。
- 增加举报、拉黑、内容审核、年龄门槛和风控。
- 使用 TURN 服务提升 WebRTC NAT 穿透成功率。
