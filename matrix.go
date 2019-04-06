package main

import (
	"strings"
)

func textToMatrix(input string) [][]rune {
	//takes text and returns it as rune matrix
	lines := strings.Split(input, "\n")
	max := 0
	for _, line := range lines {
		if len(line) > max {
			max = len(line)
		}
	}
	output := make([][]rune, len(lines))
	for i, line := range lines {
		output[i] = make([]rune, max)
		for j, r := range line {
			output[i][j] = r
		}
	}
	return output
}

//combines 2 matrix, each matrix can be offset, matrix2 will overwrite matrix1 if they overlap
func combineMatrix(input1 [][]rune, x1 int, y1 int, input2 [][]rune, x2 int, y2 int) [][]rune {
	m1y := len(input1)
	m1x := len(input1[0])
	m2y := len(input2)
	m2x := len(input2[0])

	maxX := m2x + x2
	maxY := m2y + y2
	if m1x+x1 > m2x+x2 {
		maxX = m1x + x1
	}
	if m1y+y1 > m2y+y2 {
		maxY = m1y + y1
	}

	//create base
	output := make([][]rune, maxY)
	for y := 0; y < len(output); y++ {
		output[y] = make([]rune, maxX)
	}
	//write matrix1
	for y := 0; y < m1y; y++ {
		for x := 0; x < m1x; x++ {
			output[y+y1][x+x1] = input1[y][x]
		}
	}
	//write matrix2
	for y := 0; y < m2y; y++ {
		for x := 0; x < m2x; x++ {
			val := input2[y][x]
			output[y+y2][x+x2] = val
		}
	}
	return output
}
