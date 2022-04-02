package player

// package main

import (
	"fmt"
	"go-api/game"
	"math"
	"math/rand"
	"time"
)

// var Game game.Game = game.NewGame(9)
// var color = "black"

// func main() {
// setupPoints := []game.Point{
// {X: 1, Y: 2, Color: "black"},
// {X: 2, Y: 2, Color: "white"},
// {X: 7, Y: 4, Color: "black"},
// {X: 2, Y: 3, Color: "white"},
// {X: 7, Y: 5, Color: "black"},
// {X: 0, Y: 2, Color: "white"},
// {X: 3, Y: 2, Color: "black"},
// {X: 8, Y: 5, Color: "white"},
// {X: 0, Y: 2, Color: "black"},
// }

// for _, p := range setupPoints {
// if Game.IsValidMove(p) {
// Game.Play(p)
// }
// }

// newMove := Move(Game, color)
// fmt.Println(newMove)
// }

// func legalMoves(g game.Game, color string) []game.Point {
// moves := []game.Point{}
// Game.Board.ForEachPoint(func(p *game.Point) {
// if g.IsValidMove(game.Point{X: p.X, Y: p.Y, Color: color}) {
// moves = append(moves, *p)
// }
// })
// return moves
// }

// func evaluateMoves(g game.Game, color string) [][2]int {
// moves := [][2]int{}
// maxScore := math.Inf(-1)
// Game.Board.ForEachPoint(func(p *game.Point) {
// if g.IsValidMove(game.Point{X: p.X, Y: p.Y, Color: color}) {
// gameCopy := g.DeepCopy()
// gameCopy.Play(game.Point{X: p.X, Y: p.Y, Color: color})
// totalScore := float64(gameCopy.Score[color] - gameCopy.Score[game.OppositeColor(color)])
// coords := [2]int{p.X, p.Y}
// if totalScore > maxScore {
// moves = [][2]int{coords}
// maxScore = totalScore
// } else if totalScore == maxScore {
// moves = append(moves, coords)
// }
// }
// })
// return moves
// }

func staticEvalByGroup(g game.Game, color string) float64 {
	score := map[string]float64{"black": 0, "white": 0}
	for _, grp := range g.Board.Groups {
		numEyes := 0
		numLiberties := 0
		xMax := math.Inf(-1)
		xMin := math.Inf(1)
		yMax := math.Inf(-1)
		yMin := math.Inf(1)

		for _, b := range grp.Bounds {
			bPoint := *g.Board.At(b[0], b[1])

			xMax = math.Max(float64(b[0]), xMax)
			xMin = math.Min(float64(b[0]), xMin)
			yMax = math.Max(float64(b[1]), yMax)
			yMin = math.Min(float64(b[1]), yMin)

			if numEyes < 2 && bPoint.Color == "" {
				numLiberties++
				eye := 1

				for _, p := range bPoint.AdjPoints(g.Board) {
					if p.GroupId != grp.ID {
						eye = 0
						break
					}
					numEyes += eye
				}
			}

			if numEyes > 1 {
				score[grp.Color] += 4 // multiple eyes weighted high
			} else {
				if numEyes == 1 {
					score[grp.Color] += 2 // single eye weighted slightly lower
				}
				score[grp.Color] += float64(numLiberties) * .25 // liberties weighted lower than eyes
			}

			area := (xMax - xMin - 1) * (yMax - yMin - 1)
			score[grp.Color] += area * .5 // area weighted lower than liberties

			size := float64(grp.Size())
			score[grp.Color] += size * .2 // size weighted lower than area

		}
	}

	return score[color] - score[game.OppositeColor(color)]
}

