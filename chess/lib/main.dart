import 'package:chess/screens/home_screen.dart';
import 'package:flutter/material.dart';

void main() {
  runApp(
    const Myapp()
  );
}

class Myapp extends StatelessWidget {
  const Myapp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(debugShowCheckedModeBanner: false, home: HomeScreen());
  }
}
