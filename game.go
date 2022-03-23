package main

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/rs/xid"
)

var board [][]point = generateBoard(boardSize)

const boardSize = 9

func main() {
	p0 := point{X: 2, Y: 5, Color: "white"}
	p0.assignGroup(groups)
	groups[p0.Group].addPoint(p0)
	board[p0.Y][p0.X] = p0

	fmt.Println(p0.adjPoints())

	p1 := point{X: 2, Y: 6, Color: "white"}
	p1.assignGroup(groups)
	groups[p1.Group].addPoint(p1)
	board[p1.Y][p1.X] = p1

	p2 := point{X: 1, Y: 5, Color: "black"}
	p2.assignGroup(groups)
	groups[p2.Group].addPoint(p2)
	board[p2.Y][p2.X] = p2

	// TODO: implement group logic to add point, calculate bounds, count liberties, merge adjacent groups, and capture

	fmt.Println(p1.adjPoints())
	fmt.Println(p0.adjPoints())
	for _, g := range groups {
		fmt.Println(*g, g.countLiberties(board))
	}

	printBoard(board)

}

var groups = map[string]*group{}

type group struct {
	ID     string
	Color  string
	Bounds [][2]int
}

// a "liberty" is an empty point adjacent to the group
func (g group) countLiberties(gameBoard [][]point) int {
	numOfLiberties := 0
	for _, b := range g.Bounds {
		p := gameBoard[b[1]][b[0]]
		if p.Color == "" {
			numOfLiberties++
		}
	}
	return numOfLiberties
}

func (g *group) addPoint(p point) {
	// add adjacent points to selected group, unless the point belongs to the group
	for _, adjP := range p.adjPoints() {
		bound := [2]int{adjP.X, adjP.Y}
		if adjP.Group != g.ID {
			g.Bounds = append(g.Bounds, bound)
		}
	}
	// remove new point from selected group
	pMatchesXY := func(b [2]int) bool { return b[0] == p.X && b[1] == p.Y }

	if i := slices.IndexFunc[[2]int](g.Bounds, pMatchesXY); i > -1 {
		g.Bounds = append(g.Bounds[:i], g.Bounds[i+1:]...)
	}
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

func (p *point) assignGroup(groupMap map[string]*group) {
	for _, adjPoint := range p.adjPoints() {
		if adjPoint.Color == p.Color {
			p.Group = adjPoint.Group
			return
		}
	}
	p.Group = xid.New().String()
	g := group{
		ID:     p.Group,
		Bounds: [][2]int{},
		Color:  p.Color,
	}
	groupMap[p.Group] = &g
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
