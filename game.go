package main

import (
	"github.com/rs/xid"
)

// TODO: implement end of game and user settings (board size, scoring style, and?)
// TODO: optimize speed of moves/scoring using concurrency (and improved algorithms for less iteration?)
// TODO: implement multiple concurrent games, multiple online players, AI single player mode
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
	// no point can be played twice (unless captured)
	p.Permit = map[string]bool{"black": false, "white": false}
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
}

func (b *gameBoard) applyPermissions(ko [2]int) {
	// check for eyes and apply permissions to prevent suicide
	b.forEachPoint(func(p *point) {
		if p.isAnEye(*b) {
			p.Permit = p.calculateEyePermissions(*b)
		}
	})
	// apply ko rule
	if ko[0] >= 0 {
		b.at(ko[0], ko[1]).Permit = map[string]bool{"black": false, "white": false}
	}
}

func (b *gameBoard) doCaptures(friendlyColor string) (capturedPoints map[string][]point) {
	enemyColor := oppositeColor(friendlyColor)

	captureGroupsByColor := func(color string) []*group {
		captured := []*group{}
		for _, group := range b.groups {
			if group.Color == color && group.countLiberties(*b) < 1 {
				captured = append(captured, group)
			}
		}
		return captured
	}

	removeCapturedGroups := func(captured []*group) []point {
		capturedPoints := []point{}
		for _, g := range captured {
			for _, p := range g.Points {
				b.at(p.X, p.Y).Color = ""
				b.at(p.X, p.Y).GroupId = ""
				b.at(p.X, p.Y).Permit = map[string]bool{"black": true, "white": true}
				capturedPoints = append(capturedPoints, *p)
			}
			delete(b.groups, g.ID)
		}
		return capturedPoints
	}
	capturedGroups := make(map[string][]*group)
	capturedPoints = map[string][]point{"black": {}, "white": {}}
	capturedGroups[enemyColor] = captureGroupsByColor(enemyColor)
	if len(capturedGroups[enemyColor]) >= 1 {
		capturedPoints[enemyColor] = removeCapturedGroups(capturedGroups[enemyColor])
	}

	// capturing friendlies impossible unless suicide is enabled
	capturedGroups[friendlyColor] = captureGroupsByColor(friendlyColor)
	if len(capturedGroups[friendlyColor]) >= 1 {
		capturedPoints[friendlyColor] = removeCapturedGroups(capturedGroups[friendlyColor])
	}

	return capturedPoints
}
func (b *gameBoard) forEachPoint(f func(*point)) {
	for _, row := range b.points {
		for _, p := range row {
			f(p)
		}
	}
}

func (b *gameBoard) Score() map[string]int {
	score := map[string]int{"black": 0, "white": 0}
	// count groups of enclosed free points

	territories := make(map[string]*group)

	boardBuffer := make([][]map[string]string, b.size(), b.size())
	for i := 0; i < b.size(); i++ {
		boardBuffer[i] = make([]map[string]string, b.size(), b.size())
	}

	// logic to determine which color owns an territory based on its neighboring points
	compareColors := func(territoryColor, neighborColor string) (newterritoryColor string) {
		if territoryColor != neighborColor && neighborColor != "" {
			if territoryColor == "" {
				return neighborColor
			} else {
				return "both"
			}
		}
		return territoryColor
	}

	// overwrite all bufferPoint points of a given territory with a new value
	updateterritory := func(territory string, newValue map[string]string) {
		e := territories[territory]
		for _, p := range e.Points {
			boardBuffer[p.Y][p.X] = newValue
		}
		e.Color = newValue["tColor"]
	}

	b.forEachPoint(func(p *point) {
		// check territory above and to the left of point
		var up map[string]string
		var left map[string]string
		if p.Y > 0 {
			up = boardBuffer[p.Y-1][p.X]
		} else {
			up = map[string]string{"tId": "none", "tColor": ""}
		}
		if p.X > 0 {
			left = boardBuffer[p.Y][p.X-1]
		} else {
			left = map[string]string{"tId": "none", "tColor": ""}
		}

		// determine the territory this bufferPoint point belongs to
		bufferPoint := make(map[string]string)
		if p.Color == "" {
			switch {
			case up["tId"] != "none":
				bufferPoint = up
				e := territories[up["tId"]]
				e.Points = append(e.Points, p)
			case left["tId"] != "none":
				bufferPoint = left
				e := territories[left["tId"]]
				e.Points = append(e.Points, p)
			default:
				bufferPoint["tId"] = xid.New().String()
				bufferPoint["tColor"] = ""
				territories[bufferPoint["tId"]] = &group{
					ID:     bufferPoint["tId"],
					Color:  bufferPoint["tColor"],
					Points: []*point{p},
				}
			}
			bufferPoint["tColor"] = compareColors(bufferPoint["tColor"], up["tColor"])
			bufferPoint["tColor"] = compareColors(bufferPoint["tColor"], left["tColor"])

			// merge adjacent territories as necessary
			if left["tId"] != bufferPoint["tId"] && left["tId"] != "none" {
				updateterritory(left["tId"], bufferPoint)
				territories[bufferPoint["tId"]].Points = append(territories[bufferPoint["tId"]].Points, territories[left["tId"]].Points...)
				delete(territories, left["tId"])
			}

			// update territories map with findings
			updateterritory(bufferPoint["tId"], bufferPoint)

		} else {
			score[p.Color]++
			bufferPoint = map[string]string{"tId": "none", "tColor": p.Color}
			// update the color of territories up and left
			updateNeighborColor := func(neighbor map[string]string) {
				if neighbor["tId"] != "none" {

					newNeighborColor := compareColors(neighbor["tColor"], p.Color)
					if neighbor["tColor"] != newNeighborColor {
						updateterritory(
							neighbor["tId"],
							map[string]string{"tId": neighbor["tId"], "tColor": newNeighborColor},
						)
					}
				}
			}
			updateNeighborColor(up)
			updateNeighborColor(left)
		}
		boardBuffer[p.Y][p.X] = bufferPoint
	})

	for _, e := range territories {
		score[e.Color] += len(e.Points)
	}

	return score
}