func staticEval(g game.Game, color string) float64 {
	basicScore := g.Score[color] - g.Score[game.OppositeColor(color)]
	groupScore := 0.0

	eyeScore := func(grp game.Group) int {
		score := 0
		for _, b := range grp.Bounds {
			numEyes := 0
			if g.Board.At(b[0], b[1]).IsAnEye(g.Board) {
				numEyes++
			}
			if numEyes > 0 {
				score += 1
			}
			if numEyes > 1 {
				score += 3
			}
		}
		return score
	}

	libsPerColor := map[string]float64{"white": 0, "black": 0}
	groupsPerColor := map[string]float64{"white": 0, "black": 0}
	for _, grp := range g.Board.Groups {
		if grp.Color == color {
			libsPerColor[color] += float64(grp.CountLiberties(g.Board))
			groupsPerColor[color]++
			groupScore += float64(eyeScore(*grp))

		} else if grp.Color == game.OppositeColor(color) {
			libsPerColor[game.OppositeColor(color)] += float64(grp.CountLiberties(g.Board))
			groupsPerColor[game.OppositeColor(color)]++
			groupScore -= float64(eyeScore(*grp))
		}
	}
	groupScore += ((groupsPerColor[color]) / libsPerColor[color])
	groupScore -= (groupsPerColor[game.OppositeColor(color)] / libsPerColor[game.OppositeColor(color)])

	return float64(basicScore) + groupScore
}

// Recursively evaluate possible moves and counter-moves using minimax algorithm
// returns eval score and slice of moves which result in that score
func minimax(g game.Game, depth int, alpha float64, beta float64, maximize bool) (float64, []game.Point) {
	if depth == 0 || g.Ended {
		var eval float64
		if maximize {
			eval = staticEvalByGroup(g, g.Turn)
		} else {
			eval = staticEvalByGroup(g, game.OppositeColor(g.Turn))
		}
		return float64(eval), []game.Point{}
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
		evaluate := func(testGame game.Game, p *game.Point) {
			eval, _ := minimax(testGame, depth-1, alpha, beta, false)
			if eval > maxEval {
				diff := eval - maxEval
				moves = []game.Point{*p}
				maxEval = eval
				if depth == 3 {
					fmt.Printf("Move: %v, %v, Diff: %v, maxEval: %v\n", p.X, p.Y, diff, maxEval)
				}
			} else if eval == maxEval {
				moves = append(moves, *p)
				if depth == 3 {
					fmt.Printf("Move: %v, %v, Diff: %v, maxEval: %v\n", p.X, p.Y, eval-maxEval, maxEval)
				}
			}
			alpha = math.Max(alpha, eval)
		}

		testPass := g.DeepCopy()
		testPass.Pass()
		evaluate(testPass, &game.Point{X: -1, Y: -1, Color: ""})

	OUTERMAX:
		for _, row := range g.Board.Getpoints() {
			for _, p := range row {
				if testGame, ok := testPoint(p); ok {
					evaluate(testGame, p)
					if beta <= alpha {
						break OUTERMAX
					}
				}
			}
		}
		return maxEval, moves

	} else {
		minEval := math.Inf(1)
		moves := []game.Point{}

		evaluate := func(testGame game.Game, p *game.Point) float64 {
			eval, _ := minimax(testGame, depth-1, alpha, beta, true)
			if eval < minEval {
				moves = []game.Point{*p}
				minEval = eval
			} else if eval == minEval {
				moves = append(moves, *p)
			}
			return eval
		}
	OUTERMIN:
		for _, row := range g.Board.Getpoints() {
			for _, p := range row {
				if testGame, ok := testPoint(p); ok {
					eval := evaluate(testGame, p)
					beta = math.Min(alpha, eval)
					if beta <= alpha {
						break OUTERMIN
					}
				}
			}
		}
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

	eval, moves := minimax(g, 4, math.Inf(-1), math.Inf(1), true)
	fmt.Printf("Eval Score: %v\nNum Equiv Moves: %v\n", eval, len(moves))

	if len(moves) == 0 {
		return p
	}

	return RandomMove(color, moves)
}
