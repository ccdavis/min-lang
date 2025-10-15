# VM Type Checking Optimization

## Summary

With compile-time type checking for arrays and maps now in place, we've streamlined the VM by eliminating redundant runtime type checks and dispatch overhead.

## Changes Made

### 1. Compiler Changes

**Before:** The compiler emitted `OpArrayGet` and `OpArraySet` for both arrays AND maps.

**After:** The compiler now emits specialized opcodes:
- `OpArrayGet` / `OpArraySet` - Only for arrays and string indexing
- `OpMapGet` / `OpMapSet` - Only for maps

Location: `compiler/compiler.go`
- Line 1015-1020: IndexExpression emits OpMapGet for maps, OpArrayGet for arrays
- Line 702-707: Assignment emits OpMapSet for maps, OpArraySet for arrays

### 2. VM Changes

**OpArrayGet** (vm/vm.go:647-693)
- **Removed:** MapType case in switch statement (was handling maps)
- **Kept:** ArrayType and StringType cases
- **Kept:** Integer index type checking (necessary since indices can be any expression)
- **Kept:** Bounds checking (necessary for safety)
- **Benefit:** No more runtime type dispatch for arrays vs maps

**OpArraySet** (vm/vm.go:695-718)
- **Removed:** MapType case in switch statement
- **Removed:** Switch statement entirely (now just an if check)
- **Kept:** Integer index type checking
- **Kept:** Bounds checking
- **Benefit:** Simplified code path, no type dispatch

**OpMapGet** (vm/vm.go:742-761)
- **Removed:** `if mapVal.Type != MapType` type check
- **Benefit:** Compiler guarantees correctness, no runtime check needed

**OpMapSet** (vm/vm.go:763-771)
- **Removed:** `if mapVal.Type != MapType` type check
- **Benefit:** Compiler guarantees correctness, no runtime check needed

## Type Safety Guarantees

The compiler now enforces:

1. **Array element types**: All elements in an array must match the declared element type
   - Example: `var nums: []int = [1, 2, "three"]` → Compilation error

2. **Map key types**: All keys used to access or assign to a map must match the declared key type
   - Example: `var ages: map[string]int; ages[123] = 30` → Compilation error

3. **Map value types**: All values assigned to a map must match the declared value type
   - Example: `ages["Bob"] = "thirty"` (where ages is map[string]int) → Compilation error

4. **Nested collections**: Type checking works recursively for nested arrays and maps
   - Example: `var matrix: [][]int = [[1, 2], [3, "four"]]` → Compilation error

## Performance Benefits

### Eliminated Operations:
1. **Runtime type dispatch**: The VM no longer needs to check container type (array vs map) at runtime
2. **Redundant type checks**: Map operations no longer verify the container is actually a map

### Measurements:
- All existing tests pass ✓
- Type error tests correctly catch violations at compile time ✓
- 10,000 array/map operations complete in ~5-6ms

## What's Still Checked at Runtime

1. **Array indices must be integers**: Since index expressions can be any type, we still verify integer type
2. **Bounds checking**: Array/string index bounds are checked for safety
3. **Nil checks**: Map lookups returning nil are still handled

## Impact

This optimization follows the same philosophy as the Phase 1 and Phase 2 optimizations for arithmetic and comparison operations:
- **Move type checking from runtime to compile time**
- **Use specialized opcodes to avoid runtime dispatch**
- **Let the compiler do the work once, so the VM doesn't repeat it**

The result is a faster VM with the same safety guarantees, now enforced at compile time instead of runtime.
