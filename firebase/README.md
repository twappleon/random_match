# Firebase

把真实项目生成的配置通过文件或环境变量接入。

当前 Firebase 项目：

```text
Project ID: random-match-7370c
Web app: random-match-web
```

- Android: `apps/mobile/android/app/google-services.json`
- iOS: `apps/mobile/ios/Runner/GoogleService-Info.plist`
- Web: 复制 `apps/h5/.env.example` 到 `apps/h5/.env`，把 Firebase web config 填入 `VITE_FIREBASE_*`
- Server: 把 Firebase service account JSON 通过服务器文件或 secret 挂载到容器内，并设置：

```env
FIREBASE_PROJECT_ID=random-match-7370c
GOOGLE_APPLICATION_CREDENTIALS=/app/firebase-service-account.json
```

这些文件不要提交到 Git：

- `apps/mobile/android/app/google-services.json`
- `apps/mobile/ios/Runner/GoogleService-Info.plist`
- Firebase service account JSON

建议启用：

- Firebase Authentication
  - Email/Password: 已启用
  - Phone: 已启用
  - Google: 已启用
  - Apple: 已启用
- Firebase Cloud Messaging
- Firebase Analytics
- Firebase Crashlytics

H5 登录流程：

```text
Firebase Auth 登录
  -> H5 取得 Firebase ID token
  -> POST /api/v1/auth/firebase
  -> 后端 Firebase Admin SDK 验证 token
  -> 后端创建/更新 users.firebaseUid
  -> 后端返回本项目 JWT
```

注意：

- Phone Auth 需要把 H5 域名加入 Firebase Authentication Authorized domains。
- 当前 Authorized domains 已包含 `localhost`、`random-match-7370c.firebaseapp.com`、`random-match-7370c.web.app`。正式域名上线后也要加入。
- Apple 登录已在 Apple Developer 和 Firebase Console 完成 Sign in with Apple 配置：
  - Team ID: `PWNBD8U4DL`
  - App ID: `com.leon456.randommatch`
  - Service ID: `com.leon456.randommatch.web`
  - Key ID: `JZ7B4L4R36`
  - Redirect URL: `https://random-match-7370c.firebaseapp.com/__/auth/handler`
  - Private key local path: `/Users/liuleon/Documents/random_match_secrets/apple/AuthKey_JZ7B4L4R36.p8`
- 后端必须设置 `FIREBASE_PROJECT_ID` 和 `GOOGLE_APPLICATION_CREDENTIALS`，否则 `/api/v1/auth/firebase` 无法验证 ID token。
