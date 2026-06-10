import 'package:flutter_test/flutter_test.dart';

import 'package:random_match_mobile/main.dart';

void main() {
  testWidgets('renders match controls', (WidgetTester tester) async {
    await tester.pumpWidget(const RandomMatchApp());

    expect(find.text('视讯'), findsOneWidget);
    expect(find.text('资料'), findsOneWidget);
    expect(find.text('会员'), findsOneWidget);
    expect(find.text('随机匹配'), findsOneWidget);
  });
}
