#include <stdio.h>

int main() {
    printf("=== Heavy Mandelbrot Benchmark ===\n");
    printf("This benchmark is designed to minimize startup time effects\n\n");

    // Test 1: Large resolution, high iteration count
    printf("Test 1: 200x200 @ 1000 iterations\n");
    const int WIDTH1 = 200;
    const int HEIGHT1 = 200;
    const int MAX_ITER1 = 1000;

    long long totalIterations = 0;

    for (int row = 0; row < HEIGHT1; row++) {
        for (int col = 0; col < WIDTH1; col++) {
            double cx = -2.5 + 3.5 * col / (double)WIDTH1;
            double cy = -1.25 + 2.5 * row / (double)HEIGHT1;

            double x = 0.0;
            double y = 0.0;
            int iter = 0;

            while (iter < MAX_ITER1) {
                double x2 = x * x;
                double y2 = y * y;

                if (x2 + y2 > 4.0) {
                    break;
                }

                double xtemp = x2 - y2 + cx;
                y = 2.0 * x * y + cy;
                x = xtemp;

                iter++;
            }

            totalIterations += iter;
        }
    }

    printf("Pixels calculated: %d\n", WIDTH1 * HEIGHT1);
    printf("Total iterations: %lld\n", totalIterations);
    printf("Average iterations per pixel: %lld\n\n", totalIterations / (WIDTH1 * HEIGHT1));

    // Test 2: Very high iteration count (deep zoom)
    printf("Test 2: 150x150 @ 2000 iterations (deep zoom)\n");
    const int WIDTH2 = 150;
    const int HEIGHT2 = 150;
    const int MAX_ITER2 = 2000;

    const double ZOOM_X = -0.5;
    const double ZOOM_Y = 0.0;
    const double ZOOM_SIZE = 0.5;

    totalIterations = 0;

    for (int row = 0; row < HEIGHT2; row++) {
        for (int col = 0; col < WIDTH2; col++) {
            double cx = ZOOM_X - ZOOM_SIZE + 2.0 * ZOOM_SIZE * col / (double)WIDTH2;
            double cy = ZOOM_Y - ZOOM_SIZE + 2.0 * ZOOM_SIZE * row / (double)HEIGHT2;

            double x = 0.0;
            double y = 0.0;
            int iter = 0;

            while (iter < MAX_ITER2) {
                double x2 = x * x;
                double y2 = y * y;

                if (x2 + y2 > 4.0) {
                    break;
                }

                double xtemp = x2 - y2 + cx;
                y = 2.0 * x * y + cy;
                x = xtemp;

                iter++;
            }

            totalIterations += iter;
        }
    }

    printf("Pixels calculated: %d\n", WIDTH2 * HEIGHT2);
    printf("Total iterations: %lld\n", totalIterations);
    printf("Average iterations per pixel: %lld\n\n", totalIterations / (WIDTH2 * HEIGHT2));

    // Test 3: Multiple medium-resolution renders
    printf("Test 3: 30 frames of 100x100 @ 500 iterations\n");
    const int FRAMES = 30;
    const int FRAME_WIDTH = 100;
    const int FRAME_HEIGHT = 100;
    const int FRAME_ITERS = 500;

    int totalPixels = 0;
    totalIterations = 0;

    for (int frame = 0; frame < FRAMES; frame++) {
        double zoomFactor = 1.0 - frame * 0.02;

        for (int row = 0; row < FRAME_HEIGHT; row++) {
            for (int col = 0; col < FRAME_WIDTH; col++) {
                double cx = -2.0 * zoomFactor + 3.0 * zoomFactor * col / (double)FRAME_WIDTH;
                double cy = -1.0 * zoomFactor + 2.0 * zoomFactor * row / (double)FRAME_HEIGHT;

                double x = 0.0;
                double y = 0.0;
                int iter = 0;

                while (iter < FRAME_ITERS) {
                    double x2 = x * x;
                    double y2 = y * y;

                    if (x2 + y2 > 4.0) {
                        break;
                    }

                    double xtemp = x2 - y2 + cx;
                    y = 2.0 * x * y + cy;
                    x = xtemp;

                    iter++;
                }

                totalIterations += iter;
                totalPixels++;
            }
        }
    }

    printf("Frames calculated: %d\n", FRAMES);
    printf("Total pixels: %d\n", totalPixels);
    printf("Total iterations: %lld\n", totalIterations);
    printf("Average iterations per pixel: %lld\n\n", totalIterations / totalPixels);

    printf("=== Benchmark Complete ===\n");
    return 0;
}
