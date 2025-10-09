# Cache Efficiency Optimization Results

**Date**: 2025-10-09
**Branch**: cache-efficiency-optimizations
**Optimization**: Convert Frame pointers to value types

## Summary

**Result**: ❌ **NEGATIVE IMPACT** (~1.9% slower on mandelbrot, neutral on fibonacci)

The optimization was theoretically sound but showed no benefit in practice. This is a valuable lesson in empirical performance work: **measure, don't assume**.

## Changes Implemented

### Code Changes
- Changed `frames []*Frame` to `frames []Frame` (vm/vm.go:56)
- Pre-allocate all 1024 frames upfront as value array
- Updated all frame access to use `&vm.frames[i]` instead of `vm.frames[i]`
- Eliminated nil checks and lazy heap allocation

### Benefits (Theoretical)
- ✅ All frames in contiguous memory (cache-friendly)
- ✅ Zero heap allocations during function calls
- ✅ Predictable memory layout
- ✅ Better spatial locality

### What We Got (Actual)
- ❌ 1.9% slower on compute-intensive workload
- ➖ Neutral on function-call-intensive workload

## Benchmark Results

### Baseline (master - pointer frames)

**Mandelbrot Heavy (122M iterations):**
```
Run 1: 11.52s
Run 2: 11.51s
Run 3: 11.64s
Run 4: 11.64s
Run 5: 11.68s
Average: 11.60s
```

**Fibonacci Heavy (fib(30) × 5):**
```
Run 1: 0.59s
Run 2: 0.59s
Run 3: 0.60s
Run 4: 0.60s
Run 5: 0.58s
Average: 0.59s
```

### Optimized (cache-friendly frames)

**Mandelbrot Heavy:**
```
Run 1: 11.77s
Run 2: 11.79s
Run 3: 11.83s
Run 4: 11.84s
Run 5: 11.86s
Average: 11.82s
```
**Impact**: +1.9% slower ❌

**Fibonacci Heavy:**
```
Run 1: 0.59s
Run 2: 0.58s
Run 3: 0.59s
Run 4: 0.59s
Run 5: 0.58s
Average: 0.59s
```
**Impact**: Neutral (±0%)

## Analysis: Why Did This Fail?

### Hypothesis 1: Frame Size Problem ⭐⭐⭐⭐⭐ (Most Likely)

**Frame struct size:**
```go
type Frame struct {
    cl          *Closure      // 8 bytes
    ip          int           // 8 bytes
    basePointer int           // 8 bytes
    tempClosure Closure       // 32 bytes (Fn *Function + Free []Value)
}
// Total: 56 bytes per frame
```

**Memory impact:**
- Old: `[]*Frame` = 1024 × 8 = 8 KB (pointer array)
- New: `[]Frame` = 1024 × 56 = 56 KB (value array)

**Problem**: We pre-allocate 56KB of zero-initialized frames!

**Cache pollution:**
1. **VM initialization**: Zeroing 56KB touches ~875 cache lines
2. **VM struct bloat**: VM struct now includes massive 56KB array
   - Pushes other hot fields (sp, framesIndex, constants) further apart
   - May span multiple pages
3. **Cold frames thrash hot data**: Unused frames (90%+ of array) can evict hot VM data from cache

**The irony**: We tried to make frames cache-friendly, but made the **VM struct** cache-**un**friendly!

### Hypothesis 2: False Sharing (Unlikely)

Since VM is single-threaded, false sharing between cores doesn't apply.

### Hypothesis 3: Lazy Allocation Was Actually Good ⭐⭐⭐

**Old approach benefits:**
- Frames allocated on-demand on heap
- Most programs use <10 frames, not 1024
- Heap allocations are localized near each other (heap has spatial locality too!)
- Smaller VM struct keeps hot fields together

**New approach downsides:**
- Pre-allocating 1024 frames wastes memory
- Only frames [0-N] are used, rest are cold
- 56KB footprint affects VM struct cache behavior

### Hypothesis 4: Go's Heap Allocator Is Really Good ⭐⭐⭐⭐

Modern Go uses **tcmalloc**-style allocator:
- Small objects allocated from thread-local cache (fast!)
- Similar-sized allocations grouped together (spatial locality!)
- Frames likely allocated near each other anyway

