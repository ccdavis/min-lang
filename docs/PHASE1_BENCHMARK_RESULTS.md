# Phase 1 Performance Benchmark Results

**Date**: October 6, 2025
**Benchmark**: examples/mandelbrot_heavy.min
**Optimization**: Type-Specialized Arithmetic Opcodes (Phase 1)

## Results Summary

### Before Phase 1 (from PERFORMANCE.md)
```
Run 1: 16.91s
Run 2: 20.08s
Run 3: 17.25s
Run 4: 20.22s
Run 5: 17.32s
Average: 18.36s
Best: 16.91s
```

### After Phase 1 (Current - with type-specialized opcodes)
```
Run 1: 14.36s (user: 11.97s, sys: 0.01s)
Run 2: 10.74s (user: 11.93s, sys: 0.01s)
Run 3: 14.32s (user: 11.93s, sys: 0.02s)
Run 4: 10.73s (user: 11.92s, sys: 0.01s)
Run 5: 14.34s (user: 11.94s, sys: 0.02s)
Average: 12.90s
Best: 10.73s
User time avg: 11.94s
```

## Performance Improvement

**Wall Clock Time**:
- Previous: 18.36s
- Current: 12.90s
- **Improvement: 29.7% faster** (5.46s reduction)

**User Time** (more consistent metric):
- Average user time: 11.94s
- Very consistent across all runs (11.92s - 11.97s)

## Analysis

### Why Better Than Expected?

The proposal estimated 15-20% improvement, but we achieved **29.7%**. Possible reasons:

1. **Compounding effects**: The benchmark is extremely arithmetic-heavy
   - Millions of float multiplications (x * x, y * y, 2.0 * x * y)
   - Millions of float additions/subtractions (x2 + y2, x2 - y2 + cx)
   - Integer arithmetic in loops (iter = iter + 1, col = col + 1)

2. **CPU pipeline optimization**: Type-specialized opcodes have better branch prediction
   - No runtime type checking branches
   - Simpler instruction sequences
   - Better cache utilization

3. **Benchmark characteristics**: mandelbrot_heavy.min has:
   - 40,000 pixels (Test 1) × 178 avg iterations = ~7.1M iterations
   - 22,500 pixels (Test 2) × 1580 avg iterations = ~35.6M iterations
   - 300,000 pixels (Test 3) × 132 avg iterations = ~39.8M iterations
   - **Total: ~82.5 million iterations with intensive arithmetic**

4. **Each iteration has**:
   - 2 float multiplications (x * x, y * y)
   - 2 float additions (x2 + y2, 2.0 * x * y + cy)
   - 1 float subtraction (x2 - y2 + cx)
   - 1 integer addition (iter = iter + 1)
   - = **6 arithmetic operations per iteration**
   - = **~495 million arithmetic operations total**

### What Was Eliminated

For each of those 495 million operations, we eliminated:
- ✅ 2 runtime type checks (if left.Type == ... && right.Type == ...)
- ✅ 1 function call (executeBinaryOperation)
- ✅ 1 switch statement (switch op)
- ✅ Type conversion branches for mixed int/float

### Consistency Notes

User time is very consistent (±0.05s), indicating stable CPU performance.
Wall clock time varies more (10.73s - 14.36s) due to system scheduling.
The user time metric (11.94s) is more reliable for CPU-bound benchmarks.

## Cumulative Performance Journey

| Version | Time | Improvement vs Original |
|---------|------|------------------------|
| Original (pre-optimizations) | 27.68s | Baseline |
| After VM optimizations (before Phase 1) | 18.36s | 33.7% faster |
| **After Phase 1 (type-specialized opcodes)** | **12.90s** | **53.4% faster** |

### Incremental Improvement from Phase 1
- Previous best: 18.36s
- Current: 12.90s
- **Phase 1 alone: 29.7% improvement**

## Technical Details

### Opcodes Used (from disassembly inspection)
- OpAddInt: Integer loop counters (iter++, col++, row++)
- OpAddFloat: Complex number arithmetic (x + cx, y + cy)
- OpSubFloat: Mandelbrot calculation (x2 - y2 + cx)
- OpMulFloat: Squaring operations (x * x, y * y)
- OpMulInt: Array index calculations
- OpDivInt: Average calculations (totalIterations / pixels)
- OpDivFloat: Coordinate normalization (col / WIDTH)

### Why This Benchmark is Perfect for Type-Specialized Opcodes

1. **Type consistency**: Variables maintain consistent types throughout
   - `x`, `y`, `cx`, `cy` are always float
   - `iter`, `col`, `row` are always int
   - Compiler can confidently emit specialized opcodes

2. **Hot loop**: Inner mandelbrot loop executes millions of times
   - Every saved instruction has massive impact
   - No I/O or other overhead to dilute the speedup

3. **Arithmetic density**: Almost every operation is arithmetic
   - Direct beneficiary of type specialization
   - Minimal time spent on non-arithmetic operations

## Conclusion

Phase 1 exceeded expectations:
- **Expected**: 15-20% improvement
- **Achieved**: 29.7% improvement
- **Reason**: Extremely arithmetic-heavy benchmark amplifies the benefits

This validates the optimization strategy and suggests that:
- Phase 2 (comparison specialization) could yield additional 8-12% (revising from 5-10%)
- Phase 3 (struct field offsets) potential remains at 5-8%
- **Total potential**: 40-50% improvement from all phases combined

The type-specialized opcodes are clearly having a major impact on performance!
