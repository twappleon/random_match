# App Store + Apple IAP Release Checklist

## App Store Connect

1. Create the app with bundle ID `com.danawang.randommatch`.
2. Create an auto-renewable subscription group named `Match Pass`.
3. Create product ID `premium_monthly`.
4. Set price to USD 6.99 and fill subscription review metadata.
5. Create sandbox testers for purchase testing.
6. Create an In-App Purchase key and record:
   - Issuer ID
   - Key ID
   - `.p8` private key

## Backend Environment

Set these on the production server before rebuilding `backend`:

```env
APPLE_BUNDLE_ID=com.danawang.randommatch
APPLE_IAP_PRODUCT_ID=premium_monthly
APPLE_IAP_ISSUER_ID=
APPLE_IAP_KEY_ID=
APPLE_IAP_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
APPLE_IAP_ENVIRONMENT=Production
APPLE_IAP_ALLOW_UNVERIFIED=false
```

For sandbox testing, use:

```env
APPLE_IAP_ENVIRONMENT=Sandbox
```

Do not enable `APPLE_IAP_ALLOW_UNVERIFIED=true` in production.

## Deploy

```bash
cd ~/random_match
git pull
docker-compose -f deploy/docker-compose.prod.yml --env-file .env up -d --build --force-recreate backend h5
```

## iOS Build

Before archive/upload:

1. Add `apps/mobile/ios/Runner/GoogleService-Info.plist` from Firebase.
2. Confirm `PrivacyInfo.xcprivacy` is included in the Runner target.
3. Confirm App Store Connect privacy answers match the app:
   - User ID
   - User content
   - Purchase history
   - Push notification token / device identifiers if used
4. Archive with an App Store distribution profile.

## App Review Notes

Mention these moderation and safety features:

- Users must confirm they are 18+ before matching.
- Users can leave, report, and block during calls.
- Blocked users are excluded from future matching where possible.
- Match Pass is purchased through Apple IAP and unlocks unlimited matching, priority queue, and Gems.
