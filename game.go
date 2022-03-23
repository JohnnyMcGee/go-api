package main

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/rs/xid"
)

func main() {

	board.addPoint(point{X: 2, Y: 5, Color: "white"})
	board.addPoint(point{X: 2, Y: 6, Color: "white"})
	board.addPoint(point{X: 1, Y: 5, Color: "black"})

	// // TODO: implement group logic to add point, calculate bounds, count liberties, merge adjacent groups, and capture

	for _, g := range board.groups {
		fmt.Println(*g, g.countLiberties(board))
	}

	for _, row := range board.Points() {
		fmt.Println(row)
	}

}

var board = NewGameBoard(boardSize)

const boardSize = 9

func NewGameBoard(size int) gameBoard {
	gbPoints := make([][]*point, size, size)
	for y := 0; y < size; y++ {
		col := make([]*point, size, size)
		for x := 0; x < size; x++ {
			p := point{Color: "", Group: "", X: x, Y: y}
			col[x] = &p
		}
		gbPoints[y] = col
	}
	return gameBoard{
		points: gbPoints,
		groups: map[string]*group{},
	}
}

type gameBoard struct {
	points [][]*point
	groups map[string]*group
}

func (b gameBoard) at(x, y int) *point {
	return b.points[y][x]
}

func (b gameBoard) Points() [][]point {
	boardPoints := make([][]point, len(b.points), len(b.points))
	for j, col := range b.points {
		pCol := make([]point, len(col), len(col))
		for i := range col {
			p := col[i]
			pCol[i] = *p
		}
		boardPoints[j] = pCol
	}
	return boardPoints
}

func (b *gameBoard) addPoint(p point) {
	// assign group to point
	p.assignGroup(*b)
	group := b.groups[p.Group]
	// add point to group
	group.addPoint(p, *b)
	// add point to board
	*b.at(p.X, p.Y) = p
}

type group struct {
	ID     string
	Color  string
	Bounds [][2]int
}

// a "liberty" is an empty point adjacent to the group
func (g group) countLiberties(board gameBoard) int {
	numOfLiberties := 0
	for _, b := range g.Bounds {
		p := board.at(b[0], b[1])
		if p.Color == "" {
			numOfLiberties++
		}
	}
	return numOfLiberties
}

func (g *group) addPoint(p point, board gameBoard) {
	// add adjacent points to selected group, unless the point belongs to the group
	for _, adjP := range p.adjPoints(board) {
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

// TODO: change p.Group to p.GroupId
type point struct {
	Color string `json:"color"`
	Group string `json:"group"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

func (p point) adjPoints(board gameBoard) []point {
	adjPoints := []point{}
	// top
	if p.Y > 0 {
		adjPoints = append(adjPoints, *board.at(p.X, p.Y-1))
	}
	// right
	if p.X < boardSize-1 {
		adjPoints = append(adjPoints, *board.at(p.X+1, p.Y))
	}
	// bottom
	if p.Y < boardSize-1 {
		adjPoints = append(adjPoints, *board.at(p.X, p.Y+1))
	}
	// left
	if p.X > 0 {
		adjPoints = append(adjPoints, *board.at(p.X-1, p.Y))
	}
	return adjPoints
}

func (p *point) assignGroup(board gameBoard) {
	for _, adjPoint := range p.adjPoints(board) {
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
	board.groups[p.Group] = &g
}

func isValidMove(x int, y int, color string, board gameBoard) bool {
	inRangeXY := x < boardSize && x >= 0 && y < boardSize && y >= 0
	validColor := color == "white" || color == "black"
	pointIsOpen := board.at(x, y).Color == ""
	return inRangeXY && validColor && pointIsOpen
}
