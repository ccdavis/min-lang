# MinLang Performance Tests

This document summarizes the performance characteristics of MinLang using fractal rendering benchmarks.

## Test Programs

### 1. Mandelbrot Set Renderer (`mandelbrot.min`)
- **Resolution**: 80×40 pixels (3,200 pixels)
- **Max Iterations**: 100
- **Features Used**: Functions, nested loops, float arithmetic, string operations
- **Performance**: ~0.026s
- **Output**: ASCII art visualization of the Mandelbrot set

### 2. Performance Benchmark (`mandelbrot_benchmark.min`)
Comprehensive benchmark testing various scenarios:

#### Test 1: High Resolution
- 100×50 pixels @ 500 max iterations
- Total pixels: 5,000
- Total iterations: 461,020
- Average: 92 iterations/pixel

#### Test 2: Deep Zoom
- 60×30 pixels @ 1,000 max iterations
- Total pixels: 1,800
- Total iterations: 1,419,533
- Average: 788 iterations/pixel

#### Test 3: Multi-Frame
- 10 frames of 40×20 pixels
- Total pixels: 8,000
- Frame calculation with varying zoom levels

#### Test 4: Stress Test
- Single point @ 10,000 iterations
- Tests loop performance and float operations

**Total Performance**: ~0.451s for all tests combined

### 3. Fractal Explorer (`fractal_explorer.min`)
Advanced program demonstrating multiple language features:

- **Fractals**: Mandelbrot, Julia Set, Burning Ship
- **Resolution**: 3× 70×30 pixels (6,300 total pixels)
- **Max Iterations**: 100 per fractal
- **Features Used**:
  - Enums (FractalType)
  - Structs (Config, JuliaParams)
  - Multiple functions (5 calculation functions)
  - Nested loops
  - Float arithmetic
  - String concatenation
  - Conditional logic

**Performance**: ~0.046s (all three fractals)

## Performance Summary

| Metric | Value |
|--------|-------|
| Total pixels rendered | 11,500+ |
| Total iterations computed | 1,880,553+ |
| Execution time | <0.5s |
| Pixels/second | ~23,000+ |
| Iterations/second | ~3,761,106+ |

## Language Features Tested

✅ **Enums**: Type-safe fractal selection
✅ **Structs**: Configuration and parameter organization
✅ **Functions**: Modular fractal calculations (5+ functions)
✅ **Nested Loops**: Up to 3 levels deep
✅ **Float Arithmetic**: Complex number calculations
✅ **Conditionals**: Fractal type selection and rendering logic
✅ **String Operations**: Dynamic line building
✅ **Constants**: Configuration values
✅ **Break Statements**: Early loop termination for optimization

## Optimization Observations

1. **Break Statement Impact**: Using `break` when points escape reduces unnecessary iterations
2. **Float Performance**: Float arithmetic is reasonably fast for interpreted code
3. **String Concatenation**: Line-by-line rendering is efficient
4. **Function Calls**: Minimal overhead for function invocations
5. **VM Performance**: The bytecode VM handles tight loops well

## Comparison Notes

For reference, these benchmarks compute millions of iterations and render complex fractals in under half a second on modest hardware. This demonstrates that MinLang's bytecode VM architecture provides reasonable performance for computational tasks while maintaining simplicity and ease of implementation.

## Running the Benchmarks

```bash
# Quick visual test
./minlang examples/mandelbrot.min

# Performance benchmark
time ./minlang examples/mandelbrot_benchmark.min

# Feature demonstration
./minlang examples/fractal_explorer.min
```

## Future Optimization Opportunities

- JIT compilation for hot loops
- Integer-only fast path for iteration counting
- SIMD operations for parallel pixel processing
- Inline caching for struct field access
- Constant folding in the compiler
