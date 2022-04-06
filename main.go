package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"go-api/config"
	"go-api/db"
	"go-api/game"
	"go-api/player"
)

var connStr = fmt.Sprintf("postgresql://%v:%v@%v/GO-db?sslmode=disable", config.Username, config.Password, config.Address)

var DB, _ = db.ConnectDB(connStr)

func main() {
	Game = handleNewGame(9)
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
	router.GET("/pass", getPass)
	router.GET("/resign", getResign)
	router.GET("/player-move/:color", getPlayerMove)
	router.GET("/random-move/:color", getRandomMove)
	router.POST("/moves", postMove)
	router.Run("localhost:8080")
}

var Game game.Game

func handleNewGame(size int) game.Game {
	Game = game.NewGame(9)
	db.CreateGame(&Game, DB)
	return Game
}

func handleMove(c *gin.Context, p *game.Point) {
	if Game.IsValidMove(*p) {
		db.CreateMove(&Game, p, DB)
		Game.Play(*p)
		db.UpdateGame(&Game, DB)
		c.JSON(http.StatusOK, *Game.Board.At(p.X, p.Y))

	} else if p.X == -1 || p.Y == -1 {
		getPass(c)
	} else {
		c.JSON(400, gin.H{"status": "Bad Request", "message": "move data invalid"})
	}
}

func getPlayerMove(c *gin.Context) {
	color := c.Param("color")
	move := player.Move(Game, color)
	handleMove(c, &move)
}

func getRandomMove(c *gin.Context) {
	color := c.Param("color")
	move := player.RandomMove(Game, color)
	handleMove(c, &move)
}

func getResign(c *gin.Context) {
	db.CreateMove(&Game, &game.Point{X: -1, Y: -1, Color: ""}, DB)
	Game.Resign(Game.Turn)
	c.JSON(http.StatusOK, "Game Over")
	db.UpdateGame(&Game, DB)
}

func getPass(c *gin.Context) {
	db.CreateMove(&Game, &game.Point{X: -1, Y: -1, Color: ""}, DB)
	Game.Pass()
	db.UpdateGame(&Game, DB)
	if Game.Ended {
		c.JSON(http.StatusOK, "Game Over")
	} else {
		c.JSON(http.StatusOK, Game.Turn)
	}
}

func postMove(c *gin.Context) {
	var newPoint game.Point
	if err := c.BindJSON(&newPoint); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON data"})
		return
	}
	handleMove(c, &newPoint)
}

func getNewGame(c *gin.Context) {
	Game = handleNewGame(9)
	c.JSON(http.StatusOK, "")
}

// simplify gameboard before sending to client
func getBoard(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, simplifyBoard(Game.Board))
}

func getGroups(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, Game.Board.Groups)
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
	c.IndentedJSON(http.StatusOK, simplifyGame(Game))
}

type simplePoint struct {
	X         int             `json:"x"`
	Y         int             `json:"y"`
	Color     string          `json:"color"`
	Permit    map[string]bool `json:"permit"`
	Territory string          `json:"territory"`
}

func simplifyPoint(p game.Point) simplePoint {
	return simplePoint{X: p.X, Y: p.Y, Color: p.Color, Permit: p.Permit, Territory: p.Territory}
}

func simplifyBoard(b game.GameBoard) [][]simplePoint {
	var simpleBoard [][]simplePoint
	for _, row := range Game.Board.Points() {
		var simpleRow []simplePoint
		for _, point := range row {
			simpleRow = append(simpleRow, simplifyPoint(point))
		}
		simpleBoard = append(simpleBoard, simpleRow)
	}
	return simpleBoard
}

type simpleGame struct {
	Board  [][]simplePoint `json:"board"`
	Score  map[string]int  `json:"score"`
	Turn   string          `json:"turn"`
	Passed bool            `json:"passed"`
	Ended  bool            `json:"ended"`
	Winner string          `json:"winner"`
}

func simplifyGame(g game.Game) simpleGame {
	return simpleGame{
		Board:  simplifyBoard(g.Board),
		Score:  g.Score,
		Turn:   g.Turn,
		Passed: g.Passed,
		Ended:  g.Ended,
		Winner: g.Winner,
	}
}
