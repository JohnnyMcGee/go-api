package game

import (
	"github.com/rs/xid"
)

// TODO: implement user settings (board size, scoring style, and?)
// TODO: implement multiple concurrent games, multiple online players, AI single player mode
func NewGameBoard(size int) GameBoard {
	gbPoints := make([][]*Point, size, size)
	for y := 0; y < size; y++ {
		col := make([]*Point, size, size)
		for x := 0; x < size; x++ {
			p := Point{Color: "", GroupId: "", X: x, Y: y, Permit: map[string]bool{"black": true, "white": true}}
			col[x] = &p
		}
		gbPoints[y] = col
	}
	return GameBoard{
		points: gbPoints,
		Groups: map[string]*group{},
	}
}

func (b GameBoard) DeepCopy() GameBoard {
	bPointsCopy := make([][]*Point, b.size(), b.size())
	for i := range bPointsCopy {
		bPointsCopy[i] = make([]*Point, b.size(), b.size())
	}
	b.ForEachPoint(func(p *Point) {
		pCopy := *p
		bPointsCopy[p.Y][p.X] = &pCopy
	})
	bGroupsCopy := map[string]*group{}
	for id, g := range b.Groups {
		gCopy := g.DeepCopy()
		bGroupsCopy[id] = &gCopy
	}
	return GameBoard{
		points: bPointsCopy,
		Groups: bGroupsCopy,
	}
}

type GameBoard struct {
	points [][]*Point
	Groups map[string]*group
}

func (b GameBoard) At(x, y int) *Point {
	return b.points[y][x]
}

func (b GameBoard) size() int {
	return len(b.points)
}

func (b GameBoard) Points() [][]Point {
	boardPoints := make([][]Point, len(b.points), len(b.points))
	for j, col := range b.points {
		pCol := make([]Point, len(col), len(col))
		for i := range col {
			p := col[i]
			pCol[i] = *p
		}
		boardPoints[j] = pCol
	}
	return boardPoints
}

func (b *GameBoard) addPoint(p Point) {
	// no point can be played twice (unless captured)
	p.Permit = map[string]bool{"black": false, "white": false}
	// bind point to group (create new group if needed)
	p.assignGroup(*b)
	pointGroup := b.Groups[p.GroupId]
	pointGroup.addPoint(p, *b)
	// merge any overlapping groups into one
	pointsByGroup := map[string][]Point{}
	for _, adjPoint := range p.adjPoints(*b) {
		if adjPoint.Color == p.Color && adjPoint.GroupId != p.GroupId {
			pointsByGroup[adjPoint.GroupId] = append(pointsByGroup[adjPoint.GroupId], adjPoint)
		}
	}
	for groupId, points := range pointsByGroup {
		adjGroup := b.Groups[groupId]
		points = append(points, p)
		pointGroup.connectGroup(*adjGroup, *b, points...)
	}
	// add point to board
	*b.At(p.X, p.Y) = p
}

func (b *GameBoard) applyPermissions(ko [2]int) {
	// check for eyes and apply permissions to prevent suicide
	b.ForEachPoint(func(p *Point) {
		if p.isAnEye(*b) {
			p.Permit = p.calculateEyePermissions(*b)
		}
	})
	// apply ko rule
	if ko[0] >= 0 {
		b.At(ko[0], ko[1]).Permit = map[string]bool{"black": false, "white": false}
	}
}

