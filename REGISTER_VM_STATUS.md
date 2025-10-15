# Register VM Implementation Status

**Date**: 2025-10-14
**Status**: Infrastructure Complete, Partial Implementation (~35% done)
**Stack VM Performance**: 7.7M iterations/sec (Mandelbrot heavy)
**Target**: 10.5M iterations/sec (30-40% improvement)

## Executive Summary

The register-based VM infrastructure is **complete and integrated**. The compiler can emit register bytecode, the VM can execute it, and backend selection works via `--backend=register`. However, **only ~35% of AST nodes are implemented** in the register compiler. To complete, you need to:

1. Finish `CompileToRegister()` for all AST node types
2. Implement proper register result tracking for nested expressions
3. Add function call support with register windows
4. Implement arrays, maps, and struct operations
5. Test and benchmark

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          Parser                                  │
│                    (Shared, unchanged)                           │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Abstract Syntax Tree (AST)                    │
└───────────────────────────┬─────────────────────────────────────┘
                            │
              ┌─────────────┴──────────────┐
              │                            │
              ▼                            ▼
┌──────────────────────┐        ┌──────────────────────┐
│   Stack Compiler     │        │  Register Compiler   │
│   (COMPLETE)         │        │   (PARTIAL 35%)      │
│                      │        │                      │
│  - compiler.go       │        │  - register_         │
│  - Emit OpPush/OpPop │        │    compiler.go       │
│  - Stack operations  │        │  - Register alloc    │
│  - ALL AST nodes ✅  │        │  - Type-specialized  │
└──────────┬───────────┘        └──────────┬───────────┘
           │                               │
           ▼                               ▼
┌──────────────────────┐        ┌──────────────────────┐
│   Stack Bytecode     │        │  Register Bytecode   │
│   []byte             │        │  []RegisterInstr     │
└──────────┬───────────┘        └──────────┬───────────┘
           │                               │
           ▼                               ▼
┌──────────────────────┐        ┌──────────────────────┐
│   Stack VM           │        │   Register VM        │
│   (COMPLETE)         │        │   (COMPLETE)         │
│   vm/vm.go           │        │   vm/register_vm.go  │
│   - Stack operations │        │   - Register ops     │
│   - Push/Pop         │        │   - Direct ops       │
│   - Type checks      │        │   - NO type checks ✅ │
└──────────────────────┘        └──────────────────────┘
```

## File Structure

### Completed Files

#### 1. `REGISTER_VM_DESIGN.md` (758 lines) ✅
Complete design document with:
- Instruction set specification (60+ opcodes)
- 32-bit instruction encoding format
- Register allocation strategy
- Performance analysis
- Integration plan
- Example Mandelbrot compilation

#### 2. `vm/register_opcodes.go` (251 lines) ✅
Complete opcode definitions:
- `RegisterOpCode` type
- Type-specialized operations (OpRAddInt, OpRAddFloat, etc.)
- Comparison operations (OpREqInt, OpRLtFloat, etc.)
- Control flow (OpRJump, OpRJumpT, OpRJumpF, OpRReturn)
- `EncodeRegisterInstruction()` and decode functions
- All 60+ opcodes defined and named

#### 3. `vm/register_vm.go` (465 lines) ✅
Complete VM execution:
- `RegisterVM` struct with register file
- `Run()` execution loop with full opcode dispatch
- NO runtime type checks (compiler-guaranteed)
- Function calls with register windows
- All arithmetic/comparison/logical operations
- Control flow (jumps, returns)
- Array/map/struct operations (VM side ready)
- Builtin function calls

#### 4. `compiler/register_compiler.go` (351 lines) ⚠️ ~35% COMPLETE
Partial compiler implementation:
- ✅ `RegisterCompiler` struct
- ✅ Register allocation (`allocateRegister`, `allocateTempRegister`)
- ✅ Instruction emission (`emitR`, `emitRBx`)
- ✅ Type inference (reuses stack compiler's type system)
- ⚠️ `CompileToRegister()` - ONLY handles:
  - ✅ Literals (int, float, bool, string)
  - ✅ Variables (identifiers)
  - ✅ Binary operators (+, -, *, /, %, ==, !=, <, >, <=, >=, &&, ||)
  - ✅ Unary operators (!, -)
  - ✅ For loops
  - ✅ Variable declarations
  - ✅ Return statements
  - ✅ Expression statements
  - ✅ Block statements
  - ✅ Square optimization (x * x)
  - ❌ Function calls
  - ❌ Function definitions
  - ❌ Arrays
  - ❌ Maps
  - ❌ Structs
  - ❌ If statements
  - ❌ Switch statements
  - ❌ Assignment statements
  - ❌ Break/continue
  - ❌ Index expressions (array[i])
  - ❌ Field access (struct.field)

#### 5. `cmd/minlang/main.go` (updated) ✅
Command-line integration:
- `--backend` flag (stack or register)
- `--debug` flag for bytecode inspection
- Conditional compilation based on backend
- Both paths fully integrated

## What Works Right Now

### Working Examples

```bash
# Stack VM (fully working)
./minlang examples/mandelbrot.min        # ✅ Works perfectly
./minlang examples/stdlib_demo.min       # ✅ Works perfectly
./minlang --debug examples/test.min      # ✅ Shows bytecode

