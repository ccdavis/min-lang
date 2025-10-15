# CPU Cache Efficiency Analysis

**Date**: 2025-10-09
**Branch**: cache-efficiency-optimizations

## Overview

This analysis evaluates MinLang's VM implementation against CPU cache-friendly design principles from modern systems programming, specifically focusing on techniques outlined in "CPU Cache-Friendly Go" by skoredin.pro.

## Key CPU Cache Principles

**Cache hierarchy (typical x86-64):**
- L1 Cache: ~4 cycles (1ns) - 32-64 KB
- L2 Cache: ~12 cycles (3ns) - 256-512 KB
- L3 Cache: ~40 cycles (10ns) - 8-16 MB
- RAM: 200+ cycles (60ns+)

**Cache line size**: 64 bytes on modern CPUs

**Golden Rule**: "Modern CPUs are fast. Memory is slow. The gap grows every year."

## Current VM Implementation Analysis

### ✅ EXCELLENT: Value Struct (vm/value.go:45-49)

```go
type Value struct {
    Type ValueType  // 1 byte
    _    [7]byte    // 7 bytes padding (explicit!)
    Data uint64     // 8 bytes
}
```

**Size**: 16 bytes (perfectly aligned)

**Cache friendliness**: ⭐⭐⭐⭐⭐
- **4 Values fit in a single 64-byte cache line**
- Tagged union avoids pointer indirection for primitives
- Explicit padding shows awareness of alignment
- Zero heap allocation for int/float/bool
- Linear memory layout in slices

**Impact**: This is textbook cache-friendly design. No changes needed.

### ✅ GOOD: Stack-Based Architecture (vm/vm.go:51)

```go
stack []Value  // Array of 2048 Values = 32KB
```

**Cache friendliness**: ⭐⭐⭐⭐
- Stack operations access contiguous memory
- 32KB fits in L1 cache on modern CPUs
- Sequential access pattern (push/pop)
- Predictable for CPU prefetcher

**Minor concern**: Stack is part of VM struct, which may cause false sharing if VM instances were used concurrently (currently single-threaded, so not an issue).

### ⚠️ ISSUE #1: Frame Pointers (vm/vm.go:56-57)

```go
type VM struct {
    // ...
    frames      []*Frame  // Array of POINTERS to Frame
    framesIndex int
}
```

**Cache friendliness**: ⭐⭐ (Poor)

**Problem**: Pointer chasing
- Each frame access requires dereferencing a pointer
- Frames are heap-allocated separately
- Non-contiguous memory layout
- Cache misses when switching frames

**Current behavior**:
```
frames[0] -> *Frame at 0x12340000
frames[1] -> *Frame at 0x56780000  // Different cache line!
frames[2] -> *Frame at 0x9abc0000  // Another cache miss!
```

**Impact**: Every function call/return causes potential cache miss.

**Article principle violated**: "Prefer linear memory access over random access"

### ⚠️ ISSUE #2: Frame Allocation Pattern (vm/vm.go:1504-1508)

```go
frame := vm.frames[vm.framesIndex]
if frame == nil {
    frame = &Frame{}  // Heap allocation
    vm.frames[vm.framesIndex] = frame
}
```

**Cache friendliness**: ⭐⭐

**Problem**: Lazy allocation with unpredictable heap locations
- First allocation creates heap object
- Subsequent frames allocated at arbitrary addresses
- No spatial locality guarantee

**Better approach**: Pre-allocate all frames in contiguous array

### ✅ GOOD: Embedded Closure Optimization (vm/vm.go:29-30)

```go
type Frame struct {
    // ...
    tempClosure Closure  // Embedded! Avoids allocation
}
```

**Cache friendliness**: ⭐⭐⭐⭐
- Inline data structure (no pointer chase)
- Already demonstrates understanding of cache locality

**This shows you understand the principle!** Now apply it to frames themselves.

### ⚠️ ISSUE #3: Object Pools Growing Unbounded (vm/value.go:12-20)

```go
var stringPool []*string
var functionPool []*Function
var closurePool []*Closure
var arrayPool []*ArrayValue
var mapPool []*MapValue
var structPool []*StructValue
```

**Cache friendliness**: ⭐⭐⭐ (Declining over time)

