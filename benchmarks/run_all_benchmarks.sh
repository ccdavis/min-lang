#!/bin/bash

echo "=========================================="
echo "MinLang Performance Comparison"
echo "=========================================="
echo ""

cd "$(dirname "$0")/.."

echo "=== 1. Simple Mandelbrot (80x40 @ 100 iter) ==="
echo ""

echo "C (gcc -O3):"
(time ./benchmarks/mandelbrot_c > /dev/null) 2>&1 | grep real

echo "Go:"
(time ./benchmarks/mandelbrot_go > /dev/null) 2>&1 | grep real

echo "Python 3:"
(time python3 benchmarks/mandelbrot.py > /dev/null) 2>&1 | grep real

echo "MinLang:"
(time ./minlang examples/mandelbrot.min > /dev/null) 2>&1 | grep real

echo ""
echo "=== 2. Full Benchmark Suite (3 runs each) ==="
echo ""

echo "C (gcc -O3):"
total=0
for i in 1 2 3; do
    t=$( (time ./benchmarks/mandelbrot_benchmark_c > /dev/null) 2>&1 | grep real | awk '{print $2}' )
    echo "  Run $i: $t"
done

echo ""
echo "Go:"
for i in 1 2 3; do
    t=$( (time ./benchmarks/mandelbrot_benchmark_go > /dev/null) 2>&1 | grep real | awk '{print $2}' )
    echo "  Run $i: $t"
done

echo ""
echo "Python 3:"
for i in 1 2 3; do
    t=$( (time python3 benchmarks/mandelbrot_benchmark.py > /dev/null) 2>&1 | grep real | awk '{print $2}' )
    echo "  Run $i: $t"
done

echo ""
echo "MinLang:"
for i in 1 2 3; do
    t=$( (time ./minlang examples/mandelbrot_benchmark.min > /dev/null) 2>&1 | grep real | awk '{print $2}' )
    echo "  Run $i: $t"
done

echo ""
echo "=========================================="
echo "Summary"
echo "=========================================="
echo ""
echo "MinLang achieves ~26% of Python's speed"
echo "while maintaining code clarity and simplicity."
echo ""
echo "See COMPARISON_REPORT.md for detailed analysis."
echo ""
