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

type move struct {
	ID    int    `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

var moves = []move{}

// var groups = map[uint]map[uint]bool{}

type point struct {
	Color string `json:"color"`
	Group uint   `json:"group"`
}

var board [][]string = generateBoard(boardSize)

const boardSize = 9

func generateBoard(size int) [][]string {
	var board = [][]string{}
	for y := 0; y < size; y++ {
		col := []string{}
		for x := 0; x < size; x++ {
			col = append(col, "")
		}
		board = append(board, col)
	}
	return board
}

func getBoard(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, board)
}

func getMoves(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, moves)
}

func postMove(c *gin.Context) {
	var newMove move

	if err := c.BindJSON(&newMove); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON data"})
		return
	}

	outOfRangeXY := newMove.X > 2 || newMove.X < 0 || newMove.Y > 2 || newMove.Y < 0
	invalidColor := newMove.Color != "white" && newMove.Color != "black"
	pointUnavailable := board[newMove.Y][newMove.X] != ""
	if outOfRangeXY || invalidColor || pointUnavailable {
		c.IndentedJSON(400, gin.H{"status": "Bad Request", "message": "move data invalid"})
		return
	}

	// boundary := [4][2]int{
	// 	[2]int{newMove.Y + 1, newMove.X},
	// 	[2]int{newMove.Y, newMove.X + 1},
	// 	[2]int{newMove.Y - 1, newMove.X},
	// 	[2]int{newMove.Y, newMove.X - 1},
	// }

	board[newMove.Y][newMove.X] = newMove.Color
	newMove.ID = len(moves)
	moves = append(moves, newMove)
	c.IndentedJSON(http.StatusCreated, newMove)
}