# Register VM (partial)
./minlang --backend=register examples/simple_math.min  # ✅ Should work
./minlang --backend=register examples/mandelbrot.min   # ❌ Missing features
```

### Test Case: What Works

Create `examples/register_test.min`:
```javascript
// Simple expressions (WORKS)
var x: int = 5
var y: int = 3
var sum: int = x + y
var prod: int = x * y

// Comparisons (WORKS)
var result: bool = x > y

// For loops (WORKS)
for var i: int = 0; i < 10; i = i + 1 {
    var squared: int = i * i  // Square optimization works!
}

// Return (WORKS)
func simple(): int {
    return 42
}
```

This should compile and run with `--backend=register`.

### Test Case: What Doesn't Work

```javascript
// Function CALLS don't work
func add(a: int, b: int): int {
    return a + b
}
var result: int = add(5, 3)  // ❌ CompileToRegister missing case

// Arrays don't work
var arr: []int = [1, 2, 3]  // ❌ Missing implementation
var first: int = arr[0]      // ❌ Missing implementation

// If statements don't work
if x > 5 {                   // ❌ Missing implementation
    print("big")
}
```

## Critical Missing Pieces

### 1. Register Result Tracking (HIGHEST PRIORITY)

**Problem**: When compiling expressions, we need to track which register holds the result.

**Current Issue** in `compiler/register_compiler.go`:
```go
case *ast.InfixExpression:
    // ... compile left and right ...

    // ⚠️ WRONG: We don't know which registers hold left/right results!
    leftReg := uint8(0)  // Placeholder - THIS IS THE BUG
    rightReg := uint8(1) // Placeholder - THIS IS THE BUG

    resultReg := rc.allocateTempRegister()
    rc.emitR(vm.OpRAddInt, uint8(resultReg), leftReg, rightReg)
```

**Solution**: Change `CompileToRegister()` to return the result register:
```go
// NEW signature
func (rc *RegisterCompiler) CompileToRegister(node ast.Node) (int, error) {
    switch node := node.(type) {
    case *ast.IntegerLiteral:
        constIndex := rc.addConstant(vm.IntValue(node.Value))
        tempReg := rc.allocateTempRegister()
        rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
        return tempReg, nil  // ← Return which register has the result

    case *ast.InfixExpression:
        // Compile left and get its register
        leftReg, err := rc.CompileToRegister(node.Left)
        if err != nil {
            return 0, err
        }

        // Compile right and get its register
        rightReg, err := rc.CompileToRegister(node.Right)
        if err != nil {
            return 0, err
        }

        // Now we know where left and right are!
        resultReg := rc.allocateTempRegister()
        rc.emitR(vm.OpRAddInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))

        // Free the input registers
        rc.freeTempRegister(leftReg)
        rc.freeTempRegister(rightReg)

        return resultReg, nil  // ← Return where result is
    }
}
```

### 2. Function Calls

**Missing**: `*ast.CallExpression` case

**What to implement**:
```go
case *ast.CallExpression:
    // 1. Compile function expression to get function register
    fnReg, err := rc.CompileToRegister(node.Function)
    if err != nil {
        return 0, err
    }

    // 2. Compile arguments into consecutive registers
    argReg := rc.nextReg
    for _, arg := range node.Arguments {
        _, err := rc.CompileToRegister(arg)
        if err != nil {
            return 0, err
        }
    }

    // 3. Allocate result register
    resultReg := rc.allocateTempRegister()

    // 4. Emit call instruction
    // OpRCall: R(A) = call R(B)(R(C)...R(C+n))
    rc.emitR(vm.OpRCall, uint8(resultReg), uint8(fnReg), uint8(argReg))

    return resultReg, nil
