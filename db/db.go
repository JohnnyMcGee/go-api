package db

// package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"

	"go-api/game"
)

// var connStr = fmt.Sprintf("postgresql://%v:%v@%v/GO-db?sslmode=disable", config.Username, config.Password, config.Address)

// func main() {
// 	Game := game.NewGame(9)
// 	setupPoints := []game.Point{
// 		{X: 1, Y: 2, Color: "black"},
// 		{X: 2, Y: 2, Color: "white"},
// 		{X: 7, Y: 4, Color: "black"},
// 		{X: 2, Y: 3, Color: "white"},
// 		{X: 7, Y: 5, Color: "black"},
// 		{X: 0, Y: 2, Color: "white"},
// 		{X: 3, Y: 2, Color: "black"},
// 		{X: 8, Y: 5, Color: "white"},
// 		{X: 0, Y: 2, Color: "black"},
// 	}

// 	for _, p := range setupPoints {
// 		if Game.IsValidMove(p) {
// 			Game.Play(p)
// 		}
// 	}

// 	db, _ := ConnectDB(connStr)
// 	CreateBoard(&Game.Board, db)
// }

func ConnectDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}

type Score struct {
	Black int `json:"black"`
	White int `json:"white"`
}

func CreateBoard(board *game.GameBoard, db *sql.DB) {
	columnStr := "("
	valueStr := "("
	board.ForEachPoint(func(p *game.Point) {
		columnStr += fmt.Sprintf("\"%v,%v\", ", p.X, p.Y)
		valueStr += fmt.Sprintf("'%v', ", p.Color)
	})
	columnStr = columnStr[:len(columnStr)-2] + ")"
	valueStr = valueStr[:len(valueStr)-2] + ")"

	fmt.Println(columnStr)
	fmt.Println(valueStr)
	q := fmt.Sprintf("INSERT INTO board %v VALUES %v", columnStr, valueStr)
	db.Query(q)
}

func GetHandler(c *gin.Context, db *sql.DB) {
	var res Score
	var scores []Score

	rows, err := db.Query("SELECT * FROM scores")
	defer rows.Close()
	if err != nil {
		log.Fatalln(err)
		c.JSON(http.StatusInternalServerError, "An error occured")
	}

	for rows.Next() {
		rows.Scan(&res.Black, &res.White)
		scores = append(scores, res)
	}

	c.JSON(http.StatusOK, scores)
}

func PostHandler(c *gin.Context, db *sql.DB) {
	newScore := Score{}

	if err := c.BindJSON(&newScore); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON data"})
		return
	}
	fmt.Printf("%v\n", newScore)

	if newScore.Black >= 0 && newScore.White >= 0 {
		_, err := db.Exec("INSERT INTO scores (black, white) VALUES ($1, $2)", newScore.Black, newScore.White)
		if err != nil {
			log.Fatalf("An error occured while executing query: %v", err)
		}
	}
}
