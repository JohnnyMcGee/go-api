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
type scanContext struct {
	game            game.Game
	point           game.Point
	group           game.Group
	scanned         []game.Point
	depth           int
	connectionDepth float64
}

// recursively explore surrounding points
//  check for eyes (enclosed areas within group)
// and distance to groups of same color (potential connections)
func proximityScan(c scanContext) (scannedPoints []game.Point, isAnEye bool, connDepth float64) {
	// verify this point was not scanned previously
	for _, scannedPoint := range c.scanned {
		if scannedPoint.X == c.point.X && scannedPoint.Y == c.point.Y {
			c.scanned = append(c.scanned, c.point)
			return c.scanned, true, c.connectionDepth
		}
	}
	// search the neighboring points for friends and enemies
	for _, adjP := range c.point.AdjPoints(c.game.Board) {
		if adjP.GroupId != c.group.ID {
			if adjP.Color == c.group.Color {
				c.connectionDepth = math.Max(c.connectionDepth, float64(c.depth))
			}
			if c.depth == 0 && adjP.Color != "" {
				return c.scanned, false, c.connectionDepth
			}
			c.scanned = append(c.scanned, c.point)
			return proximityScan(scanContext{
				game:            c.game,
				point:           adjP,
				group:           c.group,
				scanned:         c.scanned,
				depth:           c.depth - 1,
				connectionDepth: c.connectionDepth,
			})
		}
	}
	c.scanned = append(c.scanned, c.point)
	return c.scanned, true, c.connectionDepth
}

type EvalConfig struct {
	complexity      int
	eyeRecursion    int
	eyeWeight       float64
	libWeight       float64
	areaWeight      float64
	sizeWeight      float64
	captureWeight   float64
	koWeight        float64
	densityWeight   float64
	connDepthWeight float64
	groupAvgWeight  float64
}

var DefaultConfig = EvalConfig{
	complexity:      6e7,
	eyeRecursion:    8,
	eyeWeight:       1,
	libWeight:       .1,
	areaWeight:      .5,
	sizeWeight:      .1,
	captureWeight:   1.2,
	koWeight:        .1,
	densityWeight:   0.3,
	connDepthWeight: .6,
	groupAvgWeight:  .05,
}

func staticEvalByGroup(g game.Game, color string, config EvalConfig) float64 {
	score := map[string]float64{"black": 0, "white": 0}
	groupCount := map[string]int{"black": 0, "white": 0}

	for _, grp := range g.Board.Groups {
		groupCount[grp.Color]++
		numEyes := 0
		numLiberties := 0
		xMax := math.Inf(-1)
		xMin := math.Inf(1)
		yMax := math.Inf(-1)
		yMin := math.Inf(1)

		ScannedPoints := []game.Point{}

		ConnectionDepth := math.Inf(-1)

		for _, b := range grp.Bounds {
			bPoint := *g.Board.At(b[0], b[1])

			xMax = math.Max(float64(b[0]), xMax)
			xMin = math.Min(float64(b[0]), xMin)
			yMax = math.Max(float64(b[1]), yMax)
			yMin = math.Min(float64(b[1]), yMin)

			if numEyes < 2 && bPoint.Color == "" {
				numLiberties++
				// verify that this point was not part of a prior 'isAnyEye' search
				var isInScannedPoints bool
				for _, cp := range ScannedPoints {
					isInScannedPoints = cp.X == bPoint.X && cp.Y == bPoint.Y
				}
				// search area around the point to see if it is fully enclosed (is an eye)
				Eye := false
				if !isInScannedPoints {
					scannedPoints, eye, connectionDepth := proximityScan(scanContext{
						game:            g,
						point:           bPoint,
						group:           *grp,
						scanned:         ScannedPoints,
						depth:           config.eyeRecursion,
						connectionDepth: ConnectionDepth,
					})
					ScannedPoints, Eye, ConnectionDepth = scannedPoints, eye, math.Max(ConnectionDepth, connectionDepth)
				}
				if Eye {
					numEyes++
				}
			}

			if numEyes > 1 {
				score[grp.Color] += 6 * config.eyeWeight // multiple eyes weighted high
			} else {
				if numEyes == 1 {
					score[grp.Color] += 2 * config.eyeWeight // single eye weighted slightly lower
				}
				score[grp.Color] += float64(numLiberties) * config.libWeight // liberties weighted lower than eyes
			}

			area := (xMax - xMin) * (yMax - yMin)
			score[grp.Color] += area * config.areaWeight // area weighted lower than liberties

			size := float64(grp.Size())
			score[grp.Color] += size * config.sizeWeight // size weighted lower than area

			if area > 0 {
				score[grp.Color] += (size / area) * 100 * config.densityWeight
			}

			if !math.IsInf(ConnectionDepth, -1) {
				score[grp.Color] += ConnectionDepth * config.connDepthWeight
			}

		}
	}

	oppColor := game.OppositeColor(color)
	captureScore := float64(g.Captures[oppColor]-g.Captures[color]) * 100

	var koScore float64 = 0
	if g.Ko != [2]int{-1, -1} {
		koScore = -100 * config.koWeight
	}

	if groupCount["white"] > 0 && groupCount["black"] > 0 {
		score[color] += (score[color] / float64(groupCount[color])) * config.groupAvgWeight
		score[oppColor] += (score[oppColor] / float64(groupCount[oppColor])) * config.groupAvgWeight
	}

	return score[color] - score[oppColor] + captureScore*config.captureWeight + koScore
}