type group struct {
	ID     string   `json:"id"`
	Color  string   `json:"color"`
	Bounds [][2]int `json:"bounds"`
	Points []*point `json:"points"`
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

// calculate number of stones (colored points) in a group
func (g group) size(b gameBoard) int {
	return len(g.Points)
}

func (g *group) addPoint(p point, board gameBoard) {
	g.Points = append(g.Points, &p)
	// add adjacent points to selected group, unless the point belongs to the group
	for _, adjP := range p.adjPoints(board) {

		bound := [2]int{adjP.X, adjP.Y}

		if adjP.GroupId != g.ID {
			g.Bounds = append(g.Bounds, bound)
		}
	}
	if g.size(board) > 1 {
		g.recalculateBounds(p)
	}
}

func (g *group) recalculateBounds(removePoints ...point) {
	previouslyEncountered := make(map[[2]int]bool)
	uniqueBounds := make([][2]int, 0)
OUTER:
	for _, bound := range g.Bounds {
		// remove bound if it contains a remove point
		for _, p := range removePoints {
			if bound[0] == p.X && bound[1] == p.Y {
				// g.Bounds = append(g.Bounds[:i], g.Bounds[i+1:]...)
				continue OUTER
			}
		}

		// filter bound if it is not unique
		if _, ok := previouslyEncountered[bound]; !ok {
			previouslyEncountered[bound] = ok
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
	g.recalculateBounds(connection...)

	// transfer points
	g.Points = append(g.Points, newGroup.Points...)

	// update point GroupIds on board
	for _, p := range newGroup.Points {
		board.at(p.X, p.Y).GroupId = g.ID
	}
	// clean up unneeded group
	delete(board.groups, newGroup.ID)
}

type point struct {
	Color     string          `json:"color"`
	GroupId   string          `json:"group"`
	X         int             `json:"x"`
	Y         int             `json:"y"`
	Permit    map[string]bool `json:"permit"`
	Territory string          `json:"territory"`
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
	if p.Color != "" {
		return false
	}
	for _, adjP := range p.adjPoints(board) {
		if adjP.GroupId == "" {
			return false
		}
	}
	return true
}

// assumes point is an eye (has no open point on any side)
func (p *point) calculateEyePermissions(board gameBoard) map[string]bool {

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

	return map[string]bool{"white": whitePermitted, "black": blackPermitted}
}

func oppositeColor(color string) string {
	if color == "white" {
		return "black"
	}
	return "white"
}

type game struct {
	Board    gameBoard      `json:"board"`
	Captures map[string]int `json:"captures"`
	Score    map[string]int `json:"score"`
	Ko       [2]int         `json:"ko"`
	Turn     string         `json:"turn"`
}

func NewGame(boardSize int) game {
	return game{
		Board:    NewGameBoard(boardSize),
		Captures: map[string]int{"black": 0, "white": 0},
		Score:    map[string]int{"black": 0, "white": 0},
		Ko:       [2]int{-1, -1},
		Turn:     "black",
	}
}

func (g *game) isValidMove(p point) bool {
	inRangeXY := p.X < g.Board.size() && p.X >= 0 && p.Y < g.Board.size() && p.Y >= 0
	validColor := p.Color == g.Turn
	playIsPermitted := g.Board.at(p.X, p.Y).Permit[p.Color]
	return inRangeXY && validColor && playIsPermitted
}

func (g *game) play(p point) (score map[string]int) {
	board := &g.Board
	board.addPoint(p)
	capturedPoints := board.doCaptures(p.Color)
	for clr, points := range capturedPoints {
		(g.Captures)[clr] += len(points)
	}
	// apply the rule of ko:
	// A move may not revert the board back to its previous state
	g.Ko = [2]int{-1, -1}
	singlePointCaptured := len(capturedPoints["white"])+len(capturedPoints["black"]) == 1
	if singlePointCaptured {
		newGroup := board.groups[board.at(p.X, p.Y).GroupId]
		newPointInDanger := newGroup.size(*board) == 1 && newGroup.countLiberties(*board) == 1

		if newPointInDanger {
			koPoint := capturedPoints[oppositeColor(p.Color)][0]
			g.Ko = [2]int{koPoint.X, koPoint.Y}
		}
	}
	g.Score = board.Score()
	board.applyPermissions(g.Ko)

	g.Turn = oppositeColor(p.Color)

	return g.Score
}