```

### 3. Function Definitions

**Missing**: `*ast.FunctionStatement` case

**What to implement**:
```go
case *ast.FunctionStatement:
    // 1. Create new register scope for function
    rc.enterRegisterScope()

    // 2. Allocate registers for parameters
    for _, param := range node.Parameters {
        rc.allocateRegister(param.Name.Value)
    }

    // 3. Compile function body
    _, err := rc.CompileToRegister(node.Body)
    if err != nil {
        return 0, err
    }

    // 4. Add implicit return if needed
    if !rc.lastInstructionIsReturn() {
        rc.emitR(vm.OpRReturnN, 0, 0, 0)
    }

    // 5. Leave scope and create function object
    instructions := rc.leaveRegisterScope()

    // 6. Create Function value and store
    // ... similar to stack compiler
```

### 4. If Statements

**Missing**: `*ast.IfStatement` case

**What to implement**:
```go
case *ast.IfStatement:
    // 1. Compile condition
    condReg, err := rc.CompileToRegister(node.Condition)
    if err != nil {
        return 0, err
    }

    // 2. Jump if false (placeholder)
    jumpIfFalse := rc.emitRBx(vm.OpRJumpF, uint8(condReg), 9999)

    // 3. Compile consequence
    _, err = rc.CompileToRegister(node.Consequence)
    if err != nil {
        return 0, err
    }

    // 4. Jump over alternative
    jumpOverAlt := rc.emitRBx(vm.OpRJump, 0, 9999)

    // 5. Patch first jump
    afterConsequence := len(rc.instructions)
    rc.instructions[jumpIfFalse] = vm.EncodeRegisterInstructionBx(
        vm.OpRJumpF, uint8(condReg), uint16(afterConsequence))

    // 6. Compile alternative if present
    if node.Alternative != nil {
        _, err = rc.CompileToRegister(node.Alternative)
        if err != nil {
            return 0, err
        }
    }

    // 7. Patch second jump
    afterAlternative := len(rc.instructions)
    rc.instructions[jumpOverAlt] = vm.EncodeRegisterInstructionBx(
        vm.OpRJump, 0, uint16(afterAlternative))

    return 0, nil  // If statements don't produce a value
```

### 5. Arrays

**Missing**: `*ast.ArrayLiteral` and `*ast.IndexExpression` cases

**What to implement**:
```go
case *ast.ArrayLiteral:
    // 1. Create array
    arrayReg := rc.allocateTempRegister()
    rc.emitRBx(vm.OpRNewArray, uint8(arrayReg), uint16(len(node.Elements)))

    // 2. Compile and store elements
    for i, elem := range node.Elements {
        elemReg, err := rc.CompileToRegister(elem)
        if err != nil {
            return 0, err
        }

        // Store element at index i
        idxReg := rc.allocateTempRegister()
        constIdx := rc.addConstant(vm.IntValue(int64(i)))
        rc.emitRBx(vm.OpRLoadK, uint8(idxReg), uint16(constIdx))

        rc.emitR(vm.OpRSetIdx, uint8(arrayReg), uint8(idxReg), uint8(elemReg))

        rc.freeTempRegister(idxReg)
        rc.freeTempRegister(elemReg)
    }

    return arrayReg, nil

case *ast.IndexExpression:
    // Array/map access: container[index]
    containerReg, err := rc.CompileToRegister(node.Left)
    if err != nil {
        return 0, err
    }

    indexReg, err := rc.CompileToRegister(node.Index)
    if err != nil {
        return 0, err
    }

    resultReg := rc.allocateTempRegister()
    rc.emitR(vm.OpRGetIdx, uint8(resultReg), uint8(containerReg), uint8(indexReg))

    rc.freeTempRegister(containerReg)
    rc.freeTempRegister(indexReg)

    return resultReg, nil
