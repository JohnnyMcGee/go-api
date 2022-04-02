package player

// package main

import (
	"fmt"
	"go-api/game"
	"math"
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

func evaluateMoves(g game.Game, color string) [][2]int {
	moves := [][2]int{}
	maxScore := math.Inf(-1)
	Game.Board.ForEachPoint(func(p *game.Point) {
		if g.IsValidMove(game.Point{X: p.X, Y: p.Y, Color: color}) {
			gameCopy := g.DeepCopy()
			gameCopy.Play(game.Point{X: p.X, Y: p.Y, Color: color})
			totalScore := float64(gameCopy.Score[color] - gameCopy.Score[game.OppositeColor(color)])
			coords := [2]int{p.X, p.Y}
			if totalScore > maxScore {
				moves = [][2]int{coords}
				maxScore = totalScore
			} else if totalScore == maxScore {
				moves = append(moves, coords)
			}
		}
	})
	return moves
}

// Recursively evaluate possible moves and counter-moves using minimax algorithm
// returns eval score and slice of moves which result in that score
func minimax(g game.Game, depth int, maximize bool) (float64, []game.Point) {
	if depth == 0 || g.Ended {
		var staticEval int
		if maximize {
			staticEval = g.Score[g.Turn] - g.Score[game.OppositeColor(g.Turn)]
		} else {
			staticEval = g.Score[game.OppositeColor(g.Turn)] - g.Score[g.Turn]
		}
		return float64(staticEval), []game.Point{}
	}

	// Evaluate resulting game state for a given point
	// Returns result state, true for a valid move, or current state, false for invalid move
	testPoint := func(p *game.Point) (game.Game, bool) {
		testPoint := game.Point{X: p.X, Y: p.Y, Color: g.Turn}
		if g.IsValidMove(testPoint) {
			testGame := g.DeepCopy()
			testGame.Play(testPoint)
			return testGame, true
		} else {
			return g, false
		}
	}

	if maximize {
		maxEval := math.Inf(-1)
		moves := []game.Point{}
		g.Board.ForEachPoint(func(p *game.Point) {
			if testGame, ok := testPoint(p); ok {
				eval, _ := minimax(testGame, depth-1, false)
				if eval > maxEval {
					moves = []game.Point{*p}
					maxEval = eval
				} else if eval == maxEval {
					moves = append(moves, *p)
				}
			}
			if depth == 3 {
				fmt.Printf("Move: %v, %v: maxEval: %v\n", p.X, p.Y, maxEval)
			}
		})
		return maxEval, moves

	} else {
		minEval := math.Inf(1)
		moves := []game.Point{}
		g.Board.ForEachPoint(func(p *game.Point) {
			if testGame, ok := testPoint(p); ok {
				eval, _ := minimax(testGame, depth-1, true)
				if eval < minEval {
					moves = []game.Point{*p}
					minEval = eval
				} else if eval == minEval {
					moves = append(moves, *p)
				}
			}
		})
		return minEval, moves
	}
}

// pick a random move from list of moves
func RandomMove(color string, moves []game.Point) game.Point {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	n := r.Intn(len(moves))
	return game.Point{
		X:     moves[n].X,
		Y:     moves[n].Y,
		Color: color,
	}
}

func Move(g game.Game, color string) game.Point {
	p := game.Point{X: -1, Y: -1, Color: ""}

	eval, moves := minimax(g, 3, true)
	fmt.Printf("Eval Score: %v\nNum Equiv Moves: %v\n", eval, len(moves))

	if len(moves) == 0 {
		return p
	}

	return RandomMove(color, moves)
}
