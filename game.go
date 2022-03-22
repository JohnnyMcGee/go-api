package main

import (
	"fmt"

	"github.com/rs/xid"
)

var board [][]point = generateBoard(boardSize)

const boardSize = 9

func main() {
	p0 := point{X: 2, Y: 5, Color: "white"}
	// p0 = placePoint(p0)
	p0.assignGroup(groups)
	board[p0.Y][p0.X] = p0

	fmt.Println(p0.adjPoints())

	p1 := point{X: 2, Y: 6, Color: "white"}
	// p1 = placePoint(p1)
	p1.assignGroup(groups)
	board[p1.Y][p1.X] = p1

	fmt.Println(p1.adjPoints())
	fmt.Println(p0.adjPoints())
	fmt.Println(groups)

}

var groups = map[string]group{}

type group struct {
	ID     string
	Color  string
	Bounds []point
}

type point struct {
	Color string `json:"color"`
	Group string `json:"group"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

func (p point) adjPoints() []point {
	adjPoints := []point{}
	// top
	if p.Y > 0 {
		adjPoints = append(adjPoints, board[p.Y-1][p.X])
	}
	// right
	if p.X < boardSize-1 {
		adjPoints = append(adjPoints, board[p.Y][p.X+1])
	}
	// bottom
	if p.Y < boardSize-1 {
		adjPoints = append(adjPoints, board[p.Y+1][p.X])
	}
	// left
	if p.X > 0 {
		adjPoints = append(adjPoints, board[p.Y][p.X-1])
	}
	return adjPoints
}

func (p *point) assignGroup(groupMap map[string]group) {
	for _, adjPoint := range p.adjPoints() {
		if adjPoint.Color == p.Color {
			p.Group = adjPoint.Group
			return
		}
	}
	p.Group = xid.New().String()
	groupMap[p.Group] = group{
		ID:     p.Group,
		Bounds: []point{},
		Color:  p.Color,
	}
}

func generateBoard(size int) [][]point {
	var board = [][]point{}
	for y := 0; y < size; y++ {
		col := []point{}
		for x := 0; x < size; x++ {
			col = append(col, point{Color: "", Group: "", X: x, Y: y})
		}
		board = append(board, col)
	}
	return board
}

func printBoard(b [][]point) {
	for _, col := range b {
		fmt.Println(col)
	}
}

// func isInvalidMove(newMove move) bool {
// 	outOfRangeXY := newMove.X >= boardSize || newMove.X < 0 || newMove.Y >= boardSize || newMove.Y < 0
// 	invalidColor := newMove.Color != "white" && newMove.Color != "black"
// 	pointUnavailable := board[newMove.Y][newMove.X].Color != ""
// 	return outOfRangeXY || invalidColor || pointUnavailable
// }
