package main

import (
	"github.com/rs/xid"
)

// TODO: make sure all struct fields are capitalized for consistency
// func main() {
// 	board.addPoint(point{X: 2, Y: 5, Color: "white"})
// 	board.addPoint(point{X: 2, Y: 7, Color: "white"})
// 	board.addPoint(point{X: 3, Y: 6, Color: "white"})
// 	board.addPoint(point{X: 2, Y: 6, Color: "black"})

// 	for _, row := range board.Points() {
// 		fmt.Println(row)

// 	}
// 	fmt.Println()
// 	fmt.Println()

// 	// white captures black
// 	board.addPoint(point{X: 1, Y: 6, Color: "white"})
// 	capturedPoints := board.doCaptures()

// 	for _, row := range board.Points() {
// 		fmt.Println(row)
// 	}

// 	for _, g := range board.groups {
// 		fmt.Println(*g, g.countLiberties(board))
// 	}

// 	fmt.Println(capturedPoints)

// 	// TODO: implement group logic to capture

// 	// for _, row := range board.Points() {
// 	// 	fmt.Println(row)
// 	// }

// }

// var board = NewGameBoard(boardSize)

// const boardSize = 9

func NewGameBoard(size int) gameBoard {
	gbPoints := make([][]*point, size, size)
	for y := 0; y < size; y++ {
		col := make([]*point, size, size)
		for x := 0; x < size; x++ {
			p := point{Color: "", GroupId: "", X: x, Y: y}
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

func (b gameBoard) size() int {
	return len(b.points)
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
	// bind point to group (create new group if needed)
	p.assignGroup(*b)
	group := b.groups[p.GroupId]
	group.addPoint(p, *b)
	// merge any overlapping groups into one
	pointsByGroup := map[string][]point{}
	for _, adjPoint := range p.adjPoints(*b) {
		if adjPoint.Color == p.Color && adjPoint.GroupId != p.GroupId {
			pointsByGroup[adjPoint.GroupId] = append(pointsByGroup[adjPoint.GroupId], adjPoint)
		}
	}
	for groupId, points := range pointsByGroup {
		adjGroup := b.groups[groupId]
		points = append(points, p)
		group.connectGroup(*adjGroup, *b, points...)
	}
	// add point to board
	*b.at(p.X, p.Y) = p
}

func (b *gameBoard) doCaptures() map[string]int {
	captured := []string{}
	for _, group := range b.groups {
		if group.countLiberties(*b) < 1 {
			captured = append(captured, group.ID)
		}
	}
	// reset points on the board belonging to those groups
	// track the number of points captured
	capturedPoints := map[string]int{
		"black": 0,
		"white": 0,
	}
	for _, col := range b.points {
		for _, p := range col {
			for _, id := range captured {
				if id == p.GroupId {
					capturedPoints[p.Color]++
					p.Color, p.GroupId = "", ""
					break
				}
			}
		}
	}
	return capturedPoints
}

type group struct {
	ID     string   `json:"id"`
	Color  string   `json:"color"`
	Bounds [][2]int `json:"bounds"`
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
		if adjP.GroupId != g.ID {
			g.Bounds = append(g.Bounds, bound)
		}
	}
	g.removePointFromBounds(p)
}

func (g *group) removePointFromBounds(p point) {
	for i, bound := range g.Bounds {
		if bound[0] == p.X && bound[1] == p.Y {
			g.Bounds = append(g.Bounds[:i], g.Bounds[i+1:]...)
		}
	}
}

// for each additional group adjacent to point,
// merge into point group:

func (g *group) connectGroup(newGroup group, board gameBoard, connection ...point) {
	for _, bound := range newGroup.Bounds {
		if board.at(bound[0], bound[1]).GroupId != g.ID {
			g.Bounds = append(g.Bounds, bound)
		}
	}
	for _, connectPoint := range connection {
		g.removePointFromBounds(connectPoint)
	}
	delete(board.groups, newGroup.ID)
}

// clean up the partial groups

type point struct {
	Color   string `json:"color"`
	GroupId string `json:"group"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

func (p point) adjPoints(board gameBoard) []point {
	adjPoints := []point{}
	// top
	if p.Y > 0 {
		adjPoints = append(adjPoints, *board.at(p.X, p.Y-1))
	}
	// right
	if p.X < board.size()-1 {
		adjPoints = append(adjPoints, *board.at(p.X+1, p.Y))
	}
	// bottom
	if p.Y < board.size()-1 {
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
			p.GroupId = adjPoint.GroupId
			return
		}
	}
	p.GroupId = xid.New().String()
	g := group{
		ID:     p.GroupId,
		Bounds: [][2]int{},
		Color:  p.Color,
	}
	board.groups[p.GroupId] = &g
}

func isValidMove(p point, board gameBoard) bool {
	inRangeXY := p.X < board.size() && p.X >= 0 && p.Y < board.size() && p.Y >= 0
	validColor := p.Color == "white" || p.Color == "black"
	pointIsOpen := board.at(p.X, p.Y).Color == ""
	return inRangeXY && validColor && pointIsOpen
}
