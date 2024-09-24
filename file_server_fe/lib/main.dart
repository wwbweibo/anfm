import 'package:flutter/material.dart';
import 'package:file_server_fe/pages/file_explorer.dart';

void main() {
  runApp(const FileServerApp());
}

class FileServerApp extends StatelessWidget {
  const FileServerApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'File Server',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.white),
        useMaterial3: true,
      ),
      home: const FileServerHomePage(),
    );
  } 
}

class FileServerHomePage extends StatefulWidget {
  const FileServerHomePage({super.key});
  final String title = 'File Server';

  @override
  State<FileServerHomePage> createState() => _FileServerHomePageState();
}

class _FileServerHomePageState extends State<FileServerHomePage> {
  int _selectedIndex = 0;
  static const TextStyle optionStyle = TextStyle(fontSize: 30, fontWeight: FontWeight.bold);
  static const List<Widget> _widgetOptions = <Widget>[
    FileExplorer(),
    Text(
      'Index 1: Files',
      style: optionStyle,
    ),
    Text(
      'Index 2: Settings',
      style: optionStyle,
    ),
  ];

  void _onItemTapped(int index) {
    setState(() {
      _selectedIndex = index;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
        leading: Builder(
          builder: (context) {
            return IconButton(
              icon: const Icon(Icons.menu),
              onPressed: () {
                Scaffold.of(context).openDrawer();
              },
            );
          },
        ),
      ),
      body: Center(
        child: _widgetOptions.elementAt(_selectedIndex),
      ),
      drawer: Drawer(
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            const DrawerHeader(
              decoration: BoxDecoration(color: Colors.blue),
              child: Text('File Server')
            ),
            ListTile(
              title: const Text('文件浏览器'),
              selected: _selectedIndex == 0,
              onTap: () {
                  Navigator.pop(context);
              },
            ),
          ],
        ),
      ),
    );
  }
}
