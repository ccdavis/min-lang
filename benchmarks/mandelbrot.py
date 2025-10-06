#!/usr/bin/env python3

WIDTH = 80
HEIGHT = 40
MAX_ITER = 100
XMIN = -2.5
XMAX = 1.0
YMIN = -1.0
YMAX = 1.0

def mandelbrot(cx, cy):
    x = 0.0
    y = 0.0
    iter = 0

    while iter < MAX_ITER:
        x2 = x * x
        y2 = y * y

        if x2 + y2 > 4.0:
            return iter

        xtemp = x2 - y2 + cx
        y = 2.0 * x * y + cy
        x = xtemp

        iter += 1

    return MAX_ITER

def iter_to_char(iter):
    if iter == MAX_ITER:
        return " "
    if iter > 80:
        return "."
    if iter > 60:
        return ":"
    if iter > 40:
        return "-"
    if iter > 20:
        return "="
    if iter > 10:
        return "+"
    if iter > 5:
        return "*"
    return "#"

def main():
    print("Rendering Mandelbrot Set...")
    print(f"Size: {WIDTH} x {HEIGHT}")
    print(f"Max iterations: {MAX_ITER}")
    print()

    for row in range(HEIGHT):
        line = ""
        for col in range(WIDTH):
            cx = XMIN + (XMAX - XMIN) * col / WIDTH
            cy = YMIN + (YMAX - YMIN) * row / HEIGHT

            iter = mandelbrot(cx, cy)
            char = iter_to_char(iter)
            line += char

        print(line)

    print()
    print("Rendering complete!")

if __name__ == "__main__":
    main()
