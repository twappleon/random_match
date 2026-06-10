import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:get/get.dart';

import 'controllers/match_controller.dart';
import 'pages/home_page.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  try {
    await Firebase.initializeApp();
  } catch (_) {
    // Native Firebase files are added later by the real Firebase project.
  }
  runApp(const RandomMatchApp());
}

class RandomMatchApp extends StatelessWidget {
  const RandomMatchApp({super.key});

  @override
  Widget build(BuildContext context) {
    return GetMaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Random Match',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xff2fd276),
          brightness: Brightness.dark,
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
