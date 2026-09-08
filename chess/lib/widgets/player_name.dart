import 'package:chess/services/player_name_service.dart';
import 'package:flutter/material.dart';

class PlayerName extends StatefulWidget {
  const PlayerName({super.key});

  @override
  State<PlayerName> createState() => _PlayerNameState();
}

class _PlayerNameState extends State<PlayerName> {
  bool _isEditing = false;
  String _playerName = '';

  late final TextEditingController _controller;
  late final FocusNode _focusNode;

  @override
  void initState() {
    super.initState();

    _controller = TextEditingController();
    _focusNode = FocusNode();

    _focusNode.addListener(_handleFocusChange);

    _loadName();
  }

  @override
  void dispose() {
    _focusNode.removeListener(_handleFocusChange);
    _focusNode.dispose();
    _controller.dispose();

    super.dispose();
  }

  Future<void> _loadName() async {
    final name = await PlayerService.getPlayerName();

    if (!mounted) return;

    setState(() {
      _playerName = name;
    });
  }

  void _handleFocusChange() {
    if (!_focusNode.hasFocus && _isEditing) {
      _saveName();
    }
  }

  Future<void> _saveName() async {
    final newName = _controller.text.trim();

    if (newName.isEmpty || newName.length < 3) {
      setState(() {
        _isEditing = false;
      });
      return;
    }

    await PlayerService.savePlayerName(newName);

    if (!mounted) return;

    setState(() {
      _playerName = newName;
      _isEditing = false;
    });
  }

  void _startEditing() {
    _controller.text = _playerName;

    setState(() {
      _isEditing = true;
    });

    _focusNode.requestFocus();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    if (_playerName.isEmpty) {
      return const SizedBox(
        height: 48,
        child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
      );
    }

    if (_isEditing) {
      return ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 240),
        child: TextField(
          controller: _controller,
          focusNode: _focusNode,
          autofocus: true,
          textAlign: TextAlign.center,
          style: theme.textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.w600,
          ),
          decoration: const InputDecoration(
            isDense: true,
            contentPadding: EdgeInsets.symmetric(vertical: 8),
          ),
          onSubmitted: (_) => _saveName(),
        ),
      );
    }

    return GestureDetector(
      onTap: _startEditing,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        decoration: BoxDecoration(
          color: colors.surface,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: colors.outline),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              _playerName,
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(width: 8),
            Icon(
              Icons.edit_outlined,
              size: 16,
              color: colors.onSurface.withValues(alpha: 0.6),
            ),
          ],
        ),
      ),
    );
  }
}