**Problems**:
1. **Append-only growth**: Pools never shrink
2. **Pointer array**: Each access requires indirection
3. **No spatial locality**: Objects allocated at random heap addresses
4. **Memory bloat**: Long-running programs accumulate garbage

**Impact**: As pools grow, cache hit rate decreases.

**Article principle**: "Group hot data together, separate hot and cold data"

### ⚠️ ISSUE #4: Constants Array Access (vm/vm.go:119-147)

**Hot path**: `vm.constants[constIndex]` accessed in OpPush (very frequent)

```go
case OpPush:
    constIndex, _ := ReadOperand(ins, ip)
    ip += 2
    err := vm.push(vm.constants[constIndex])  // Array access
```

**Current behavior**:
- `constants` is `[]Value` (good - contiguous)
- Accessed by index (good - predictable)
- But constants are accessed randomly based on bytecode

**Potential optimization**: Group frequently-accessed constants together at beginning of array (requires compiler cooperation).

**Cache friendliness**: ⭐⭐⭐ (Good but could be better)

### ✅ EXCELLENT: Direct Local Operations (vm/vm.go:169-235)

```go
case OpAddLocal:
    localIndex, _ := ReadOperand(ins, ip)
    ip += 2

    tos := vm.pop()
    local := vm.stack[frame.basePointer+localIndex]  // Direct stack access
    result := tos.AsInt() + local.AsInt()
```

**Cache friendliness**: ⭐⭐⭐⭐⭐
- Accesses local variables directly from stack
- No indirection through globals array
- Stack already in L1 cache from recent push/pop
- Excellent cache locality

**This optimization is brilliant** - locals are hot, keep them in hot memory.

### ✅ GOOD: Struct Field Array Access (vm/value.go:209-210)

```go
type StructValue struct {
    TypeName    string
    Fields      map[string]Value  // Name-based (backward compat)
    FieldsArray []Value          // Offset-based (Phase 3 opt)
    FieldOrder  []string
}
```

**Cache friendliness**:
- Map access: ⭐⭐ (hash lookup, pointer chase)
- Array access: ⭐⭐⭐⭐ (direct offset, contiguous)

**Phase 3 optimization is working as intended** - array access is cache-friendly.

**One issue**: Storing both map AND array duplicates data (memory overhead).

## Recommendations

### 🔥 HIGH IMPACT: Make Frames Value Types

**Change**:
```go
// Before
frames []*Frame

// After
frames []Frame  // Direct array of Frames (no pointers!)
```

**Benefits**:
- All frames in contiguous memory
- Cache line sharing between adjacent frames
- Eliminates pointer dereferencing overhead
- Predictable memory layout

**Estimated impact**: 5-10% on function-call-heavy workloads

**Complexity**: LOW (main change in vm.go:New and frame access patterns)

**Example**:
```go
// Old: frame := vm.frames[vm.framesIndex-1]  (pointer deref)
// New: frame := &vm.frames[vm.framesIndex-1] (address of array element)
```

### 🔥 HIGH IMPACT: Pre-allocate Frame Array

**Change**:
```go
// Before
frames := make([]*Frame, MaxFrames)
frames[0] = mainFrame

// After
frames := make([]Frame, MaxFrames)  // Pre-allocate all frames
frames[0] = mainFrame
```

**Benefits**:
- Zero heap allocations during execution
- All frames in same memory region
- Better cache prefetching
- Eliminates nil checks

**Estimated impact**: 3-5% on recursive workloads

**Complexity**: LOW

### 🔵 MEDIUM IMPACT: Optimize Constant Access Pattern

**Change**: Add compiler hint to place frequently-used constants first

```go
// Compiler emits most-used constants at indices 0-N
// These stay hot in cache
```

**Benefits**:
- Hot constants stay in L1 cache
- Fewer cache evictions
- Better prefetcher behavior

**Estimated impact**: 2-4% on const-heavy workloads

**Complexity**: MEDIUM (requires compiler changes)

### 🔵 MEDIUM IMPACT: Cache-Aligned Hot Structures

**Add padding to cache-align frequently accessed fields**:

```go
type VM struct {
    // Hot path fields (frequently accessed together)
    sp          int
    framesIndex int
    _           [56]byte  // Pad to cache line (64 bytes)

    // Warm fields
    stack    []Value
    frames   []Frame
    _        [48]byte  // Pad to cache line

    // Cold fields (rarely accessed during execution)
    constants []Value
    globals   []Value
}
```

