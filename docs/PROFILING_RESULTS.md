# VM Profiling Results

## Summary

CPU profiling was performed on both the stack VM and register VM using the mandelbrot_heavy and fibonacci_heavy benchmarks. The results reveal significant performance differences and clear bottlenecks.

## Performance Comparison

### Mandelbrot Heavy Benchmark
- **Stack VM**: 10.75s
- **Register VM**: 7.18s
- **Performance Gain**: 33% faster with register VM

### Fibonacci Heavy Benchmark
- **Stack VM**: 540ms
- **Register VM**: Failed (call stack overflow - known issue)

## Stack VM Bottlenecks (vm/vm.go)

### Mandelbrot Heavy Profile
```
Total samples: 10.75s
Major bottlenecks:
- pop():           3.65s (33.95%)  ← CRITICAL BOTTLENECK
- VM.Run():        3.92s (36.47%)
- ReadOperand():   1.31s (12.19%)
- push():          1.01s (9.40%)   ← CRITICAL BOTTLENECK

Combined push/pop overhead: 4.66s (43.35%)
```

### Fibonacci Heavy Profile
```
Total samples: 540ms
Major bottlenecks:
- VM.Run():        240ms (44.44%)
- pop():           120ms (22.22%)  ← CRITICAL BOTTLENECK
- ReadOperand():    80ms (14.81%)
- push():           50ms (9.26%)   ← CRITICAL BOTTLENECK

Combined push/pop overhead: 170ms (31.48%)
```

### Analysis
The stack VM spends **31-43% of execution time** just managing the stack with push/pop operations. These are very simple operations (vm/vm.go:95-112):

```go
func (vm *VM) push(val Value) error {
    if vm.sp >= StackSize {
        return ErrStackOverflow
    }
    vm.stack[vm.sp] = val
    vm.sp++
    return nil
}

func (vm *VM) pop() Value {
    if vm.sp <= 0 {
        panic(fmt.Sprintf("stack underflow: sp=%d", vm.sp))
    }
    val := vm.stack[vm.sp-1]
    vm.sp--
    return val
}
```

Even though these functions are highly optimized, they are called so frequently (every arithmetic operation, every variable access) that they dominate execution time.

## Register VM Performance (vm/register_vm.go)

### Mandelbrot Heavy Profile
```
Total samples: 7.18s
Time distribution:
- RegisterVM.Run():  6.57s (91.50%)
- Value.IsTruthy():  0.28s (3.90%)
- Value.AsBool():    0.21s (2.92%)
```

### Analysis
The register VM has **NO push/pop overhead** - it operates directly on registers:
- All time is spent in the main dispatch loop
- Direct register-to-register operations (e.g., `regs[a] = regs[b] + regs[c]`)
- No intermediate stack manipulation required

Example from vm/register_vm.go:143-146:
```go
case OpRAddInt:
    regs[a] = IntValue(regs[b].AsInt() + regs[c].AsInt())
```

This is a single assignment with no stack operations, compared to the stack VM which requires:
1. pop() right operand
2. pop() left operand
3. perform operation
4. push() result

## Key Findings

1. **Stack Operations Are The Bottleneck**: 31-43% of stack VM execution time is pure overhead from stack manipulation

2. **Register VM Is Fundamentally Faster**: By eliminating stack operations, the register VM achieves 33% better performance on compute-heavy workloads

3. **Dispatch Loop Overhead**: Both VMs spend most time in their main dispatch loops, but the register VM's is more efficient because:
   - Fewer instructions executed per operation
   - No function call overhead for push/pop
   - Direct register access vs. stack pointer manipulation

4. **ReadOperand Cost**: In the stack VM, `ReadOperand()` accounts for 12-15% of time, decoding 16-bit operands from the bytecode stream

5. **Register VM Issues**:
   - Call stack is too shallow (fails fibonacci_heavy)
   - Needs larger call stack allocation

## Optimizations Implemented

### Register VM Optimizations (Completed)
1. ✅ **Increased MaxFrames**: From 1024 to 8192 to handle deeper call stacks
2. ✅ **Optimized instruction decode**: Removed switch statement, always decode ABC format first
3. ✅ **Cached VM fields**: Cache constants and globals arrays to reduce pointer dereferences

### Performance Impact
- Mandelbrot benchmark: Maintained 33.4% speedup over stack VM (7.29s vs 10.94s)
- Instruction decode overhead reduced (removed conditional decoding logic)
- Profile shows RegisterVM.Run still at ~90% (expected for tight dispatch loop)

### Remaining Recommendations

#### For Stack VM Optimization
1. **Inline push/pop**: The compiler already tries to inline these, but explicit `//go:inline` pragmas might help
2. **Reduce stack operations**: Use more specialized opcodes that combine multiple operations
3. **Stack pointer optimizations**: Cache vm.sp in a local variable in hot loops

#### For Register VM
1. ⚠️ **Deep recursion issue**: Register VM still fails on fibonacci_heavy due to per-frame register allocation
   - Each call allocates new register array: `make([]Value, numRegs)`
   - For deep recursion (millions of calls), this causes memory exhaustion
   - Solution: Implement register window pooling or use a single register file with offsets
2. **Jump table dispatch**: Consider computed goto alternative (array of closures) for opcode dispatch
3. **Further Value optimizations**: Consider specialized Value types to reduce method call overhead

### General Optimizations
1. **Value type operations**: Consider more specialized Value operations to reduce AsInt()/AsFloat() calls
2. **Jump table dispatch**: Consider computed goto or jump table for opcode dispatch (Go doesn't support computed goto, but array of closures might work)
3. **Trace compilation**: For hot loops, consider JIT compilation or bytecode specialization

## Profiling Commands Used

```bash
# Build with profiling support
go build -o ./minlang ./cmd/minlang

# Profile stack VM
./minlang -backend=stack -cpuprofile=stack_mandelbrot.prof examples/mandelbrot_heavy.min
./minlang -backend=stack -cpuprofile=stack_fibonacci.prof examples/fibonacci_heavy.min

# Profile register VM
./minlang -backend=register -cpuprofile=register_mandelbrot.prof examples/mandelbrot_heavy.min

# Analyze profiles
go tool pprof -top stack_mandelbrot.prof
go tool pprof -top register_mandelbrot.prof
```

## Conclusion

The profiling clearly shows that **stack operations are the primary bottleneck** in the stack VM, accounting for up to 43% of execution time. The register VM eliminates this overhead entirely, resulting in 33% better performance on compute-heavy workloads. Both VMs have room for further optimization, particularly in the instruction dispatch loop and value type operations.
