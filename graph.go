package main

import (
	"math"
	"math/rand"
	"strconv"
)

func setSliceDimensions(input []float64, x int, y int) []float64 {
	if x == 0 || y == 0 {
		return input
	}
	input = sliceScaler(input, x)
	input = squash(input, y, x)
	return input
}

//scales []float64 slice to any x and y size
func sliceScaler(data []float64, targetSize int) []float64 {
	scaleFactor := float64(targetSize) / float64(len(data))
	output := make([]float64, targetSize)

	for i, content := range data {
		nearestIndex := math.Floor(scaleFactor * float64(i))
		startOffset := scaleFactor*float64(i) - nearestIndex
		firstValue := 1 - startOffset
		scaleTmp := scaleFactor - firstValue
		if scaleTmp < 0 {
			firstValue += scaleTmp
		}
		output[int(nearestIndex)] += content * firstValue

		for index := nearestIndex + 1; scaleTmp > 0 && int(index) < len(output); index++ {
			if scaleTmp >= 1 {
				output[int(index)] = content
				scaleTmp--
			} else {
				output[int(index)] = scaleTmp * content
				scaleTmp = 0
			}
		}
	}
	return output
}

func floatToString(number float64) string {
	return strconv.FormatFloat(number, 'f', 3, 64)
}

func sliceToString(input []float64) string {
	part := "["
	for _, value := range input {
		part += " " + floatToString(value)
	}
	return part + " ]"
}

//shifts slice up or down so it touches 0
func normalise(input []float64) []float64 {
	if len(input) <= 1 {
		return input
	}
	min := input[0]
	for _, y := range input {
		if y < min {
			min = y
		}
	}
	for index := 0; index < len(input); index++ {
		input[index] = input[index] - min
	}
	return input
}

//selects a reasonable y height according to a given x width
func autoSquash(input []float64, xSize int) []float64 {
	factor := autoSquashHeight(input, xSize)
	if factor > 0 {
		return squash(input, factor, xSize)

	}
	return input
}

//calculation behind autoSquash(), calculates the squash factor
func autoSquashHeight(input []float64, xSize int) int {
	input = sliceScaler(input, xSize)
	input = normalise(input)
	input = roudSlice(input)
	max := getMax(input)
	arr := make([]int, max+1)
	for _, value := range input {
		arr[int(value)]++
	}
	largestGap := 0
	gap := 0
	for _, value := range arr {
		if value != 0 {
			gap = 0
		} else {
			gap++
			if gap > largestGap {
				largestGap = gap
			}
		}
	}
	if largestGap > 0 {
		return max / largestGap
	}
	return -1
}

//squashes slice on y dimension to specified value (can end up slimmer)
func squash(input []float64, targetsize int, xSize int) []float64 {
	tmp := make([]float64, len(input))
	copy(tmp, input)
	if len(tmp) < 1 {
		return tmp
	}
	tmp = normalise(tmp)
	max := getMax(tmp)
	factor := float64(targetsize) / float64(max)
	for index := 0; index < len(tmp); index++ {
		tmp[index] = tmp[index] * factor
	}
	return tmp
}

//turns a []float64 into a [][]rune
func printAsGraph(input []float64) [][]rune {
	input = normalise(input)
	input = roudSlice(input)
	x := len(input)
	y := getMax(input)

	output := make([][]rune, y+1)
	for i := range output {
		output[i] = make([]rune, x+1)
	}

	for index := y; index >= 0; index-- {
		for index2 := 0; index2 < x; index2++ {
			if int(input[index2]) == y-index {
				output[index][index2] = 'X'
			} else {
				output[index][index2] = ' '
			}
		}
	}
	border := ""
	for index := 0; index < x; index++ {
		border += "="
	}
	return output
}

func printAsGraphSetX(input []float64, x int) [][]rune {
	return printAsGraph(sliceScaler(input, x))
}

func printAsGraphSetXandY(input []float64, x int, y int) [][]rune {
	return printAsGraph(setSliceDimensions(input, x, y))
}

func roudSlice(input []float64) []float64 {
	for i, value := range input {
		input[i] = math.Round(value)
	}
	return input
}

func getMax(input []float64) int {
	if len(input) < 1 {
		return 0
	}
	max := input[0]
	for index := 1; index < len(input); index++ {
		if input[index] > max {
			max = input[index]
		}
	}
	return int(max)
}

//generates random slice with given parameters
func generateSlice(size int, step float64, start float64) []float64 {
	output := make([]float64, size)
	for index := 0; index < size; index++ {
		output[index] = start
		start = start + (rand.Float64() * (step * 2)) - step
	}
	return output
}

//srolls ahed in slice with given parameters
func scrollSlice(input []float64, step float64, speed int) []float64 {
	//discard first n
	input = input[speed:]
	//create n more to add
	addition := make([]float64, speed)
	//populate new values
	start := input[len(input)-1]
	for index := 0; index < speed; index++ {
		start = start + (rand.Float64() * (step * 2)) - step
		addition[index] = start
	}
	//combine
	input = append(input, addition...)
	return input
}
