package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"
)

func main() {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	router.Use(cors.New(config))
	router.GET("/board", getBoard)
	router.GET("/moves", getMoves)
	router.POST("/moves", postMove)
	router.Run("localhost:8080")
}

var moves = []move{}

// var groups = map[uint]map[uint]bool{}

func getBoard(c *gin.Context) {
	// remove all data not required by frontend application before sending
	var simpleBoard [][]string
	for _, row := range board {
		var simpleRow []string
		for _, point := range row {
			simpleRow = append(simpleRow, point.Color)
		}
		simpleBoard = append(simpleBoard, simpleRow)
	}
	c.IndentedJSON(http.StatusOK, simpleBoard)
}

func getMoves(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, moves)
}

func isInvalidMove(newMove move) bool {
	outOfRangeXY := newMove.X >= boardSize || newMove.X < 0 || newMove.Y >= boardSize || newMove.Y < 0
	invalidColor := newMove.Color != "white" && newMove.Color != "black"
	pointUnavailable := board[newMove.Y][newMove.X].Color != ""
	return outOfRangeXY || invalidColor || pointUnavailable
}

func postMove(c *gin.Context) {
	var newMove move

	if err := c.BindJSON(&newMove); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON data"})
		return
	}

	if isInvalidMove(newMove) {
		c.IndentedJSON(400, gin.H{"status": "Bad Request", "message": "move data invalid"})
		return
	}

	board[newMove.Y][newMove.X] = point{Color: newMove.Color, Group: 0}

	newMove.ID = len(moves)
	moves = append(moves, newMove)
	c.IndentedJSON(http.StatusCreated, newMove)
}

// func createPointFromMove(newMove move) point {
// var x, y int = newMove.X, newMove.Y + 1

// list groups found at adjacent points

// select the first group found of same color

// if no same-color groups found, create a new group with unique id

// categorize remaining adjacent points as liberties or non liberties

// add adjacent points to selected group

// if multiple same-color groups found, add unique adjacent points to selected group

// delete superfluous groups

// remove new point from selected group

// for each opponent color group, mark new point as non liberty

// create and return new point with the selected group
// }
