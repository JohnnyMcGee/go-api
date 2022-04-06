package main

// package netplayer
import (
	"fmt"

	deep "github.com/patrikeh/go-deep"
	"github.com/patrikeh/go-deep/training"

	"go-api/config"
	"go-api/db"
)

var data = training.Examples{}

var connStr = fmt.Sprintf("postgresql://%v:%v@%v/GO-db?sslmode=disable", config.Username, config.Password, config.Address)

func main() {

	DB, err := db.ConnectDB(connStr)
	if err != nil {
		fmt.Println(err.Error())
	}
	input, response := db.GetTrainingData(DB)
	for i := range input {
		ex := training.Example{Input: input[i], Response: []float64{response[i]}}
		data = append(data, ex)
	}

	fmt.Println(response[0])
	fmt.Println(response[5])

	n := deep.NewNeural(&deep.Config{
		/* Input dimensionality */
		Inputs: len(input[0]),
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

	// params: learning rate, momentum, alpha decay, nesterov
	optimizer := training.NewSGD(0.05, 0.1, 1e-6, true)
	// params: optimizer, verbosity (print stats at every 50th iteration)
	trainer := training.NewTrainer(optimizer, 50)

	training, heldout := data.Split(0.5)
	trainer.Train(n, training, heldout, 1000) // training, validation, iterations

	fmt.Println(data[0].Input, "=>", n.Predict(data[0].Input))
	fmt.Println(data[5].Input, "=>", n.Predict(data[5].Input))

}
