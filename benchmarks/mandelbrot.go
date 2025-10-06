package main

import (
	"fmt"
)

const (
	WIDTH    = 80
	HEIGHT   = 40
	MAX_ITER = 100
	XMIN     = -2.5
	XMAX     = 1.0
	YMIN     = -1.0
	YMAX     = 1.0
)

func mandelbrot(cx, cy float64) int {
	x := 0.0
	y := 0.0
	iter := 0

	for iter < MAX_ITER {
		x2 := x * x
		y2 := y * y

		if x2+y2 > 4.0 {
			return iter
		}

		xtemp := x2 - y2 + cx
		y = 2.0*x*y + cy
		x = xtemp

		iter++
	}

	return MAX_ITER
}

func iterToChar(iter int) string {
	if iter == MAX_ITER {
		return " "
	}
	if iter > 80 {
		return "."
	}
	if iter > 60 {
		return ":"
	}
	if iter > 40 {
		return "-"
	}
	if iter > 20 {
		return "="
	}
	if iter > 10 {
		return "+"
	}
	if iter > 5 {
		return "*"
	}
	return "#"
}

func main() {
	fmt.Println("Rendering Mandelbrot Set...")
	fmt.Printf("Size: %d x %d\n", WIDTH, HEIGHT)
	fmt.Printf("Max iterations: %d\n", MAX_ITER)
	fmt.Println()

	for row := 0; row < HEIGHT; row++ {
		line := ""
		for col := 0; col < WIDTH; col++ {
			cx := XMIN + (XMAX-XMIN)*float64(col)/float64(WIDTH)
			cy := YMIN + (YMAX-YMIN)*float64(row)/float64(HEIGHT)

			iter := mandelbrot(cx, cy)
			char := iterToChar(iter)
			line += char
		}
		fmt.Println(line)
	}

	fmt.Println()
	fmt.Println("Rendering complete!")
}