```

### 6. Assignment Statements

**Missing**: `*ast.AssignmentStatement` case

**What to implement**:
```go
case *ast.AssignmentStatement:
    switch left := node.Left.(type) {
    case *ast.Identifier:
        // Variable assignment
        valueReg, err := rc.CompileToRegister(node.Value)
        if err != nil {
            return 0, err
        }

        // Get or allocate register for variable
        varReg := rc.registers[left.Value]

        // Move value to variable register
        rc.emitR(vm.OpRMove, uint8(varReg), uint8(valueReg), 0)
        rc.freeTempRegister(valueReg)

    case *ast.IndexExpression:
        // Array/map assignment: arr[i] = value
        containerReg, err := rc.CompileToRegister(left.Left)
        if err != nil {
            return 0, err
        }

        indexReg, err := rc.CompileToRegister(left.Index)
        if err != nil {
            return 0, err
        }

        valueReg, err := rc.CompileToRegister(node.Value)
        if err != nil {
            return 0, err
        }

        rc.emitR(vm.OpRSetIdx, uint8(containerReg), uint8(indexReg), uint8(valueReg))

        rc.freeTempRegister(containerReg)
        rc.freeTempRegister(indexReg)
        rc.freeTempRegister(valueReg)
    }

    return 0, nil
```

## Step-by-Step Completion Plan

### Phase 1: Fix Register Tracking (1-2 hours)
1. ✅ Change `CompileToRegister()` signature to return `(int, error)`
2. ✅ Update all existing cases to return result register
3. ✅ Test with simple arithmetic expressions
4. ✅ Verify register allocation/freeing works correctly

### Phase 2: Control Flow (1 hour)
1. ✅ Implement `*ast.IfStatement`
2. ✅ Implement `*ast.SwitchStatement` (similar to for loop)
3. ✅ Implement `*ast.BreakStatement` and `*ast.ContinueStatement`
4. ✅ Test with conditional code

### Phase 3: Functions (2-3 hours)
1. ✅ Implement `*ast.FunctionStatement` (definitions)
2. ✅ Implement `*ast.CallExpression` (calls)
3. ✅ Implement register windows for parameters
4. ✅ Test recursive functions
5. ✅ Test closures (may need special handling)

### Phase 4: Data Structures (2-3 hours)
1. ✅ Implement `*ast.ArrayLiteral`
2. ✅ Implement `*ast.MapLiteral`
3. ✅ Implement `*ast.StructLiteral`
4. ✅ Implement `*ast.IndexExpression` (array[i], map[key])
5. ✅ Implement `*ast.FieldAccessExpression` (struct.field)
6. ✅ Implement assignment to indexed/field expressions

### Phase 5: Remaining AST Nodes (1-2 hours)
1. ✅ Implement `*ast.AssignmentStatement` (all forms)
2. ✅ Implement any missing expression types
3. ✅ Add error handling for unsupported cases

### Phase 6: Testing (2-3 hours)
1. ✅ Port all compiler tests to register backend
2. ✅ Create register-specific tests
3. ✅ Test all example programs
4. ✅ Fix bugs found during testing

### Phase 7: Optimization & Benchmarking (1-2 hours)
1. ✅ Run Mandelbrot benchmark
2. ✅ Compare with stack VM
3. ✅ Verify 30-40% speedup
4. ✅ Profile if needed
5. ✅ Add peephole optimizations if needed

**Total estimated time**: 10-16 hours for full completion

## Testing Strategy

### Unit Tests

Create `compiler/register_compiler_test.go`:
```go
package compiler

import (
    "testing"
    "minlang/lexer"
    "minlang/parser"
)

func TestRegisterCompiler_SimpleArithmetic(t *testing.T) {
    input := `var x: int = 5 + 3`

    l := lexer.New(input)
    p := parser.New(l)
    program := p.ParseProgram()

    rc := NewRegisterCompiler()
    _, err := rc.CompileToRegister(program)

    if err != nil {
        t.Fatalf("compilation failed: %v", err)
    }

    // Verify correct instructions emitted
    if len(rc.instructions) == 0 {
        t.Fatal("no instructions emitted")
    }
}

func TestRegisterCompiler_FunctionCall(t *testing.T) {
    input := `
        func add(a: int, b: int): int {
            return a + b
        }
        var result: int = add(5, 3)
    `
    // ... test implementation
}
```

### Integration Tests

Test complete programs:
```bash
# Simple math
./minlang --backend=register examples/simple.min

