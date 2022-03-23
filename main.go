package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	newGame()
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	router.Use(cors.New(config))
	router.GET("/board", getBoard)
	router.GET("/groups", getGroups)
	router.GET("/captures", getCaptures)
	router.GET("/new-game", getNewGame)
	router.POST("/moves", postMove)
	router.Run("localhost:8080")
}

var board gameBoard

var captures map[string]int

func newGame() {
	board = NewGameBoard(9)
	captures = map[string]int{
		"black": 0,
		"white": 0,
	}
}

func getNewGame(c *gin.Context) {
	newGame()
	c.JSON(http.StatusOK, "")
}

func getBoard(c *gin.Context) {
	// remove all data not required by frontend application before sending
	var simpleBoard [][]string
	for _, row := range board.Points() {
		var simpleRow []string
		for _, point := range row {
			simpleRow = append(simpleRow, point.Color)
		}
		simpleBoard = append(simpleBoard, simpleRow)
	}
	c.IndentedJSON(http.StatusOK, simpleBoard)
}

func getGroups(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, board.groups)
}

func getCaptures(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, captures)
}

type move struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

func postMove(c *gin.Context) {
	var newPoint point

	if err := c.BindJSON(&newPoint); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON data"})
		return
	}

	if !isValidMove(newPoint, board) {
		c.IndentedJSON(400, gin.H{"status": "Bad Request", "message": "move data invalid"})
		return
	}
	board.addPoint(newPoint)
	cap := board.doCaptures()
	for clr, num := range cap {
		captures[clr] += num
	}

	c.IndentedJSON(http.StatusCreated, newPoint)
}
