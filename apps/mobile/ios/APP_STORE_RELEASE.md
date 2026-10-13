# iOS App Store Release

## App

- App name: Interest Match Online
- Bundle ID: com.leon456.randommatch
- App Store Connect Apple ID: 6820517190
- Team ID: PWNBD8U4DL
- Version: 1.0.0
- Build: 11

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
  --file "build/ios/ipa-v11/Interest Match.ipa" \
  --apple-id 6820517190 \
  --apiKey V3B253TNT2 \
  --apiIssuer eda04043-d472-4013-b248-92dfc5667e30 \
  --show-progress
```

Last successful upload:

- Delivery UUID: `37d50bc4-1961-4def-b4dc-1c249cae6565`
- IPA: `build/ios/ipa-v11/Interest Match.ipa`
- Result: uploaded successfully and validated by App Store Connect
- Build status: uploaded, processed, and `VALID` in App Store Connect
- Export compliance: selected "no restricted encryption" in App Store Connect
- Deployment target has been raised to iOS 15.0 for App Store Connect compatibility.
- Firebase Analytics is included. `GoogleService-Info.plist` is bundled in `Runner.app`, and `IS_ANALYTICS_ENABLED` is `true`.
- Rejection remediation: upgraded WebRTC dependency and verified the archive no longer contains `RPSystemBroadcastPickerView` or `buttonPressed:` references.
- UGC safety remediation: added 18+ agreement copy, zero-tolerance safety policy, content filtering, hide/report/block controls, 24-hour report handling copy, and in-app contact information.
- Guideline 1.2 repositioning: renamed visible app to `Interest Match`, removed random/anonymous positioning from first-run UI, screenshots, H5 pages, and review-facing copy. The app now emphasizes interest profiles, language exchange, and mutual consent before 1v1 voice/video.

## App Store Connect Record

- Name: Interest Match Online
- Bundle ID: com.leon456.randommatch
- SKU: random-match-online-ios
- Version: 1.0.0

Privacy, age rating, screenshots, review notes, and contact details have been completed in App Store Connect.

Current status:

- Build `11` / `1.0.0` was uploaded to App Store Connect on 2026-10-13 and is `VALID`.
- Build `11` uses iOS minimum deployment target `15.0`.
- Build `10` is currently selected in the review submission and must be replaced with build `11` after processing.
- The App Store Connect API key can read/upload, but Apple rejects changing the selected build or deleting submissions with `The API key in use does not allow this request`.

## Review-facing copy for Guideline 1.2

- Promotional text: `Connect with people through shared interests, language practice, and profile-based 1v1 invitations.`
- Description: `Interest Match Online helps adults meet new people through interest profiles, language preferences, and mutual 1v1 invitations. Users can browse profiles, follow people, send a short message, and start voice or video only after choosing compatible interests and confirming the connection. The app is not a Chatroulette-style or anonymous random chat service. Safety controls include age confirmation, profile reporting, blocking, immediate exit controls, and review of reported content.`
- Keywords: `interest social,language exchange,profile match,friend discovery,voice chat,video call,safety`
- Review notes: `This build repositions the app away from random or anonymous chat. The first screen and screenshots now emphasize interest profiles, language exchange, and mutual consent before a 1v1 voice or video connection. Users can browse profiles, follow/message first, report, block, and leave immediately. The app does not encourage Chatroulette-style anonymous random chat.`
