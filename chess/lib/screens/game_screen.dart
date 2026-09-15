import 'package:chess/cubits/game_cubit.dart';
import 'package:chess/cubits/game_screen_state.dart';
import 'package:chess/services/game_socket_service.dart';
import 'package:chess/widgets/game_screen/chess_board.dart';
import 'package:chess/widgets/game_screen/player_bar.dart';
import 'package:chess/widgets/game_screen/promotion_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:chess/models/game_state.dart';

class GameScreen extends StatelessWidget {
  final GameState game;
  final String playerId;
  final GameSocketService socket;

  const GameScreen({
    super.key,
    required this.game,
    required this.playerId,
    required this.socket,
  });

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => GameCubit(game: game, playerId: playerId, socket: socket),
      child: Scaffold(
        body: SafeArea(
          child: BlocBuilder<GameCubit, GameScreenState>(
            builder: (context, state) {
              final isWhite = playerId == state.game.white?.id;

              final myPlayer = isWhite ? state.game.white : state.game.black;

              final opponent = isWhite ? state.game.black : state.game.white;

              final myTime = isWhite ? state.whiteTime : state.blackTime;

              final opponentTime = isWhite ? state.blackTime : state.whiteTime;

              final myTurn = isWhite
                  ? state.game.active == ActiveColor.white
                  : state.game.active == ActiveColor.black;

              return Column(
                children: [
                  PlayerBar(
                    name: opponent?.nickname ?? 'Opponent',
                    time: opponentTime,
                    isActive: !myTurn,
                  ),

                  const Spacer(),

                  Stack(
                    alignment: Alignment.center,
                    children: [
                      ChessBoard(
                        game: state.game,
                        isWhite: isWhite,
                        selectedSquare: state.selectedSquare,
                        legalMoves: state.legalMoves,
                        onSquareTap: (square) {
                          context.read<GameCubit>().selectSquare(square);
                        },
                      ),

                      if (state.isPromotionPending)
                        PromotionPicker(
                          isWhite: isWhite,
                          onSelected: (promotion) {
                            context.read<GameCubit>().promote(promotion);
                          },
                        ),
                    ],
                  ),

                  const Spacer(),

                  PlayerBar(
                    name: myPlayer?.nickname ?? 'You',
                    time: myTime,
                    isActive: myTurn,
                  ),
                ],
              );
            },
          ),
        ),
      ),
    );
  }
}