**Our assumption was wrong**: Heap != random addresses

### Hypothesis 5: Pointer Indirection Not the Bottleneck ⭐⭐

**Modern CPUs:**
- L1 cache hit: ~4 cycles
- Pointer dereference: 1 instruction
- Branch prediction works well for frame access pattern

**Reality**: The extra pointer deref was negligible compared to:
- Arithmetic operations (mandelbrot: ~500M ops)
- Stack push/pop operations
- Instruction decoding

## Key Lessons Learned

### 1. **Cache Optimization Is Subtle**

What helps in one context (SoA for data-parallel code) can hurt in another (VM with large structs).

### 2. **Pre-allocation Has Costs**

- Memory footprint
- Initialization overhead
- Splitting hot/cold data

Sometimes lazy allocation is better!

### 3. **Measure, Don't Assume**

Our theoretical analysis was sound, but:
- Didn't account for Frame size (56 bytes!)
- Didn't consider VM struct bloat
- Assumed heap allocations were worse than they are

### 4. **Context Matters**

The article's advice applies to:
- Data-parallel processing
- Hot loops with arrays of structs
- High-frequency access patterns

**MinLang's context:**
- Low frame count (typically <10 active)
- Infrequent function calls relative to arithmetic
- Frame access not the bottleneck

### 5. **Go's Runtime Is Clever**

Don't fight the runtime! Go's:
- Escape analysis
- Stack allocation where possible
- Fast heap allocator
- GC optimizations

...all work together well.

## What Would Actually Help?

Based on the cache article and our findings:

### 1. **Hot/Cold Field Separation in VM struct** ⭐⭐⭐⭐

```go
type VM struct {
    // Hot fields (accessed every instruction)
    sp          int
    framesIndex int
    _           [48]byte  // Pad to cache line (64 bytes)

    // Warm fields (accessed frequently)
    stack    []Value
    frames   []*Frame

    // Cold fields (rarely accessed during execution)
    constants []Value
    globals   []Value
}
```

**Why this helps:**
- `sp` and `framesIndex` on same cache line
- No 56KB bloat
- Hot fields stay hot

### 2. **Reduce Value Struct Padding** ⭐⭐⭐

Current:
```go
type Value struct {
    Type ValueType  // 1 byte
    _    [7]byte    // 7 bytes padding
    Data uint64     // 8 bytes
}
// 16 bytes total, 4 per cache line
```

Could try (risky):
```go
type Value uint64  // Pack type into upper bits?
```
**Benefit**: 8 bytes instead of 16 → 8 Values per cache line!

### 3. **Inline Hot Frames** ⭐⭐⭐⭐

Instead of pre-allocating 1024, inline first few:

```go
type VM struct {
    // ... other fields

    inlineFrames [8]Frame      // First 8 frames inline (hot!)
    spillFrames  []*Frame      // Overflow to heap (cold)
    framesIndex  int
}
```

**Benefits:**
- First 8 calls use inline frames (cache-hot)
- Deep recursion spills to heap (acceptable)
- Best of both worlds!

## Recommendation

**REVERT THIS OPTIMIZATION** ✅

The value-type frames approach:
- Adds complexity
- Hurts performance (~2%)
- Doesn't match our workload characteristics

**Keep the pointer-based approach** because:
- Go's heap allocator handles it well
- Frame reuse works fine
- Smaller VM struct footprint
- Better cache behavior for hot VM fields

## Valuable Takeaway

This "failed" optimization taught us more than a successful one would have:

1. **Profile before optimizing** - frame allocation wasn't even hot
2. **Measure everything** - theory ≠ practice
3. **Understand your workload** - low frame count changes everything
4. **Trust but verify** - even expert advice needs validation in context

**The cache article was right** - for its use case (data processing, hot loops, high-frequency access).

**We were wrong** - to blindly apply it to a VM with different characteristics.

---

## Next Steps

1. ✅ Revert changes (go back to master)
2. ⚠️ Profile actual hot paths (use `go tool pprof`)
3. 💡 Try hot/cold field separation in VM struct
4. 💡 Consider inline frames optimization
5. 📊 Benchmark with perf counters (cache-misses, branch-mispredicts)

**Remember**: Premature optimization is the root of all evil. But measuring and learning never is! 🎓

