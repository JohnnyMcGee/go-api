package player

// package main

import (
	"fmt"
	"go-api/game"
	"math/rand"
	"time"
)

var Game game.Game = game.NewGame(9)
var color = "black"

// func main() {
// for i := 0; i < 10; i++ {
// time.Sleep(100)
// move := RandomMove(Game)
// fmt.Println(move)
// Game.Play(move)
// }

// }

func legalMoves(g game.Game, color string) []game.Point {
	moves := []game.Point{}
	Game.Board.ForEachPoint(func(p *game.Point) {
		if g.IsValidMove(game.Point{X: p.X, Y: p.Y, Color: color}) {
			moves = append(moves, *p)
		}
	})
	return moves
}

func RandomMove(g game.Game, color string) game.Point {
	moves := legalMoves(g, color)
	fmt.Println(g.Turn)
	p := game.Point{X: -1, Y: -1, Color: ""}
	if len(moves) == 0 {
		fmt.Println("Pass")
		return p
	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	n := r.Intn(len(moves))
	p = game.Point{
		X:     moves[n].X,
		Y:     moves[n].Y,
		Color: color,
	}
	return p
}
