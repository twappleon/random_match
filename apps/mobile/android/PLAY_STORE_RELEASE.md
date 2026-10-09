# Android Play Store Release

## App

- App name: Random Match Online
- Package name: `com.leon456.randommatch`
- Play Console developer account: `开心工作者`
- Play Console app ID: `4973277636739938108`
- Default language: Traditional Chinese (`zh-TW`)
- App type: App
- Pricing: Free

## Build

- Version name: `1.0.0`
- Version code: `7`
- Release AAB: `/Users/liuleon/Documents/random_match/apps/mobile/build/app/outputs/bundle/release/app-release.aab`
- Verification APK: `/Users/liuleon/Documents/random_match/apps/mobile/build/app/outputs/flutter-apk/app-release.apk`

Build command:

```sh
cd /Users/liuleon/Documents/random_match/apps/mobile
flutter build appbundle --release --build-name 1.0.0 --build-number 7
```

## Signing

- Keystore: `/Users/liuleon/.random_match_signing/random-match-android-upload.jks`
- Alias: `random-match-upload`
- Local Gradle signing config: `/Users/liuleon/Documents/random_match/apps/mobile/android/key.properties`

Do not commit the keystore or `key.properties`. They are intentionally ignored.

Upload certificate fingerprints:

- SHA-1: `E9:D2:CA:3C:29:30:EB:14:B9:8D:BE:44:3E:7F:CC:32:0F:4A:E3:2A`
- SHA-256: `83:80:0A:C9:40:DA:D7:EA:AA:E4:01:66:EF:4C:71:E1:2A:13:D7:13:E1:C0:2B:4E:42:E8:52:D4:DC:6F:1D:75`

## Verified Locally

- `flutter analyze`: passed.
- Release APK manifest:
  - `package=com.leon456.randommatch`
  - `versionCode=7`
  - `versionName=1.0.0`
  - `minSdk=24`
  - `targetSdk=36`
  - `compileSdk=36`
- Release APK signature: valid.
- Firebase Analytics is present in the AAB (`firebase-analytics.properties`, version `22.5.0`).

## Play Console Status

- App record was created in Safari under `开心工作者`.
- Internal testing release `1.0.0 (7)` was uploaded and published on 2026-10-09.
- Play Console shows build `7` as available to internal testers and not yet reviewed.
- Production release is blocked by Google Play's current new-app requirement:
  - publish a closed test release,
  - have at least 12 testers opt in,
  - run the closed test for at least 14 days,
  - then request production access.

## Store Listing Inputs

- Privacy policy: `https://h5.random-match.online/privacy.html`
- Support email: `twyhleon@gmail.com`
- Category recommendation: Social
- Ads declaration: Yes, because the app includes AdMob.
- Target age recommendation: 18+ because random video/audio chat is adult-oriented.

## Next Play Console Steps

1. Add tester email list.
2. Complete app content forms:
   - Privacy policy
   - Ads
   - App access
   - Target audience and content
   - Content rating
   - Data safety
3. Fill main store listing with screenshots, icon, feature graphic, short description, and full description.
4. Start closed testing and collect at least 12 opted-in testers for 14 days.
5. Apply for production access after the closed testing requirement is satisfied.
