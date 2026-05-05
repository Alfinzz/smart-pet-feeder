import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('basic material smoke test', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: Text('Smart Pet Monitoring'))),
    );

    expect(find.text('Smart Pet Monitoring'), findsOneWidget);
  });
}
