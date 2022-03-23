package main

import (
	"github.com/rs/xid"
)

// TODO: make sure all struct fields are capitalized for consistency

func NewGameBoard(size int) gameBoard {
	gbPoints := make([][]*point, size, size)
	for y := 0; y < size; y++ {
		col := make([]*point, size, size)
		for x := 0; x < size; x++ {
			p := point{Color: "", GroupId: "", X: x, Y: y, Permit: map[string]bool{"black": true, "white": true}}
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
	pointGroup := b.groups[p.GroupId]
	pointGroup.addPoint(p, *b)
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
		pointGroup.connectGroup(*adjGroup, *b, points...)
	}
	// add point to board
	*b.at(p.X, p.Y) = p

	// check if any eyes exist in pointGroup (single free point enclosed on all sides)

	for _, row := range b.points {
		for _, p := range row {
			if p.isAnEye(*b) {
				p.calculateEyePermissions(*b)
			}
		}
	}
}

func (b *gameBoard) doCaptures(friendlyColor string) map[string]int {
	var enemyColor string
	if friendlyColor == "white" {
		enemyColor = "black"
	} else {
		enemyColor = "white"
	}

	captureGroupsByColor := func(color string) []string {
		captured := []string{}
		for _, group := range b.groups {
			if group.Color == color && group.countLiberties(*b) < 1 {
				captured = append(captured, group.ID)
			}
		}
		return captured
	}

	removeCapturedGroups := func(captured []string) int {
		capturedPoints := 0
		for _, col := range b.points {
			for _, p := range col {
				for _, id := range captured {
					if id == p.GroupId {
						capturedPoints++
						p.Color, p.GroupId = "", ""
						p.Permit = map[string]bool{
							"black": true,
							"white": true,
						}
						break
					}
				}
			}
		}
		for _, id := range captured {
			delete(board.groups, id)
		}
		return capturedPoints
	}
	capturedPoints := make(map[string]int)
	capturedEnemy := captureGroupsByColor(enemyColor)
	capturedPoints[enemyColor] = removeCapturedGroups(capturedEnemy)

	// capturing friendlies impossible unless suicide is enabled
	capturedFriendly := captureGroupsByColor(friendlyColor)
	capturedPoints[friendlyColor] = removeCapturedGroups(capturedFriendly)

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
	g.removeDuplicateBounds()
}

func (g *group) removePointFromBounds(p point) {
	for i, bound := range g.Bounds {
		if bound[0] == p.X && bound[1] == p.Y {
			g.Bounds = append(g.Bounds[:i], g.Bounds[i+1:]...)
		}
	}
}

func (g *group) removeDuplicateBounds() {
	// filter non-unique bounds
	uniqueBounds := [][2]int{}

	for _, bound := range g.Bounds {
		isUnique := true
		for _, uniqueBound := range uniqueBounds {
			xMatch := uniqueBound[0] == bound[0]
			yMatch := uniqueBound[1] == bound[1]
			if xMatch && yMatch {
				isUnique = false
			}
		}
		if isUnique {
			uniqueBounds = append(uniqueBounds, bound)
		}
	}

	g.Bounds = uniqueBounds
}

func (g *group) connectGroup(newGroup group, board gameBoard, connection ...point) {

	// copy the applicable bounds
	for _, bound := range newGroup.Bounds {
		if board.at(bound[0], bound[1]).GroupId != g.ID {
			g.Bounds = append(g.Bounds, bound)
		}
	}
	for _, connectPoint := range connection {
		g.removePointFromBounds(connectPoint)
	}

	g.removeDuplicateBounds()

	// update point GroupIds on board
	for _, row := range board.points {
		for _, p := range row {
			if p.GroupId == newGroup.ID {
				p.GroupId = g.ID
			}
		}
	}
	// clean up unneeded group
	delete(board.groups, newGroup.ID)
}

type point struct {
	Color   string          `json:"color"`
	GroupId string          `json:"group"`
	X       int             `json:"x"`
	Y       int             `json:"y"`
	Permit  map[string]bool `json:"permit"`
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

func (p point) isAnEye(board gameBoard) bool {
	isAnEye := true

	for _, adjP := range p.adjPoints(board) {
		if adjP.GroupId == "" {
			isAnEye = false
		}
	}
	return isAnEye
}

// assumes point is an eye (has no open point on any side)
func (p *point) calculateEyePermissions(board gameBoard) {

	adjGroups := map[string][]*group{
		"black": {},
		"white": {},
	}

	for _, adjP := range p.adjPoints(board) {
		adjGroup := board.groups[adjP.GroupId]
		adjGroups[adjGroup.Color] = append(adjGroups[adjGroup.Color], adjGroup)
	}
	// if point is an eye, determine its play permissions (play inside eye can be suicide)
	anyWhiteSingleLiberty := false
	anyWhiteMultiLiberty := false
	for _, wGroup := range adjGroups["white"] {
		liberties := wGroup.countLiberties(board)
		if liberties == 1 {
			anyWhiteSingleLiberty = true
		}
		if liberties > 1 {
			anyWhiteMultiLiberty = true
		}
	}

	anyBlackSingleLiberty := false
	anyBlackMultiLiberty := false
	for _, wGroup := range adjGroups["black"] {
		liberties := wGroup.countLiberties(board)
		if liberties == 1 {
			anyBlackSingleLiberty = true
		}
		if liberties > 1 {
			anyBlackMultiLiberty = true
		}
	}

	whitePermitted := anyBlackSingleLiberty || anyWhiteMultiLiberty
	blackPermitted := anyWhiteSingleLiberty || anyBlackMultiLiberty

	p.Permit = map[string]bool{"white": whitePermitted, "black": blackPermitted}
}

func isValidMove(p point, board gameBoard) bool {
	inRangeXY := p.X < board.size() && p.X >= 0 && p.Y < board.size() && p.Y >= 0
	validColor := p.Color == "white" || p.Color == "black"
	pointIsOpen := board.at(p.X, p.Y).Color == ""
	return inRangeXY && validColor && pointIsOpen
}
