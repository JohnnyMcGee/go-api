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
// 	DB, err := ConnectDB(connStr)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	input, target := GetTrainingData(DB)

// 	fmt.Println(input)
// 	fmt.Println(target)
// }

type Move struct {
	Turn          string
	Passed        bool
	Kox           int
	Koy           int
	ScoreWhite    float64
	ScoreBlack    float64
	CapturesWhite float64
	CapturesBlack float64
	X             int
	Y             int
	ID            int
	BoardId       int
	GameId        int
}

func GetTrainingData(db *sql.DB) ([][]float64, []float64) {
	input := [][]float64{}
	target := []float64{}

	var winner string
	rows, err := db.Query(`SELECT winner FROM game WHERE winner = 'white' OR winner = 'black'`)
	if err != nil {
		fmt.Println(err.Error())
	}
	for rows.Next() {
		rows.Scan(&winner)
		inp, tgt := GetEncodedMoves(38, winner, db)
		input = append(input, inp...)
		target = append(target, tgt...)
	}
	return input, target
}

func GetEncodedMoves(GameID int, winner string, db *sql.DB) ([][]float64, []float64) {
	var moves []Move
	rows, err := db.Query(`
	SELECT turn, passed, kox, koy, scorewhite, scoreblack, captureswhite, capturesblack, x, y, id, boardid, gameid
	FROM move
	WHERE gameid=38;
	`)

	if err != nil {
		fmt.Println(err.Error())
	}

	for rows.Next() {
		var move Move
		rows.Scan(&move.Turn, &move.Passed, &move.Kox, &move.Koy, &move.ScoreWhite, &move.ScoreBlack, &move.CapturesWhite, &move.CapturesBlack, &move.X, &move.Y, &move.ID, &move.BoardId, &move.GameId)
		moves = append(moves, move)
	}

	var boards [][]float64

	rows, err = db.Query(`SELECT board.* FROM "move" INNER JOIN board ON move.boardid = board.id WHERE move.gameid =38;`)

	for rows.Next() {
		input := BoardFromRow(rows)
		boards = append(boards, input)
	}

	input := make([][]float64, len(moves))
	// Target tells us whether the team which played this move won or lost
	target := make([]float64, len(moves))

	// Combine encoded input
	for i, move := range moves {
		in := EncodeMove(moves[i])
		in = append(in, boards[i]...)
		input[i] = in

		if move.Turn == winner {
			target[i] = 1
		} else {
			target[i] = 0
		}
	}
	return input, target
}

func BoardFromRow(rows *sql.Rows) []float64 {
	cols, err := rows.Columns()

	if err != nil {
		fmt.Println(err.Error())
	}

	// create slice of interfaces to represent each column
	// and a second slice to contain pointers to each item in the columns slice
	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	// Scan the result into the column pointers...
	if err := rows.Scan(columnPointers...); err != nil {
		fmt.Println(err)
	}

	input := []float64{}

	for i, colName := range cols {
		val := columnPointers[i].(*interface{})
		// m[colName] = *val
		if colName != "id" {
			input = append(input, EncodePoint(val)...)
		}
	}
	return input
}

// Takes a point color (string) returns slice []{isWhite, isBlack}
func EncodePoint(point *interface{}) []float64 {
	input := make([]float64, 2, 2)
	if fmt.Sprintf("%v", *point) == "white" {
		input = append(input, 1)
	} else {
		input = append(input, 0)
	}
	if fmt.Sprintf("%v", *point) == "black" {
		input = append(input, 1)
	} else {
		input = append(input, 0)
	}
	return input
}

func EncodeMove(move Move) []float64 {
	input := []float64{}

	if move.Turn == "white" {
		input = append(input, 1)
	} else {
		input = append(input, 0)
	}
	if move.Passed {
		input = append(input, 1)
	} else {
		input = append(input, 0)
	}
	kox := make([]float64, 10, 10)
	if move.Kox == -1 {
		kox[9] = 1
	} else {
		kox[move.Kox] = 1
	}
	input = append(input, kox...)
	koy := make([]float64, 9, 9)
	if move.Koy >= 0 {
		koy[move.Koy] = 1
	}
	input = append(input, move.ScoreWhite/81)
	input = append(input, move.ScoreBlack/81)
	input = append(input, move.CapturesBlack/81)
	input = append(input, move.CapturesWhite/81)

	x := make([]float64, 10, 10)
	if move.X == -1 {
		x[9] = 1
	} else {
		x[move.X] = 1
	}
	input = append(input, x...)
	y := make([]float64, 9, 9)
	if move.Y >= 0 {
		y[move.Y] = 1
	}
	input = append(input, y...)
	return input
}

func ConnectDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db, err
}

func CreateGame(g *game.Game, db *sql.DB) {
	var id int
	CreateBoard(&g.Board, db)
	q := fmt.Sprintf(`
		INSERT INTO game (whitescore, blackscore, turn, ended, winner, whitecaptures, blackcaptures, kox, koy, passed, boardid)
		VALUES (%v, %v, '%v', %v, '', %v, %v, %v, %v, %v, %v) RETURNING id;`,
		g.Score["white"], g.Score["black"], g.Turn, g.Ended, g.Captures["white"], g.Captures["black"], g.Ko[0], g.Ko[1], g.Passed, g.Board.ID)
	err := db.QueryRow(q).Scan(&id)
	if err != nil {
		fmt.Println("Exec err:", err.Error())
	}
	g.ID = id
}

func UpdateGame(g *game.Game, db *sql.DB) {
	winner := ""
	if g.Ended {
		winner = g.Winner
	}
	CreateBoard(&g.Board, db)

	q := fmt.Sprintf(`
	UPDATE game 
	SET whitescore = %v, blackscore = %v, turn = '%v', ended = %v, winner = '%v', whitecaptures = %v, blackcaptures = %v, kox = %v, koy = %v, passed = %v, boardid= %v
	WHERE id = %v
	`,
		g.Score["white"], g.Score["black"], g.Turn, g.Ended, winner, g.Captures["white"], g.Captures["black"], g.Ko[0], g.Ko[1], g.Passed, g.Board.ID, g.ID)

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
	rows, err := db.Query(q)
	if err != nil {
		fmt.Println(err.Error())
	}
	rows.Close()
}
