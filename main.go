package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	router.Use(cors.New(config))
	router.GET("/board", getBoard)
	router.GET("/groups", getGroups)
	router.GET("/captures", getCaptures)
	router.GET("/score", getScore)
	router.GET("/ko", getKo)
	router.GET("/active-player", getActivePlayer)
	router.GET("/game", getGame)
	router.GET("/new-game", getNewGame)
	router.POST("/moves", postMove)
	router.Run("localhost:8080")
}

var Game = NewGame(9)

func getNewGame(c *gin.Context) {
	Game = NewGame(9)
	c.JSON(http.StatusOK, "")
}

// simplify gameboard before sending to client
func getBoard(c *gin.Context) {
	type simplePoint struct {
		Color  string          `json:"color"`
		Permit map[string]bool `json:"permit"`
	}
	simplify := func(p point) simplePoint {
		return simplePoint{Color: p.Color, Permit: p.Permit}
	}
	var simpleBoard [][]simplePoint
	for _, row := range Game.Board.Points() {
		var simpleRow []simplePoint
		for _, point := range row {
			simpleRow = append(simpleRow, simplify(point))
		}
		simpleBoard = append(simpleBoard, simpleRow)
	}
	c.IndentedJSON(http.StatusOK, simpleBoard)
}

func getGroups(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, Game.Board.groups)
}

func getCaptures(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, Game.Captures)
}

func getScore(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, Game.Score)
}

func getKo(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, Game.Ko)
}

func getActivePlayer(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, Game.Turn)
}

func getGame(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, Game)
}

func postMove(c *gin.Context) {
	var newPoint point

	if err := c.BindJSON(&newPoint); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON data"})
		return
	}

	if Game.isValidMove(newPoint) {
		Game.play(newPoint)
		c.IndentedJSON(http.StatusCreated, Game.Board.at(newPoint.X, newPoint.Y))
	} else {
		c.IndentedJSON(400, gin.H{"status": "Bad Request", "message": "move data invalid"})
	}
}
