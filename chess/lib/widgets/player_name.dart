import 'package:chess/services/player_name_service.dart';
import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

class PlayerName extends StatefulWidget {
  const PlayerName({super.key});

  @override
  State<PlayerName> createState() => _PlayerNameState();
}

class _PlayerNameState extends State<PlayerName> {
  bool _isEditing = false;
  String _playerName = '';

  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();

    _controller = TextEditingController();

    _loadName();
  }

  @override
  void dispose() {
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

  Future<void> _saveName() async {
    final newName = _controller.text.trim();

    if (newName.isEmpty || newName.length < 3) {
      setState(() => _isEditing = false);
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
  }

  @override
  Widget build(BuildContext context) {
    if (_playerName.isEmpty) {
      return const SizedBox(
        height: 48,
        child: Center(
          child: CircularProgressIndicator(
            color: AppColors.primary,
            strokeWidth: 2,
          ),
        ),
      );
    }

    if (_isEditing) {
      return Container(
        constraints: const BoxConstraints(maxWidth: 240),
        padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.sm,
        ),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: AppColors.primary,
            width: 1.5,
          ),
        ),
        child: TextField(
          controller: _controller,
          autofocus: true,
          textAlign: TextAlign.center,
          style: const TextStyle(
            color: AppColors.text,
            fontSize: 18,
            fontWeight: FontWeight.w600,
          ),
          decoration: const InputDecoration(
            border: InputBorder.none,
            isDense: true,
            contentPadding: EdgeInsets.symmetric(
              vertical: AppSpacing.sm,
            ),
          ),
          onSubmitted: (_) => _saveName(),
        ),
      );
    }

    return GestureDetector(
      onTap: _startEditing,
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.lg,
          vertical: AppSpacing.sm + 4,
        ),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: AppColors.border,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              _playerName,
              style: const TextStyle(
                color: AppColors.text,
                fontSize: 18,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(width: AppSpacing.sm),
            const Icon(
              Icons.edit_outlined,
              size: 16,
              color: AppColors.textSecondary,
            ),
          ],
        ),
      ),
    );
  }
}