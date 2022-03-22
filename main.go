package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()
	router.GET("/nodes", getNodes)
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

var nodes = [][]string{
	[]string{"", "", ""},
	[]string{"", "", ""},
	[]string{"", "", ""},
}

func getNodes(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, nodes)
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
	nodeUnavailable := nodes[newMove.X][newMove.Y] != ""
	if outOfRangeXY || invalidColor || nodeUnavailable {
		c.IndentedJSON(400, gin.H{"status": "Bad Request", "message": "move data invalid"})
		return
	}

	nodes[newMove.X][newMove.Y] = newMove.Color
	newMove.ID = len(moves)
	moves = append(moves, newMove)
	c.IndentedJSON(http.StatusCreated, newMove)
}
