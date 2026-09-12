import 'package:chess/cubits/game_cubit.dart';
import 'package:chess/cubits/game_screen_state.dart';
import 'package:chess/widgets/game_screen/chess_board.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:chess/models/game_state.dart';

class GameScreen extends StatelessWidget {
  final GameState game;
  final String playerId;

  const GameScreen({super.key, required this.game, required this.playerId});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return BlocProvider(
      create: (_) => GameCubit(game),
      child: Scaffold(
        body: SafeArea(
          child: Column(
            children: [
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 16,
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Opponent', style: theme.textTheme.titleMedium),
                    Text(
                      '10:00',
                      style: theme.textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),

              const Spacer(),

              BlocBuilder<GameCubit, GameScreenState>(
                builder: (context, state) {
                  return ChessBoard(
                    game: state.game,
                    selectedSquare: state.selectedSquare,
                    legalMoves: state.legalMoves,
                    onSquareTap: (square) {
                      context.read<GameCubit>().selectSquare(square);
                    },
                  );
                },
              ),

              const Spacer(),

              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 16,
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('You', style: theme.textTheme.titleMedium),
                    Text(
                      '10:00',
                      style: theme.textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
