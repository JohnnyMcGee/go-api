package db

// package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"go-api/game"
)

// var connStr = fmt.Sprintf("postgresql://%v:%v@%v/GO-db?sslmode=disable", config.Username, config.Password, config.Address)

// func main() {
// Game := game.NewGame(9)
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

// db, _ := ConnectDB(connStr)
// CreateBoard(&Game.Board, db)
// }

func ConnectDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}

func CreateGame(g *game.Game, db *sql.DB) {
	var id int
	q := fmt.Sprintf(`
		INSERT INTO game (whitescore, blackscore, turn, ended, winner, whitecaptures, blackcaptures, kox, koy, passed)
		VALUES (%v, %v, '%v', %v, '', %v, %v, %v, %v, %v) RETURNING id;`,
		g.Score["white"], g.Score["black"], g.Turn, g.Ended, g.Captures["white"], g.Captures["black"], g.Ko[0], g.Ko[1], g.Passed)
	err := db.QueryRow(q).Scan(&id)
	if err != nil {
		fmt.Println("Exec err:", err.Error())
	}
	g.ID = id
	CreateBoard(&g.Board, db)
}

func UpdateGame(g *game.Game, db *sql.DB) {
	winner := ""
	if g.Ended {
		winner = g.Winner
	}

	q := fmt.Sprintf(`
	UPDATE game 
	SET whitescore = %v, blackscore = %v, turn = '%v', ended = %v, winner = '%v', whitecaptures = %v, blackcaptures = %v, kox = %v, koy = %v, passed = %v 
	WHERE id = %v
	`,
		g.Score["white"], g.Score["black"], g.Turn, g.Ended, winner, g.Captures["white"], g.Captures["black"], g.Ko[0], g.Ko[1], g.Passed, g.ID)

	_, err := db.Exec(q)
	if err != nil {
		fmt.Println(err.Error())
	}

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

	var id int

	q := fmt.Sprintf("INSERT INTO board %v VALUES %v RETURNING id;", columnStr, valueStr)
	err := db.QueryRow(q).Scan(&id)
	if err != nil {
		fmt.Println(err.Error())
	}
	board.ID = id
}

func CreateMove(g *game.Game, p *game.Point, db *sql.DB) {
	q := fmt.Sprintf(`
		INSERT INTO move (ScoreWhite, ScoreBlack, Turn, CapturesWhite, CapturesBlack, KoX, KoY, Passed, GameID, BoardID, X, Y, Datetime)
		VALUES (%v, %v, '%v', %v, %v, %v, %v, %v, %v, %v, %v, %v, '%v');`,
		g.Score["white"], g.Score["black"], g.Turn, g.Captures["white"], g.Captures["black"], g.Ko[0], g.Ko[1], g.Passed, g.ID, g.Board.ID, p.X, p.Y, time.Now().Format("2006-01-02 15:04:05"))
	_, err := db.Query(q)
	if err != nil {
		fmt.Println(err.Error())
	}
	CreateBoard(&g.Board, db)
	UpdateGame(g, db)
}

// func GetHandler(c *gin.Context, db *sql.DB) {
// 	var res Score
// 	var scores []Score

// 	rows, err := db.Query("SELECT * FROM scores")
// 	defer rows.Close()
// 	if err != nil {
// 		log.Fatalln(err)
// 		c.JSON(http.StatusInternalServerError, "An error occured")
// 	}

// 	for rows.Next() {
// 		rows.Scan(&res.Black, &res.White)
// 		scores = append(scores, res)
// 	}

// 	c.JSON(http.StatusOK, scores)
// }

// func PostHandler(c *gin.Context, db *sql.DB) {
// 	newScore := Score{}

// 	if err := c.BindJSON(&newScore); err != nil {
// 		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON data"})
// 		return
// 	}
// 	fmt.Printf("%v\n", newScore)

// 	if newScore.Black >= 0 && newScore.White >= 0 {
// 		_, err := db.Exec("INSERT INTO scores (black, white) VALUES ($1, $2)", newScore.Black, newScore.White)
// 		if err != nil {
// 			log.Fatalf("An error occured while executing query: %v", err)
// 		}
// 	}
// }
