# MinLang Performance Benchmarks

This directory contains performance comparison benchmarks between MinLang and mainstream languages (C, Go, Python).

## Files

### Benchmark Programs

| File | Language | Description |
|------|----------|-------------|
| `mandelbrot.c` | C | Simple Mandelbrot renderer |
| `mandelbrot.go` | Go | Simple Mandelbrot renderer |
| `mandelbrot.py` | Python | Simple Mandelbrot renderer |
| `mandelbrot_benchmark.c` | C | Full benchmark suite |
| `mandelbrot_benchmark.go` | Go | Full benchmark suite |
| `mandelbrot_benchmark.py` | Python | Full benchmark suite |

### MinLang Versions

MinLang benchmarks are in the `examples/` directory:
- `../examples/mandelbrot.min` - Simple renderer
- `../examples/mandelbrot_benchmark.min` - Full benchmark suite
- `../examples/fractal_explorer.min` - Advanced multi-fractal demo

### Reports

- `COMPARISON_REPORT.md` - Detailed performance analysis and comparison

## Building

### C Programs
```bash
gcc -O3 -o mandelbrot_c mandelbrot.c
gcc -O3 -o mandelbrot_benchmark_c mandelbrot_benchmark.c
```

### Go Programs
```bash
go build -o mandelbrot_go mandelbrot.go
go build -o mandelbrot_benchmark_go mandelbrot_benchmark.go
```

### Python Programs
Already executable (interpreted):
```bash
chmod +x mandelbrot.py mandelbrot_benchmark.py
```

## Running Benchmarks

### Quick Visual Test
```bash
# C
./mandelbrot_c

# Go
./mandelbrot_go

# Python
python3 mandelbrot.py

# MinLang
cd .. && ./minlang examples/mandelbrot.min
```

### Performance Benchmarks
```bash
# C
time ./mandelbrot_benchmark_c

# Go
time ./mandelbrot_benchmark_go

# Python
time python3 mandelbrot_benchmark.py

# MinLang
cd .. && time ./minlang examples/mandelbrot_benchmark.min
```

## Results Summary

**Quick Numbers (Benchmark Suite Average):**

| Language | Time | vs MinLang | vs C |
|----------|------|------------|------|
| C (gcc -O3) | 0.004s | 113.5× faster | 1.0× |
| Go | 0.006s | 75.7× faster | 1.5× slower |
| Python 3 | 0.116s | 3.9× faster | 29× slower |
| MinLang | 0.457s | 1.0× | 113.5× slower |

**Throughput (iterations/second):**
- C: ~470 million
- Go: ~313 million
- Python: ~16.2 million
- MinLang: ~4.1 million

**MinLang achieves 26% of Python's performance** - impressive for a simple pedagogical interpreter!

## Analysis

See `COMPARISON_REPORT.md` for detailed analysis including:
- Performance characteristics
- Optimization opportunities
- Use case recommendations
- Future improvement suggestions
