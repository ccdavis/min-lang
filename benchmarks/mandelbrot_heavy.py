#!/usr/bin/env python3

def main():
    print("=== Heavy Mandelbrot Benchmark ===")
    print("This benchmark is designed to minimize startup time effects")
    print()

    # Test 1: Large resolution, high iteration count
    print("Test 1: 200x200 @ 1000 iterations")
    WIDTH1 = 200
    HEIGHT1 = 200
    MAX_ITER1 = 1000

    total_iterations = 0

    for row in range(HEIGHT1):
        for col in range(WIDTH1):
            cx = -2.5 + 3.5 * col / WIDTH1
            cy = -1.25 + 2.5 * row / HEIGHT1

            x = 0.0
            y = 0.0
            iter = 0

            while iter < MAX_ITER1:
                x2 = x * x
                y2 = y * y

                if x2 + y2 > 4.0:
                    break

                xtemp = x2 - y2 + cx
                y = 2.0 * x * y + cy
                x = xtemp

                iter += 1

            total_iterations += iter

    print(f"Pixels calculated: {WIDTH1 * HEIGHT1}")
    print(f"Total iterations: {total_iterations}")
    print(f"Average iterations per pixel: {total_iterations // (WIDTH1 * HEIGHT1)}")
    print()

    # Test 2: Very high iteration count (deep zoom)
    print("Test 2: 150x150 @ 2000 iterations (deep zoom)")
    WIDTH2 = 150
    HEIGHT2 = 150
    MAX_ITER2 = 2000

    ZOOM_X = -0.5
    ZOOM_Y = 0.0
    ZOOM_SIZE = 0.5

    total_iterations = 0

    for row in range(HEIGHT2):
        for col in range(WIDTH2):
            cx = ZOOM_X - ZOOM_SIZE + 2.0 * ZOOM_SIZE * col / WIDTH2
            cy = ZOOM_Y - ZOOM_SIZE + 2.0 * ZOOM_SIZE * row / HEIGHT2

            x = 0.0
            y = 0.0
            iter = 0

            while iter < MAX_ITER2:
                x2 = x * x
                y2 = y * y

                if x2 + y2 > 4.0:
                    break

                xtemp = x2 - y2 + cx
                y = 2.0 * x * y + cy
                x = xtemp

                iter += 1

            total_iterations += iter

    print(f"Pixels calculated: {WIDTH2 * HEIGHT2}")
    print(f"Total iterations: {total_iterations}")
    print(f"Average iterations per pixel: {total_iterations // (WIDTH2 * HEIGHT2)}")
    print()

    # Test 3: Multiple medium-resolution renders
    print("Test 3: 30 frames of 100x100 @ 500 iterations")
    FRAMES = 30
    FRAME_WIDTH = 100
    FRAME_HEIGHT = 100
    FRAME_ITERS = 500

    total_pixels = 0
    total_iterations = 0

    for frame in range(FRAMES):
        zoom_factor = 1.0 - frame * 0.02

        for row in range(FRAME_HEIGHT):
            for col in range(FRAME_WIDTH):
                cx = -2.0 * zoom_factor + 3.0 * zoom_factor * col / FRAME_WIDTH
                cy = -1.0 * zoom_factor + 2.0 * zoom_factor * row / FRAME_HEIGHT

                x = 0.0
                y = 0.0
                iter = 0

                while iter < FRAME_ITERS:
                    x2 = x * x
                    y2 = y * y

                    if x2 + y2 > 4.0:
                        break

                    xtemp = x2 - y2 + cx
                    y = 2.0 * x * y + cy
                    x = xtemp

                    iter += 1

                total_iterations += iter
                total_pixels += 1

    print(f"Frames calculated: {FRAMES}")
    print(f"Total pixels: {total_pixels}")
    print(f"Total iterations: {total_iterations}")
    print(f"Average iterations per pixel: {total_iterations // total_pixels}")
    print()

    print("=== Benchmark Complete ===")

if __name__ == "__main__":
    main()
