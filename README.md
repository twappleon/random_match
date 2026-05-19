# Random Match

线上随机视讯/语音交友项目脚手架。

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
docker compose -f deploy/docker-compose.yml up -d
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
- `POST /api/v1/match/join`: 加入随机匹配队列
- `POST /api/v1/match/leave`: 离开匹配队列
- `GET /api/v1/ws?token=...`: WebRTC signaling WebSocket

## 下一步

- 接入真实 Firebase 项目配置。
- 增加短信/Apple/Google 登录。
- 增加举报、拉黑、内容审核、年龄门槛和风控。
- 使用 TURN 服务提升 WebRTC NAT 穿透成功率。