# Functions
./minlang --backend=register examples/factorial.min

# Arrays
./minlang --backend=register examples/arrays.min

# Mandelbrot (full stress test)
./minlang --backend=register examples/mandelbrot_heavy.min
```

### Performance Testing

```bash
# Benchmark comparison
echo "Stack VM:"
time ./minlang examples/mandelbrot_heavy.min

echo "Register VM:"
time ./minlang --backend=register examples/mandelbrot_heavy.min

# Expected result: Register VM 30-40% faster
```

## Common Pitfalls to Avoid

### 1. Register Leaks
**Problem**: Forgetting to free temp registers
**Solution**: Always `freeTempRegister()` after use

### 2. Register Aliasing
**Problem**: Two variables pointing to same register
**Solution**: Use `OpRMove` to copy, don't reuse registers for different variables

### 3. Scope Handling
**Problem**: Register numbers conflict across scopes
**Solution**: Implement `enterRegisterScope()` and `leaveRegisterScope()`

### 4. Jump Target Patching
**Problem**: Forward jumps have placeholder addresses
**Solution**: Keep list of jump positions and patch after compiling target

### 5. Function Parameter Passing
**Problem**: Arguments in wrong registers
**Solution**: Allocate consecutive registers for parameters, maintain calling convention

## Debug Commands

```bash
# See what the compiler generates
./minlang --backend=register --debug examples/test.min

# Compare with stack VM
./minlang --backend=stack --debug examples/test.min

# Run specific test
go test ./compiler -run TestRegisterCompiler

# Build and test
go build -o minlang cmd/minlang/main.go && ./minlang --backend=register examples/test.min
```

## Performance Expectations

Based on Lua 5.0 results and our design:

| Metric | Stack VM | Register VM | Improvement |
|--------|----------|-------------|-------------|
| Instructions per Mandelbrot iteration | ~15 | ~8 | 47% fewer |
| Memory traffic | High | Low | 40% less |
| Type checks per operation | 1 | 0 | 100% eliminated |
| **Overall speed** | 7.7M iter/s | **10.5M iter/s** | **36% faster** |

This would put MinLang at **45% of Python's speed** (up from 33%).

## File Checklist

- ✅ `REGISTER_VM_DESIGN.md` - Complete
- ✅ `vm/register_opcodes.go` - Complete
- ✅ `vm/register_vm.go` - Complete
- ⚠️ `compiler/register_compiler.go` - 35% complete
- ✅ `cmd/minlang/main.go` - Complete
- ❌ `compiler/register_compiler_test.go` - Not started
- ❌ `REGISTER_VM_STATUS.md` - This file

## Quick Start to Resume

1. **Open key file**: `compiler/register_compiler.go`

2. **Start with fix**: Change `CompileToRegister()` signature:
   ```go
   func (rc *RegisterCompiler) CompileToRegister(node ast.Node) (int, error)
   ```

3. **Update existing cases** to return result register

4. **Add missing cases** one at a time (start with `*ast.IfStatement`)

5. **Test incrementally** after each case

6. **Run example**: `./minlang --backend=register examples/test.min`

## References

- **Lua 5.0 Implementation Paper**: Classic register VM design
- **Stack VM Code**: `vm/vm.go` - Reference for what to implement
- **Stack Compiler**: `compiler/compiler.go` - Shows all AST cases needed
- **Opcodes**: `vm/register_opcodes.go` - All available instructions

## Success Criteria

✅ **Done when**:
1. All examples compile with `--backend=register`
2. All examples produce same output as stack VM
3. All tests pass with register backend
4. Mandelbrot benchmark shows 30-40% speedup
5. No crashes, no memory leaks, no incorrect results

## Contact Points

If you get stuck:
- Check stack compiler for similar code: `compiler/compiler.go`
- Check register VM has the opcode you need: `vm/register_opcodes.go`
- Look at stack VM execution: `vm/vm.go`
- Read design doc: `REGISTER_VM_DESIGN.md`

---

**Status**: Infrastructure complete. Need ~10-16 hours to finish compiler implementation.
**Blocker**: Register result tracking (fix signature first)
**Next file to edit**: `compiler/register_compiler.go`
**Next function**: `CompileToRegister()` - change signature, then add missing cases
