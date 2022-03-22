package main

var board [][]point = generateBoard(boardSize)

const boardSize = 9

type move struct {
	ID    int    `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

type point struct {
	Color string `json:"color"`
	Group uint   `json:"group"`
}

func generateBoard(size int) [][]point {
	var board = [][]point{}
	for y := 0; y < size; y++ {
		col := []point{}
		for x := 0; x < size; x++ {
			col = append(col, point{Color: "", Group: 0})
		}
		board = append(board, col)
	}
	return board
}
