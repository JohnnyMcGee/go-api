// package main

package netplayer

import (
	"database/sql"
	"fmt"
	"math"

	deep "github.com/patrikeh/go-deep"
	"github.com/patrikeh/go-deep/training"

	"go-api/db"
	"go-api/game"
	"go-api/player"
)

// var connStr = fmt.Sprintf("postgresql://%v:%v@%v/GO-db?sslmode=disable", config.Username, config.Password, config.Address)

// func main() {

// 	DB, err := db.ConnectDB(connStr)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}

// 	net := NewNetPlayer(DB)

// 	fmt.Println(data[0].Input, "=>", net.Predict(data[0].Input))
// 	fmt.Println(data[5].Input, "=>", net.Predict(data[5].Input))
// }

func NewNetPlayer(DB *sql.DB) *deep.Neural {
	input, response := db.GetTrainingData(DB)
	data := training.Examples{}
	for i := range input {
		ex := training.Example{Input: input[i], Response: []float64{response[i]}}
		data = append(data, ex)
	}
	n := NewNet(len(input[0]))
	TrainNet(n, &data)
	return n
}

func TrainNet(n *deep.Neural, data *training.Examples) {
	// params: learning rate, momentum, alpha decay, nesterov
	optimizer := training.NewSGD(0.05, 0.1, 1e-6, true)
	// params: optimizer, verbosity (print stats at every 50th iteration)
	trainer := training.NewTrainer(optimizer, 50)

	training, heldout := data.Split(0.5)
	trainer.Train(n, training, heldout, 1000) // training, validation, iterations
}

func NewNet(inputs int) *deep.Neural {
	n := deep.NewNeural(&deep.Config{
		/* Input dimensionality */
		Inputs: inputs,
		/* Two hidden layers consisting of two neurons each, and a single output */
		Layout: []int{2, 100, 1},
		/* Activation functions: Sigmoid, Tanh, ReLU, Linear */
		Activation: deep.ActivationSigmoid,
		/* Determines output layer activation & loss function:
		ModeRegression: linear outputs with MSE loss
		ModeMultiClass: softmax output with Cross Entropy loss
		ModeMultiLabel: sigmoid output with Cross Entropy loss
		ModeBinary: sigmoid output with binary CE loss */
		Mode: deep.ModeBinary,
		/* Weight initializers: {deep.NewNormal(μ, σ), deep.NewUniform(μ, σ)} */
		Weight: deep.NewNormal(1.0, 0.0),
		/* Apply bias */
		Bias: true,
	})

	return n
}

func BestPossibleMove(g game.Game, n *deep.Neural) game.Point {
	maxEval := math.Inf(-1)
	bestMove := game.Point{X: -1, Y: -1, Color: g.Turn}
	// b, w := g.Score["black"], g.Score["white"]

	// consider passing unless the game just started
	// if b > 10 && w > 10 && b+w > 60 {
	// 	maxEval = n.Predict(EncodeGame(&g, &bestMove))[0]
	// } else {
	// 	maxEval = math.Inf(-1)
	// }

	rng := player.NewUniqueRand(g.Board.Size())
	for {
		coord := rng.Coord()
		if coord[0] < 0 {
			break
		}
		p := g.Board.At(coord[0], coord[1])
		p = &game.Point{X: p.X, Y: p.Y, Color: g.Turn}

		if g.IsValidMove(*p) {
			encodedMove := EncodeGame(&g, p)
			eval := n.Predict(encodedMove)
			if eval[0] > maxEval {
				maxEval = eval[0]
				bestMove = *p
			}
		}
	}
	return bestMove
}

func EncodeGame(g *game.Game, p *game.Point) []float64 {
	var board []float64
	g.Board.ForEachPoint(func(p *game.Point) {
		input := make([]float64, 2, 2)
		if fmt.Sprintf("%v", *p) == "white" {
			input = append(input, 1)
		} else {
			input = append(input, 0)
		}
		if fmt.Sprintf("%v", *p) == "black" {
			input = append(input, 1)
		} else {
			input = append(input, 0)
		}
		board = append(board, input...)
	})

	move := db.Move{
		Turn:          g.Turn,
		Passed:        g.Passed,
		Kox:           g.Ko[0],
		Koy:           g.Ko[1],
		ScoreWhite:    float64(g.Score["white"]),
		ScoreBlack:    float64(g.Score["black"]),
		CapturesWhite: float64(g.Captures["white"]),
		CapturesBlack: float64(g.Captures["white"]),
		X:             p.X,
		Y:             p.Y,
	}

	encodedGameState := db.EncodeMove(move)
	encodedGameState = append(encodedGameState, board...)
	return encodedGameState
}