// Recursively evaluate possible moves and counter-moves using minimax algorithm
// returns eval score and slice of moves which result in that score
func minimax(g game.Game, depth int, alpha float64, beta float64, maximize bool, noPass bool) (float64, []game.Point) {
	if depth == 0 || g.Ended {
		var eval float64
		if maximize {
			eval = staticEvalByGroup(g, g.Turn, DefaultConfig)
		} else {
			eval = staticEvalByGroup(g, game.OppositeColor(g.Turn), DefaultConfig)
		}
		return float64(eval), []game.Point{}
	}

	// Evaluate resulting game state for a given point
	// Returns result state, true for a valid move, or current state, false for invalid move
	testPoint := func(p *game.Point) (game.Game, bool) {
		testPoint := game.Point{X: p.X, Y: p.Y, Color: g.Turn}
		if g.IsValidMove(testPoint) {
			testGame := g.DeepCopy()
			testGame.PlayWithoutScoring(testPoint)
			return testGame, true
		} else {
			return g, false
		}
	}

	if maximize {
		maxEval := math.Inf(-1)
		moves := []game.Point{}
		evaluate := func(testGame game.Game, p *game.Point) {
			eval, _ := minimax(testGame, depth-1, alpha, beta, false, noPass)
			if eval > maxEval {
				moves = []game.Point{*p}
				maxEval = eval

			} else if eval == maxEval {
				moves = append(moves, *p)
			}
			alpha = math.Max(alpha, eval)
		}

		if !noPass {
			testPass := g.DeepCopy()
			testPass.Pass()
			evaluate(testPass, &game.Point{X: -1, Y: -1, Color: ""})
		}

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
			eval, _ := minimax(testGame, depth-1, alpha, beta, true, noPass)
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

// find the max depth for which, given the number of pieces on the board (coverage)
// would yield fewer options than maxComplexity
func maximumDepth(coverage int, maxComplexity int) int {
	depth := 1
	options := 81 - coverage
	for {
		next := options * (81 - coverage - depth)
		if next >= maxComplexity {
			break
		}
		depth++
		options = next
	}
	return depth
}

func Move(g game.Game, color string, coverage int) game.Point {
	p := game.Point{X: -1, Y: -1, Color: ""}

	depth := maximumDepth(coverage, DefaultConfig.complexity)

	fmt.Printf("Coverage: %v\nPossible Moves: %v\nDepth: %v\n", coverage, 81-coverage, depth)

	// Player will not pass if <75% of board is covered
	noPass := (float64(coverage) / math.Pow(float64(g.Board.Size()), 2)) < .75

	eval, moves := minimax(g, depth, math.Inf(-1), math.Inf(1), true, noPass)
	fmt.Printf("Eval Score: %v\nNum Equiv Moves: %v\n", eval, len(moves))

	if len(moves) == 0 {
		return p
	}

	return RandomMove(color, moves)
}
