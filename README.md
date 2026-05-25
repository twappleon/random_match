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
- `GET /api/v1/me`: 读取匿名身份资料
- `PUT /api/v1/me`: 更新昵称、简介、兴趣标签和年龄确认
- `POST /api/v1/match/join`: 加入随机视讯匹配队列
- `POST /api/v1/match/leave`: 离开匹配队列
- `POST /api/v1/match/snapshot`: 保存配对成功后的截图
- `POST /api/v1/users/{id}/report`: 举报用户
- `POST /api/v1/users/{id}/block`: 拉黑用户
- `POST /api/v1/push/subscription`: 保存浏览器推送订阅
- `GET /api/v1/ws?token=...`: WebRTC signaling WebSocket

## 匿名社交资料

H5 空闲状态会显示匿名身份卡，用户可填写昵称、简介、兴趣标签，并确认已满 18 岁。开始随机匹配时会自动保存资料；配对成功后双方会看到对方的匿名资料卡，可直接举报或拉黑。

拉黑后当前通话会结束，后端会记录拉黑关系并尽量跳过后续匹配中的该用户。

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
VAPID_SUBJECT=mailto:admin@danawang8899.com
```

更新并重建 H5 和后端：

```bash
cd ~/random_match
git pull
docker-compose -f deploy/docker-compose.prod.yml --env-file .env up -d --build --force-recreate backend h5
```

测试方式：

1. 用 `https://h5.danawang8899.com` 打开 H5。
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
