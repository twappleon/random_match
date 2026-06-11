# Firebase

把真实项目生成的配置通过文件或环境变量接入：

- Android: `apps/mobile/android/app/google-services.json`
- iOS: `apps/mobile/ios/Runner/GoogleService-Info.plist`
- Web: 把 Firebase web config 填入 `apps/h5/src/firebase.ts`
- Server: 把 Firebase service account JSON 通过服务器文件或 secret 挂载到容器内，并设置：

```env
FIREBASE_PROJECT_ID=你的 Firebase project id
GOOGLE_APPLICATION_CREDENTIALS=/app/firebase-service-account.json
```

这些文件不要提交到 Git：

- `apps/mobile/android/app/google-services.json`
- `apps/mobile/ios/Runner/GoogleService-Info.plist`
- Firebase service account JSON

建议启用：

- Firebase Authentication
- Firebase Cloud Messaging
- Firebase Analytics
- Firebase Crashlytics
