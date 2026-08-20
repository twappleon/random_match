# Random Match Mobile

当前目录包含 Flutter App 的业务入口。因为本机没有 `flutter` 命令，平台目录需要在安装 Flutter 后生成：

```bash
cd apps/mobile
flutter create .
flutter pub get
flutter run --dart-define=API_BASE=https://random-match.online
flutter build apk --release --dart-define=API_BASE=https://random-match.online
```

接入 Firebase 后放入原生配置文件：

- Android: `android/app/google-services.json`
- iOS: `ios/Runner/GoogleService-Info.plist`

移动端使用 `flutter_webrtc`，上线前需要补充 Android/iOS 摄像头、麦克风、网络权限配置，并建议配置 TURN 服务。
