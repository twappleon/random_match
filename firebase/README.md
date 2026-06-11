# Firebase

把真实项目生成的配置放在这里：

- Android: `apps/mobile/android/app/google-services.json`
- iOS: `apps/mobile/ios/Runner/GoogleService-Info.plist`
- Web: 把 Firebase web config 填入 `apps/h5/src/firebase.ts`
- Server: 把 Firebase service account JSON 通过服务器 secret 挂载，并设置 `GOOGLE_APPLICATION_CREDENTIALS` 与 `FIREBASE_PROJECT_ID`

建议启用：

- Firebase Authentication
- Firebase Cloud Messaging
- Firebase Analytics
- Firebase Crashlytics
