# iOS App Store Release

## App

- App name: Random Match Online
- Bundle ID: com.leon456.randommatch
- App Store Connect Apple ID: 6820517190
- Team ID: PWNBD8U4DL
- Version: 1.0.0
- Build: 7

## Local validation

```sh
cd apps/mobile
flutter analyze
flutter build ipa --release --no-codesign
```

## Apple account setup

- App ID: `XC com leon456 randommatch` / `com.leon456.randommatch`
- Distribution certificate: `Apple Distribution: Hung LIANG Liu (PWNBD8U4DL)`
- Provisioning profile: `Random Match App Store`
- Profile UUID: `b7a9bd4f-2daa-44ba-9012-2582b72374c0`
- Private signing keychain: `~/.random_match_signing/random-match.keychain-db`

## Export

Flutter's automatic App Store export requires an Xcode account. This machine exports with the installed manual App Store profile instead:

```sh
cd apps/mobile
security unlock-keychain -p randommatch ~/.random_match_signing/random-match.keychain-db
security set-keychain-settings -lut 21600 ~/.random_match_signing/random-match.keychain-db
xcodebuild -exportArchive \
  -archivePath build/ios/archive/Runner.xcarchive \
  -exportPath build/ios/ipa-manual \
  -exportOptionsPlist ios/ExportOptions.plist
```

If macOS asks for the `login` keychain password for `random_match_distribution`, cancel it and rerun the export after unlocking the dedicated keychain above. The dedicated keychain is first in the search list and is configured for `codesign` access.

## Upload

```sh
xcrun altool --upload-app --type ios \
  --file "build/ios/ipa-manual/Random Match.ipa" \
  --apple-id 6820517190 \
  --apiKey V3B253TNT2 \
  --apiIssuer eda04043-d472-4013-b248-92dfc5667e30 \
  --show-progress
```

Last successful upload:

- Delivery UUID: `b58d6dfd-82dd-4e22-9a3a-354b0a25ef07`
- IPA: `build/ios/ipa/Random Match.ipa`
- Result: uploaded successfully and validated by App Store Connect
- Build status: `WAITING_FOR_REVIEW`, import status `VALID`, `APP_STORE_ELIGIBLE`
- Export compliance: selected "no restricted encryption" in App Store Connect
- Deployment target has been raised to iOS 15.0 for App Store Connect compatibility.
- Firebase Analytics is included. `GoogleService-Info.plist` is bundled in `Runner.app`, and `IS_ANALYTICS_ENABLED` is `true`.

## App Store Connect Record

- Name: Random Match Online
- Bundle ID: com.leon456.randommatch
- SKU: random-match-online-ios
- Version: 1.0.0

Privacy, age rating, screenshots, review notes, and contact details have been completed in App Store Connect.

Current status:

- Build `7` / `1.0.0` is attached to iOS version `1.0`.
- Build `7` uses iOS minimum deployment target `15.0`.
- Version `1.0` was resubmitted to App Review on 2026-10-09 and is `WAITING_FOR_REVIEW`.
