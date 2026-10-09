import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_analytics/firebase_analytics.dart';
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:google_mobile_ads/google_mobile_ads.dart';
import 'package:get/get.dart';

import 'core/app_translations.dart';
import 'controllers/match_controller.dart';
import 'pages/home_page.dart';

FirebaseAnalytics? analytics;

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await Firebase.initializeApp();
    analytics = FirebaseAnalytics.instance;
    await analytics?.logAppOpen();
  } catch (_) {
    // Native Firebase files are added later by the real Firebase project.
  }
  await MobileAds.instance.initialize();
  runApp(const RandomMatchApp());
}

class RandomMatchApp extends StatelessWidget {
  const RandomMatchApp({super.key, this.locale});

  final Locale? locale;

  @override
  Widget build(BuildContext context) {
    return GetMaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Random Match',
      translations: AppTranslations(),
      locale: locale ?? Get.deviceLocale,
      fallbackLocale: AppTranslations.fallbackLocale,
      supportedLocales: AppTranslations.supportedLocales,
      localizationsDelegates: GlobalMaterialLocalizations.delegates,
      navigatorObservers: [
        if (analytics != null) FirebaseAnalyticsObserver(analytics: analytics!),
      ],
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xff7c5cff),
          brightness: Brightness.dark,
        ),
        scaffoldBackgroundColor: const Color(0xff050608),
        snackBarTheme: SnackBarThemeData(
          backgroundColor: const Color(0xff181b29),
          contentTextStyle: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.w700,
          ),
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        ),
        useMaterial3: true,
      ),
      initialBinding: BindingsBuilder(() {
        if (!Get.isRegistered<MatchController>()) {
          Get.put(MatchController());
        }
      }),
      home: const HomePage(),
    );
  }
}
