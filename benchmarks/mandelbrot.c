#include <stdio.h>

#define WIDTH 80
#define HEIGHT 40
#define MAX_ITER 100
#define XMIN -2.5
#define XMAX 1.0
#define YMIN -1.0
#define YMAX 1.0

int mandelbrot(double cx, double cy) {
    double x = 0.0;
    double y = 0.0;
    int iter = 0;

    while (iter < MAX_ITER) {
        double x2 = x * x;
        double y2 = y * y;

        if (x2 + y2 > 4.0) {
            return iter;
        }

        double xtemp = x2 - y2 + cx;
        y = 2.0 * x * y + cy;
        x = xtemp;

        iter++;
    }

    return MAX_ITER;
}

char iter_to_char(int iter) {
    if (iter == MAX_ITER) return ' ';
    if (iter > 80) return '.';
    if (iter > 60) return ':';
    if (iter > 40) return '-';
    if (iter > 20) return '=';
    if (iter > 10) return '+';
    if (iter > 5) return '*';
    return '#';
}

int main() {
    printf("Rendering Mandelbrot Set...\n");
    printf("Size: %d x %d\n", WIDTH, HEIGHT);
    printf("Max iterations: %d\n\n", MAX_ITER);

    for (int row = 0; row < HEIGHT; row++) {
        for (int col = 0; col < WIDTH; col++) {
            double cx = XMIN + (XMAX - XMIN) * col / (double)WIDTH;
            double cy = YMIN + (YMAX - YMIN) * row / (double)HEIGHT;

            int iter = mandelbrot(cx, cy);
            char c = iter_to_char(iter);
            putchar(c);
        }
        putchar('\n');
    }

    printf("\nRendering complete!\n");
    return 0;
}