**Benefits**:
- sp and framesIndex on same cache line (accessed together in hot loop)
- Prevents false sharing
- Reduces cache line bouncing

**Estimated impact**: 1-3% overall

**Complexity**: MEDIUM (requires careful field ordering and measurement)

**Article principle**: "Align data for cache line efficiency"

### 🟢 LOW IMPACT: Pool Size Limits and Cleanup

**Change**: Add maximum pool sizes and periodic cleanup

```go
const MaxPoolSize = 10000

func cleanupPools() {
    if len(stringPool) > MaxPoolSize {
        stringPool = stringPool[len(stringPool)-MaxPoolSize:]
    }
    // ... other pools
}
```

**Benefits**:
- Prevents unbounded memory growth
- Keeps active pool data in cache
- Reduces memory pressure

**Estimated impact**: 0-2% (mainly on long-running programs)

**Complexity**: LOW

### 🟢 LOW IMPACT: Remove Duplicate Struct Storage

**Change**: Only store FieldsArray, compute map on demand if needed

```go
type StructValue struct {
    TypeName    string
    FieldsArray []Value
    FieldOrder  []string  // For field name lookup
}

// Add method for name-based access (slower path)
func (s *StructValue) GetByName(name string) (Value, bool) {
    for i, fieldName := range s.FieldOrder {
        if fieldName == name {
            return s.FieldsArray[i], true
        }
    }
    return NilValue(), false
}
```

**Benefits**:
- Reduces memory by ~50% per struct instance
- Fewer cache lines needed
- Offset-based access unchanged (still fast)

**Drawback**: Name-based access becomes O(n) instead of O(1)
- **OK**: Phase 3 optimization makes name-based access rare

**Estimated impact**: 1-2% (memory reduction helps cache)

**Complexity**: LOW (mainly affects struct operations)

## Cache Efficiency Best Practices Already Followed

✅ **Tagged union design** - Value struct avoids boxing
✅ **Embedded structs** - tempClosure in Frame
✅ **Pre-allocated errors** - No allocations on error paths
✅ **Array-based access** - Phase 3 struct field optimization
✅ **Stack-based architecture** - Sequential memory access
✅ **Direct local operations** - Minimizes stack push/pop

**Your VM already demonstrates strong cache-awareness!** 🎉

## Measurement Strategy

Before implementing changes:

1. **Baseline benchmark**: Run `mandelbrot_heavy.min` 10 times, record average
2. **Profile with perf**:
   ```bash
   perf stat -e cache-references,cache-misses,L1-dcache-loads,L1-dcache-load-misses \
       ./minlang examples/mandelbrot_heavy.min
   ```
3. **Implement Frame value-type change** (highest impact)
4. **Re-benchmark**: Compare cache-misses before/after
5. **Iterate**: Add optimizations incrementally, measure each

## Expected Overall Impact

**Conservative estimate**: 8-12% improvement on compute-intensive workloads
**Optimistic estimate**: 15-20% improvement on function-call-heavy workloads

**Best targets**:
- Recursive algorithms (fibonacci, tree traversal)
- Function-heavy code (many calls/returns)
- Tight loops with function calls

**Neutral impact**:
- I/O-bound programs
- Programs with few function calls

## Conclusion

MinLang's VM already shows excellent cache-awareness in many areas, particularly:
- The Value tagged union design
- Stack-based architecture
- Phase 3 struct field optimization
- Embedded tempClosure

**The #1 optimization opportunity is converting Frame pointers to value types**, which aligns perfectly with the article's principle of **"Struct of Arrays (SoA) instead of Array of Structs (AoS)"** - or in this case, **"Array of Structs instead of Array of Struct Pointers."**

This change alone could yield 5-10% improvement on function-heavy workloads, with minimal code changes and zero breaking changes to the language semantics.

## Next Steps

1. ✅ Create branch `cache-efficiency-optimizations` (DONE)
2. Implement Frame value-type change
3. Benchmark and measure cache statistics
4. Implement pre-allocation optimization
5. Profile and measure again
6. Consider other optimizations based on results

---

**References**:
- Article: https://skoredin.pro/blog/golang/cpu-cache-friendly-go
- Cache line size: 64 bytes (x86-64 standard)
- Go memory model: https://go.dev/ref/mem
