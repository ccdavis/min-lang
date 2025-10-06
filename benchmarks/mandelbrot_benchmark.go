package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== Mandelbrot Performance Benchmark ===")
	fmt.Println()

	// Test 1: Standard resolution, high iterations
	fmt.Println("Test 1: 100x50 @ 500 iterations")
	const WIDTH1 = 100
	const HEIGHT1 = 50
	const MAX_ITER1 = 500

	totalIterations := 0

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

			totalIterations += iter
		}
	}

	fmt.Printf("Pixels calculated: %d\n", WIDTH1*HEIGHT1)
	fmt.Printf("Total iterations: %d\n", totalIterations)
	fmt.Printf("Average iterations per pixel: %d\n", totalIterations/(WIDTH1*HEIGHT1))
	fmt.Println()

	// Test 2: Detail zoom (more computation per pixel)
	fmt.Println("Test 2: Deep zoom @ 1000 iterations")
	const WIDTH2 = 60
	const HEIGHT2 = 30
	const MAX_ITER2 = 1000

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

			totalIterations += iter
		}
	}

	fmt.Printf("Pixels calculated: %d\n", WIDTH2*HEIGHT2)
	fmt.Printf("Total iterations: %d\n", totalIterations)
	fmt.Printf("Average iterations per pixel: %d\n", totalIterations/(WIDTH2*HEIGHT2))
	fmt.Println()

	// Test 3: Multiple frames (animation simulation)
	fmt.Println("Test 3: Multi-frame calculation (10 frames)")
	const FRAMES = 10
	const FRAME_WIDTH = 40
	const FRAME_HEIGHT = 20
	const FRAME_ITERS = 100

	totalPixels := 0

	for frame := 0; frame < FRAMES; frame++ {
		zoomFactor := 1.0 - float64(frame)*0.05

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

				totalPixels++
			}
		}
	}

	fmt.Printf("Frames calculated: %d\n", FRAMES)
	fmt.Printf("Total pixels: %d\n", totalPixels)
	fmt.Printf("Pixels per frame: %d\n", FRAME_WIDTH*FRAME_HEIGHT)
	fmt.Println()

	// Test 4: Stress test with very high iterations
	fmt.Println("Test 4: Stress test (single point @ 10000 iterations)")
	const STRESS_ITERS = 10000

	cx := -0.7
	cy := 0.0
	x := 0.0
	y := 0.0
	iter := 0

	for iter < STRESS_ITERS {
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

	fmt.Printf("Point: %.1f + %.1fi\n", cx, cy)
	fmt.Printf("Escaped at iteration: %d\n", iter)
	fmt.Println()

	fmt.Println("=== Benchmark Complete ===")
}