func (b *GameBoard) doCaptures(friendlyColor string) (capturedPoints map[string][]Point) {
	enemyColor := OppositeColor(friendlyColor)

	captureGroupsByColor := func(color string) []*group {
		captured := []*group{}
		for _, group := range b.Groups {
			if group.Color == color && group.countLiberties(*b) < 1 {
				captured = append(captured, group)
			}
		}
		return captured
	}

	removeCapturedGroups := func(captured []*group) []Point {
		capturedPoints := []Point{}
		for _, g := range captured {
			for _, p := range g.Points {
				b.At(p.X, p.Y).Color = ""
				b.At(p.X, p.Y).GroupId = ""
				b.At(p.X, p.Y).Permit = map[string]bool{"black": true, "white": true}
				b.At(p.X, p.Y).Territory = ""
				capturedPoints = append(capturedPoints, *p)
			}
			delete(b.Groups, g.ID)
		}
		return capturedPoints
	}
	capturedGroups := make(map[string][]*group)
	capturedPoints = map[string][]Point{"black": {}, "white": {}}
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

func (b *GameBoard) ForEachPoint(f func(*Point)) {
	for _, row := range b.points {
		for _, p := range row {
			f(p)
		}
	}
}

func (b *GameBoard) Score() map[string]int {
	score := map[string]int{"black": 0, "white": 0}
	// count groups of enclosed free points

	territories := make(map[string]*group)

	b.ForEachPoint(func(p *Point) {
		// check territory above and to the left of point
		up := "none"
		left := "none"
		if p.Y > 0 {
			up = b.At(p.X, p.Y-1).Territory
		}
		if p.X > 0 {
			left = b.At(p.X-1, p.Y).Territory
		}

		// up & left are either a territory id, "none", or ""
		compareColors := func(tColor, color string) string {
			switch {
			case tColor == "":
				return color
			case tColor != color:
				return "both"
			default:
				return tColor
			}
		}
		if pIsTerritory := p.Color == ""; pIsTerritory {
			// determine which territory this point belongs to
			if _, upIsTerritory := territories[up]; upIsTerritory {
				p.Territory = up
				(territories[up].Points) = append(territories[up].Points, p)
			} else if _, leftIsTerritory := territories[left]; leftIsTerritory {
				p.Territory = left
				(territories[left].Points) = append(territories[left].Points, p)
			} else {
				p.Territory = xid.New().String()
				territories[p.Territory] = &group{
					ID:     p.Territory,
					Color:  "",
					Points: []*Point{p},
				}
			}
			t := territories[p.Territory]

			// confirm what color surrounds the territory
			if upHasColor := up == ""; upHasColor {
				t.Color = compareColors(t.Color, b.At(p.X, p.Y-1).Color)
			}
			if leftHasColor := left == ""; leftHasColor {
				t.Color = compareColors(t.Color, b.At(p.X-1, p.Y).Color)
			}

			// merge adjacent territory if necessary
			if leftIsTerritory := left != "" && left != "none" && left != t.ID; leftIsTerritory {
				leftTerritory := territories[left]
				if leftTerritory.Color != "" {
					t.Color = compareColors(t.Color, leftTerritory.Color)
				}
				t.Points = append(t.Points, leftTerritory.Points...)
				for _, p := range leftTerritory.Points {
					b.At(p.X, p.Y).Territory = t.ID
				}
				delete(territories, left)
			}
		} else { // point is occupied by a color
			score[p.Color]++
			if t, upIsTerritory := territories[up]; upIsTerritory {
				t.Color = compareColors(t.Color, p.Color)
			}
			if t, leftIsTerritory := territories[left]; leftIsTerritory {
				t.Color = compareColors(t.Color, p.Color)
			}
		}
	})

	for _, t := range territories {
		score[t.Color] += len(t.Points)
		// provide more useful information about territory
		for _, p := range t.Points {
			b.At(p.X, p.Y).Territory = t.Color
		}
	}

	return score
}

type group struct {
	ID     string   `json:"id"`
	Color  string   `json:"color"`
	Bounds [][2]int `json:"bounds"`
	Points []*Point `json:"points"`
}

func (g group) DeepCopy() group {
	gPointsCopy := make([]*Point, len(g.Points), len(g.Points))
	for i, p := range g.Points {
		pCopy := *p
		gPointsCopy[i] = &pCopy
	}
	g.Points = gPointsCopy
	return g
}

// a "liberty" is an empty point adjacent to the group
func (g group) countLiberties(board GameBoard) int {
	numOfLiberties := 0
	for _, b := range g.Bounds {
		p := board.At(b[0], b[1])
		if p.Color == "" {
			numOfLiberties++
		}
	}
	return numOfLiberties
}

// calculate number of stones (colored points) in a group
func (g group) size() int {
	return len(g.Points)
}

func (g *group) addPoint(p Point, board GameBoard) {
	g.Points = append(g.Points, &p)
	// add adjacent points to selected group, unless the point belongs to the group
	for _, adjP := range p.adjPoints(board) {

		bound := [2]int{adjP.X, adjP.Y}

		if adjP.GroupId != g.ID {
			g.Bounds = append(g.Bounds, bound)
		}
	}
	if g.size() > 1 {
		g.recalculateBounds(p)
	}
}

func (g *group) recalculateBounds(removePoints ...Point) {
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

func (g *group) connectGroup(newGroup group, board GameBoard, connection ...Point) {

	// copy the applicable bounds
	for _, bound := range newGroup.Bounds {
		if board.At(bound[0], bound[1]).GroupId != g.ID {
			g.Bounds = append(g.Bounds, bound)
		}
	}
	g.recalculateBounds(connection...)

	// transfer points
	g.Points = append(g.Points, newGroup.Points...)

	// update point GroupIds on board
	for _, p := range newGroup.Points {
		board.At(p.X, p.Y).GroupId = g.ID
	}
	// clean up unneeded group
	delete(board.Groups, newGroup.ID)
}

type Point struct {
	Color     string          `json:"color"`
	GroupId   string          `json:"group"`
	X         int             `json:"x"`
	Y         int             `json:"y"`
	Permit    map[string]bool `json:"permit"`
	Territory string          `json:"territory"`
}

func (p Point) adjPoints(board GameBoard) []Point {
	adjPoints := []Point{}
	// top
	if p.Y > 0 {
		adjPoints = append(adjPoints, *board.At(p.X, p.Y-1))
	}
	// right
	if p.X < board.size()-1 {
		adjPoints = append(adjPoints, *board.At(p.X+1, p.Y))
	}
	// bottom
	if p.Y < board.size()-1 {
		adjPoints = append(adjPoints, *board.At(p.X, p.Y+1))
	}
	// left
	if p.X > 0 {
		adjPoints = append(adjPoints, *board.At(p.X-1, p.Y))
	}
	return adjPoints
}

func (p *Point) assignGroup(board GameBoard) {
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
	board.Groups[p.GroupId] = &g
}

func (p Point) isAnEye(board GameBoard) bool {
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
func (p *Point) calculateEyePermissions(board GameBoard) map[string]bool {

	adjGroups := map[string][]*group{
		"black": {},
		"white": {},
	}

	for _, adjP := range p.adjPoints(board) {
		adjGroup := board.Groups[adjP.GroupId]
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

func OppositeColor(color string) string {
	if color == "white" {
		return "black"
	}
	return "white"
}

type Game struct {
	Board    GameBoard      `json:"board"`
	Captures map[string]int `json:"captures"`
	Score    map[string]int `json:"score"`
	Ko       [2]int         `json:"ko"`
	Turn     string         `json:"turn"`
	Passed   bool           `json:"passed"`
	Ended    bool           `json:"ended"`
	Winner   string         `json:"winner"`
}

func NewGame(boardSize int) Game {
	return Game{
		Board:    NewGameBoard(boardSize),
		Captures: map[string]int{"black": 0, "white": 0},
		Score:    map[string]int{"black": 0, "white": 0},
		Ko:       [2]int{-1, -1},
		Turn:     "black",
		Passed:   false,
		Ended:    false,
		Winner:   "",
	}
}

func (g Game) DeepCopy() Game {
	g.Board = g.Board.DeepCopy()
	return g
}

func (g *Game) IsValidMove(p Point) bool {
	inRangeXY := p.X < g.Board.size() && p.X >= 0 && p.Y < g.Board.size() && p.Y >= 0
	validColor := p.Color == g.Turn
	if !g.Ended && inRangeXY && validColor {
		return g.Board.At(p.X, p.Y).Permit[p.Color]
	}
	return false
}

func (g *Game) Play(p Point) (score map[string]int) {
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
		newGroup := board.Groups[board.At(p.X, p.Y).GroupId]
		newPointInDanger := newGroup.size() == 1 && newGroup.countLiberties(*board) == 1

		if newPointInDanger {
			koPoint := capturedPoints[OppositeColor(p.Color)][0]
			g.Ko = [2]int{koPoint.X, koPoint.Y}
		}
	}
	g.Score = board.Score()
	board.applyPermissions(g.Ko)

	g.Turn = OppositeColor(p.Color)
	g.Passed = false

	return g.Score
}

func (g *Game) Pass() {
	if g.Passed {
		g.Ended = true
		g.Turn = ""
		if g.Score["black"] > g.Score["white"] {
			g.Winner = "black"
		} else if g.Score["white"] > g.Score["black"] {
			g.Winner = "white"
		}
	} else {
		g.Passed = true
		g.Turn = OppositeColor(g.Turn)
	}
}

func (g *Game) Resign(color string) {
	g.Ended = true
	g.Winner = OppositeColor(color)
}
