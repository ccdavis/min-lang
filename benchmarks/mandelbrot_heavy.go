package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== Heavy Mandelbrot Benchmark ===")
	fmt.Println("This benchmark is designed to minimize startup time effects")
	fmt.Println()

	// Test 1: Large resolution, high iteration count
	fmt.Println("Test 1: 200x200 @ 1000 iterations")
	const WIDTH1 = 200
	const HEIGHT1 = 200
	const MAX_ITER1 = 1000

	totalIterations := int64(0)

	for row := 0; row < HEIGHT1; row++ {
		for col := 0; col < WIDTH1; col++ {
			cx := -2.5 + 3.5*float64(col)/float64(WIDTH1)
			cy := -1.25 + 2.5*float64(row)/float64(HEIGHT1)

			x := 0.0
			y := 0.0
			iter := 0

			for iter < MAX_ITER1 {
				x2 := x * x
				y2 := y * y

				if x2+y2 > 4.0 {
					break
				}

				xtemp := x2 - y2 + cx
				y = 2.0*x*y + cy
				x = xtemp

				iter++
			}

			totalIterations += int64(iter)
		}
	}

	fmt.Printf("Pixels calculated: %d\n", WIDTH1*HEIGHT1)
	fmt.Printf("Total iterations: %d\n", totalIterations)
	fmt.Printf("Average iterations per pixel: %d\n\n", totalIterations/(WIDTH1*HEIGHT1))

	// Test 2: Very high iteration count (deep zoom)
	fmt.Println("Test 2: 150x150 @ 2000 iterations (deep zoom)")
	const WIDTH2 = 150
	const HEIGHT2 = 150
	const MAX_ITER2 = 2000

	const ZOOM_X = -0.5
	const ZOOM_Y = 0.0
	const ZOOM_SIZE = 0.5

	totalIterations = 0

	for row := 0; row < HEIGHT2; row++ {
		for col := 0; col < WIDTH2; col++ {
			cx := ZOOM_X - ZOOM_SIZE + 2.0*ZOOM_SIZE*float64(col)/float64(WIDTH2)
			cy := ZOOM_Y - ZOOM_SIZE + 2.0*ZOOM_SIZE*float64(row)/float64(HEIGHT2)

			x := 0.0
			y := 0.0
			iter := 0

			for iter < MAX_ITER2 {
				x2 := x * x
				y2 := y * y

				if x2+y2 > 4.0 {
					break
				}

				xtemp := x2 - y2 + cx
				y = 2.0*x*y + cy
				x = xtemp

				iter++
			}

			totalIterations += int64(iter)
		}
	}

	fmt.Printf("Pixels calculated: %d\n", WIDTH2*HEIGHT2)
	fmt.Printf("Total iterations: %d\n", totalIterations)
	fmt.Printf("Average iterations per pixel: %d\n\n", totalIterations/(WIDTH2*HEIGHT2))

	// Test 3: Multiple medium-resolution renders
	fmt.Println("Test 3: 30 frames of 100x100 @ 500 iterations")
	const FRAMES = 30
	const FRAME_WIDTH = 100
	const FRAME_HEIGHT = 100
	const FRAME_ITERS = 500

	totalPixels := 0
	totalIterations = 0

	for frame := 0; frame < FRAMES; frame++ {
		zoomFactor := 1.0 - float64(frame)*0.02

		for row := 0; row < FRAME_HEIGHT; row++ {
			for col := 0; col < FRAME_WIDTH; col++ {
				cx := -2.0*zoomFactor + 3.0*zoomFactor*float64(col)/float64(FRAME_WIDTH)
				cy := -1.0*zoomFactor + 2.0*zoomFactor*float64(row)/float64(FRAME_HEIGHT)

				x := 0.0
				y := 0.0
				iter := 0

				for iter < FRAME_ITERS {
					x2 := x * x
					y2 := y * y

					if x2+y2 > 4.0 {
						break
					}

					xtemp := x2 - y2 + cx
					y = 2.0*x*y + cy
					x = xtemp

					iter++
				}

				totalIterations += int64(iter)
				totalPixels++
			}
		}
	}

	fmt.Printf("Frames calculated: %d\n", FRAMES)
	fmt.Printf("Total pixels: %d\n", totalPixels)
	fmt.Printf("Total iterations: %d\n", totalIterations)
	fmt.Printf("Average iterations per pixel: %d\n\n", totalIterations/int64(totalPixels))

	fmt.Println("=== Benchmark Complete ===")
}
