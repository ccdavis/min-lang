#include <stdio.h>

int main() {
    printf("=== Mandelbrot Performance Benchmark ===\n\n");

    // Test 1: Standard resolution, high iterations
    printf("Test 1: 100x50 @ 500 iterations\n");
    const int WIDTH1 = 100;
    const int HEIGHT1 = 50;
    const int MAX_ITER1 = 500;

    int totalIterations = 0;

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
    printf("Total iterations: %d\n", totalIterations);
    printf("Average iterations per pixel: %d\n\n", totalIterations / (WIDTH1 * HEIGHT1));

    // Test 2: Detail zoom
    printf("Test 2: Deep zoom @ 1000 iterations\n");
    const int WIDTH2 = 60;
    const int HEIGHT2 = 30;
    const int MAX_ITER2 = 1000;

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
    printf("Total iterations: %d\n", totalIterations);
    printf("Average iterations per pixel: %d\n\n", totalIterations / (WIDTH2 * HEIGHT2));

    // Test 3: Multiple frames
    printf("Test 3: Multi-frame calculation (10 frames)\n");
    const int FRAMES = 10;
    const int FRAME_WIDTH = 40;
    const int FRAME_HEIGHT = 20;
    const int FRAME_ITERS = 100;

    int totalPixels = 0;

    for (int frame = 0; frame < FRAMES; frame++) {
        double zoomFactor = 1.0 - frame * 0.05;

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

                totalPixels++;
            }
        }
    }

    printf("Frames calculated: %d\n", FRAMES);
    printf("Total pixels: %d\n", totalPixels);
    printf("Pixels per frame: %d\n\n", FRAME_WIDTH * FRAME_HEIGHT);

    // Test 4: Stress test
    printf("Test 4: Stress test (single point @ 10000 iterations)\n");
    const int STRESS_ITERS = 10000;

    double cx = -0.7;
    double cy = 0.0;
    double x = 0.0;
    double y = 0.0;
    int iter = 0;

    while (iter < STRESS_ITERS) {
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

    printf("Point: %.1f + %.1fi\n", cx, cy);
    printf("Escaped at iteration: %d\n\n", iter);

    printf("=== Benchmark Complete ===\n");
    return 0;
}
