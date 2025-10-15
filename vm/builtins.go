package vm

import (
	"fmt"
	"unsafe"
)

// BuiltinFunction represents a built-in function
type BuiltinFunction func(args ...Value) Value

// Builtins is a list of built-in functions
var Builtins = []BuiltinFunction{
	printBuiltin,
	lenBuiltin,
	deleteBuiltin,
	appendBuiltin,
	keysBuiltin,
	valuesBuiltin,
	copyBuiltin,
	enumNameBuiltin,
	enumValueBuiltin,
	absBuiltin,
	minBuiltin,
	maxBuiltin,
	sqrtBuiltin,
	powBuiltin,
	floorBuiltin,
	ceilBuiltin,
	splitBuiltin,
	substringBuiltin,
	intBuiltin,
	floatBuiltin,
	stringBuiltin,
}

// EnumRegistry stores enum type information at runtime
var EnumRegistry = make(map[string]map[int]string) // enumTypeName -> (value -> name)

// printBuiltin implements the print function
func printBuiltin(args ...Value) Value {
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg.String())
	}
	fmt.Println()
	return NilValue()
}

// lenBuiltin implements the len function
func lenBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("len: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	arg := args[0]
	switch arg.Type {
	case ArrayType:
		return IntValue(int64(len(arg.AsArray().Elements)))
	case MapType:
		return IntValue(int64(len(arg.AsMap().Pairs)))
	case StringType:
		return IntValue(int64(len(arg.AsString())))
	default:
		fmt.Printf("len: argument not supported for type %d\n", arg.Type)
		return NilValue()
	}
}

// deleteBuiltin implements the delete function for maps
func deleteBuiltin(args ...Value) Value {
	if len(args) != 2 {
		fmt.Printf("delete: wrong number of arguments. got=%d, want=2\n", len(args))
		return NilValue()
	}

	mapVal := args[0]
	key := args[1]

	if mapVal.Type != MapType {
		fmt.Printf("delete: first argument must be a map\n")
		return NilValue()
	}

	mapKey := key.ToMapKey()
	mapData := mapVal.AsMap()
	delete(mapData.Pairs, mapKey)

	return NilValue()
}

// appendBuiltin implements the append function for arrays
func appendBuiltin(args ...Value) Value {
	if len(args) < 2 {
		fmt.Printf("append: wrong number of arguments. got=%d, want=2+\n", len(args))
		return NilValue()
	}

	arrayVal := args[0]
	if arrayVal.Type != ArrayType {
		fmt.Printf("append: first argument must be an array\n")
		return NilValue()
	}

	oldArray := arrayVal.AsArray()
	newElements := make([]Value, len(oldArray.Elements)+len(args)-1)

	// Copy existing elements
	copy(newElements, oldArray.Elements)

	// Append new elements
	for i := 1; i < len(args); i++ {
		newElements[len(oldArray.Elements)+i-1] = args[i]
	}

	arr := &ArrayValue{Elements: newElements}
	// Add to pool to keep it alive for GC (critical - without this the pointer becomes dangling!)
	arrayPool = append(arrayPool, arr)
	return Value{
		Type: ArrayType,
		Data: uint64(uintptr(unsafe.Pointer(arr))),
	}
}

// keysBuiltin implements the keys function for maps
func keysBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("keys: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	mapVal := args[0]
	if mapVal.Type != MapType {
		fmt.Printf("keys: argument must be a map\n")
		return NilValue()
	}

	mapData := mapVal.AsMap()
	keys := make([]Value, 0, len(mapData.Pairs))

	for mapKey := range mapData.Pairs {
		if mapKey.IsInt {
			keys = append(keys, IntValue(mapKey.IntVal))
		} else {
			keys = append(keys, StringValue(mapKey.StrVal))
		}
	}

	arr := &ArrayValue{Elements: keys}
	// Add to pool to keep it alive for GC
	arrayPool = append(arrayPool, arr)
	return Value{
		Type: ArrayType,
		Data: uint64(uintptr(unsafe.Pointer(arr))),
	}
}

// valuesBuiltin implements the values function for maps
func valuesBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("values: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	mapVal := args[0]
	if mapVal.Type != MapType {
		fmt.Printf("values: argument must be a map\n")
		return NilValue()
	}

	mapData := mapVal.AsMap()
	values := make([]Value, 0, len(mapData.Pairs))

	for _, value := range mapData.Pairs {
		values = append(values, value)
	}

	arr := &ArrayValue{Elements: values}
	// Add to pool to keep it alive for GC
	arrayPool = append(arrayPool, arr)
	return Value{
		Type: ArrayType,
		Data: uint64(uintptr(unsafe.Pointer(arr))),
	}
}

// copyBuiltin implements the copy function for arrays
func copyBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("copy: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	arrayVal := args[0]
	if arrayVal.Type != ArrayType {
		fmt.Printf("copy: argument must be an array\n")
		return NilValue()
	}

	oldArray := arrayVal.AsArray()
	newElements := make([]Value, len(oldArray.Elements))
	copy(newElements, oldArray.Elements)

	arr := &ArrayValue{Elements: newElements}
	// Add to pool to keep it alive for GC
	arrayPool = append(arrayPool, arr)
	return Value{
		Type: ArrayType,
		Data: uint64(uintptr(unsafe.Pointer(arr))),
	}
}

// enumNameBuiltin implements enumName(enumType, value) -> string
func enumNameBuiltin(args ...Value) Value {
	if len(args) != 2 {
		fmt.Printf("enumName: wrong number of arguments. got=%d, want=2\n", len(args))
		return NilValue()
	}

	enumTypeName := args[0]
	enumValue := args[1]

	if enumTypeName.Type != StringType {
		fmt.Printf("enumName: first argument must be string (enum type name)\n")
		return NilValue()
	}

	if enumValue.Type != IntType {
		fmt.Printf("enumName: second argument must be int (enum value)\n")
		return NilValue()
	}

	typeName := enumTypeName.AsString()
	value := int(enumValue.AsInt())

	// Look up enum type in registry
	enumType, ok := EnumRegistry[typeName]
	if !ok {
		fmt.Printf("enumName: unknown enum type '%s'\n", typeName)
		return NilValue()
	}

	// Look up variant name
	name, ok := enumType[value]
	if !ok {
		fmt.Printf("enumName: invalid value %d for enum type '%s'\n", value, typeName)
		return NilValue()
	}

	return StringValue(name)
}

// enumValueBuiltin implements enumValue(enumType, name) -> int or error
func enumValueBuiltin(args ...Value) Value {
	if len(args) != 2 {
		fmt.Printf("enumValue: wrong number of arguments. got=%d, want=2\n", len(args))
		return NilValue()
	}

	enumTypeName := args[0]
	variantName := args[1]

	if enumTypeName.Type != StringType {
		fmt.Printf("enumValue: first argument must be string (enum type name)\n")
		return NilValue()
	}

	if variantName.Type != StringType {
		fmt.Printf("enumValue: second argument must be string (variant name)\n")
		return NilValue()
	}

	typeName := enumTypeName.AsString()
	name := variantName.AsString()

	// Look up enum type in registry
	enumType, ok := EnumRegistry[typeName]
	if !ok {
		fmt.Printf("enumValue: unknown enum type '%s'\n", typeName)
		return NilValue()
	}

	// Find variant value by name
	for value, varName := range enumType {
		if varName == name {
			return IntValue(int64(value))
		}
	}

	fmt.Printf("enumValue: unknown variant '%s' for enum type '%s'\n", name, typeName)
	return NilValue()
}

// absBuiltin implements abs(n) - absolute value
func absBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("abs: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	arg := args[0]
	switch arg.Type {
	case IntType:
		val := arg.AsInt()
		if val < 0 {
			return IntValue(-val)
		}
		return arg
	case FloatType:
		val := arg.AsFloat()
		if val < 0 {
			return FloatValue(-val)
		}
		return arg
	default:
		fmt.Printf("abs: argument must be int or float\n")
		return NilValue()
	}
}

// minBuiltin implements min(a, b) - minimum of two numbers
func minBuiltin(args ...Value) Value {
	if len(args) != 2 {
		fmt.Printf("min: wrong number of arguments. got=%d, want=2\n", len(args))
		return NilValue()
	}

	a, b := args[0], args[1]

	// Handle int, int
	if a.Type == IntType && b.Type == IntType {
		if a.AsInt() < b.AsInt() {
			return a
		}
		return b
	}

	// Handle float cases
	if (a.Type == FloatType || a.Type == IntType) && (b.Type == FloatType || b.Type == IntType) {
		var aFloat, bFloat float64
		if a.Type == IntType {
			aFloat = float64(a.AsInt())
		} else {
			aFloat = a.AsFloat()
		}
		if b.Type == IntType {
			bFloat = float64(b.AsInt())
		} else {
			bFloat = b.AsFloat()
		}

		if aFloat < bFloat {
			return FloatValue(aFloat)
		}
		return FloatValue(bFloat)
	}

	fmt.Printf("min: arguments must be int or float\n")
	return NilValue()
}

// maxBuiltin implements max(a, b) - maximum of two numbers
func maxBuiltin(args ...Value) Value {
	if len(args) != 2 {
		fmt.Printf("max: wrong number of arguments. got=%d, want=2\n", len(args))
		return NilValue()
	}

	a, b := args[0], args[1]

	// Handle int, int
	if a.Type == IntType && b.Type == IntType {
		if a.AsInt() > b.AsInt() {
			return a
		}
		return b
	}

	// Handle float cases
	if (a.Type == FloatType || a.Type == IntType) && (b.Type == FloatType || b.Type == IntType) {
		var aFloat, bFloat float64
		if a.Type == IntType {
			aFloat = float64(a.AsInt())
		} else {
			aFloat = a.AsFloat()
		}
		if b.Type == IntType {
			bFloat = float64(b.AsInt())
		} else {
			bFloat = b.AsFloat()
		}

		if aFloat > bFloat {
			return FloatValue(aFloat)
		}
		return FloatValue(bFloat)
	}

	fmt.Printf("max: arguments must be int or float\n")
	return NilValue()
}

// sqrtBuiltin implements sqrt(n) - square root
func sqrtBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("sqrt: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	arg := args[0]
	var val float64

	switch arg.Type {
	case IntType:
		val = float64(arg.AsInt())
	case FloatType:
		val = arg.AsFloat()
	default:
		fmt.Printf("sqrt: argument must be int or float\n")
		return NilValue()
	}

	if val < 0 {
		fmt.Printf("sqrt: argument must be non-negative\n")
		return NilValue()
	}

	// Simple Newton-Raphson implementation
	if val == 0 {
		return FloatValue(0)
	}

	x := val
	for i := 0; i < 20; i++ {
		x = (x + val/x) / 2
	}

	return FloatValue(x)
}

// powBuiltin implements pow(base, exp) - power
func powBuiltin(args ...Value) Value {
	if len(args) != 2 {
		fmt.Printf("pow: wrong number of arguments. got=%d, want=2\n", len(args))
		return NilValue()
	}

	base, exp := args[0], args[1]

	var baseFloat, expFloat float64

	// Convert base
	switch base.Type {
	case IntType:
		baseFloat = float64(base.AsInt())
	case FloatType:
		baseFloat = base.AsFloat()
	default:
		fmt.Printf("pow: base must be int or float\n")
		return NilValue()
	}

	// Convert exponent
	switch exp.Type {
	case IntType:
		expFloat = float64(exp.AsInt())
	case FloatType:
		expFloat = exp.AsFloat()
	default:
		fmt.Printf("pow: exponent must be int or float\n")
		return NilValue()
	}

	// Simple power implementation for integer exponents
	if expFloat == float64(int64(expFloat)) && expFloat >= 0 {
		result := 1.0
		expInt := int64(expFloat)
		for i := int64(0); i < expInt; i++ {
			result *= baseFloat
		}
		return FloatValue(result)
	}

	// For non-integer or negative exponents, we'd need a full math library
	// For now, just handle simple cases
	fmt.Printf("pow: only non-negative integer exponents are supported\n")
	return NilValue()
}

// floorBuiltin implements floor(n) - round down
func floorBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("floor: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	arg := args[0]
	var val float64

	switch arg.Type {
	case IntType:
		return arg // Already an integer
	case FloatType:
		val = arg.AsFloat()
	default:
		fmt.Printf("floor: argument must be int or float\n")
		return NilValue()
	}

	// Manual floor implementation
	intVal := int64(val)
	if val < 0 && val != float64(intVal) {
		intVal--
	}

	return IntValue(intVal)
}

// ceilBuiltin implements ceil(n) - round up
func ceilBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("ceil: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	arg := args[0]
	var val float64

	switch arg.Type {
	case IntType:
		return arg // Already an integer
	case FloatType:
		val = arg.AsFloat()
	default:
		fmt.Printf("ceil: argument must be int or float\n")
		return NilValue()
	}

	// Manual ceil implementation
	intVal := int64(val)
	if val > 0 && val != float64(intVal) {
		intVal++
	}

	return IntValue(intVal)
}

// splitBuiltin implements split(str, separator) - split string into array
func splitBuiltin(args ...Value) Value {
	if len(args) != 2 {
		fmt.Printf("split: wrong number of arguments. got=%d, want=2\n", len(args))
		return NilValue()
	}

	str := args[0]
	sep := args[1]

	if str.Type != StringType {
		fmt.Printf("split: first argument must be string\n")
		return NilValue()
	}

	if sep.Type != StringType {
		fmt.Printf("split: second argument must be string\n")
		return NilValue()
	}

	strVal := str.AsString()
	sepVal := sep.AsString()

	if sepVal == "" {
		// Split into individual characters
		elements := make([]Value, len(strVal))
		for i, ch := range strVal {
			elements[i] = StringValue(string(ch))
		}

		arr := &ArrayValue{Elements: elements}
		arrayPool = append(arrayPool, arr)
		return Value{
			Type: ArrayType,
			Data: uint64(uintptr(unsafe.Pointer(arr))),
		}
	}

	// Split by separator
	var elements []Value
	start := 0
	sepLen := len(sepVal)

	for i := 0; i <= len(strVal)-sepLen; i++ {
		if strVal[i:i+sepLen] == sepVal {
			elements = append(elements, StringValue(strVal[start:i]))
			start = i + sepLen
			i += sepLen - 1
		}
	}
	elements = append(elements, StringValue(strVal[start:]))

	arr := &ArrayValue{Elements: elements}
	arrayPool = append(arrayPool, arr)
	return Value{
		Type: ArrayType,
		Data: uint64(uintptr(unsafe.Pointer(arr))),
	}
}

// substringBuiltin implements substring(str, start, end) - get substring
func substringBuiltin(args ...Value) Value {
	if len(args) != 3 {
		fmt.Printf("substring: wrong number of arguments. got=%d, want=3\n", len(args))
		return NilValue()
	}

	str := args[0]
	start := args[1]
	end := args[2]

	if str.Type != StringType {
		fmt.Printf("substring: first argument must be string\n")
		return NilValue()
	}

	if start.Type != IntType {
		fmt.Printf("substring: second argument must be int\n")
		return NilValue()
	}

	if end.Type != IntType {
		fmt.Printf("substring: third argument must be int\n")
		return NilValue()
	}

	strVal := str.AsString()
	startIdx := int(start.AsInt())
	endIdx := int(end.AsInt())

	// Bounds checking
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > len(strVal) {
		endIdx = len(strVal)
	}
	if startIdx > endIdx {
		startIdx = endIdx
	}

	return StringValue(strVal[startIdx:endIdx])
}

// intBuiltin implements int(x) - convert to int
func intBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("int: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	arg := args[0]

	switch arg.Type {
	case IntType:
		return arg
	case FloatType:
		return IntValue(int64(arg.AsFloat()))
	case BoolType:
		if arg.AsBool() {
			return IntValue(1)
		}
		return IntValue(0)
	case StringType:
		// Simple integer parsing
		str := arg.AsString()
		if str == "" {
			return IntValue(0)
		}

		var result int64
		var negative bool
		start := 0

		if str[0] == '-' {
			negative = true
			start = 1
		}

		for i := start; i < len(str); i++ {
			if str[i] < '0' || str[i] > '9' {
				fmt.Printf("int: invalid integer string '%s'\n", str)
				return NilValue()
			}
			result = result*10 + int64(str[i]-'0')
		}

		if negative {
			result = -result
		}

		return IntValue(result)
	default:
		fmt.Printf("int: cannot convert type to int\n")
		return NilValue()
	}
}

// floatBuiltin implements float(x) - convert to float
func floatBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("float: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	arg := args[0]

	switch arg.Type {
	case FloatType:
		return arg
	case IntType:
		return FloatValue(float64(arg.AsInt()))
	case BoolType:
		if arg.AsBool() {
			return FloatValue(1.0)
		}
		return FloatValue(0.0)
	case StringType:
		// Simple float parsing
		str := arg.AsString()
		if str == "" {
			return FloatValue(0.0)
		}

		var result float64
		var negative bool
		var afterDecimal bool
		var decimalPlaces float64 = 1
		start := 0

		if str[0] == '-' {
			negative = true
			start = 1
		}

		for i := start; i < len(str); i++ {
			if str[i] == '.' {
				if afterDecimal {
					fmt.Printf("float: invalid float string '%s'\n", str)
					return NilValue()
				}
				afterDecimal = true
				continue
			}

			if str[i] < '0' || str[i] > '9' {
				fmt.Printf("float: invalid float string '%s'\n", str)
				return NilValue()
			}

			if afterDecimal {
				decimalPlaces *= 10
				result += float64(str[i]-'0') / decimalPlaces
			} else {
				result = result*10 + float64(str[i]-'0')
			}
		}

		if negative {
			result = -result
		}

		return FloatValue(result)
	default:
		fmt.Printf("float: cannot convert type to float\n")
		return NilValue()
	}
}

// stringBuiltin implements string(x) - convert to string
func stringBuiltin(args ...Value) Value {
	if len(args) != 1 {
		fmt.Printf("string: wrong number of arguments. got=%d, want=1\n", len(args))
		return NilValue()
	}

	// Just use the existing String() method
	return StringValue(args[0].String())
}

// Cached builtin Values to avoid recreating them and growing the pool unnecessarily
var builtinValueCache []Value

// initBuiltinCache initializes the builtin value cache
func init() {
	// Pre-allocate the pool with exact capacity to avoid reallocation
	// This is critical: if the pool reallocates, pointers in builtinValueCache become invalid!
	builtinFunctionPool = make([]interface{}, 0, len(Builtins))
	builtinValueCache = make([]Value, len(Builtins))

	for i := range Builtins {
		fn := Builtins[i]
		// Add to pool to prevent garbage collection
		builtinFunctionPool = append(builtinFunctionPool, fn)
		fnPtr := &builtinFunctionPool[i]  // Use index instead of len-1

		builtinValueCache[i] = Value{
			Type: BuiltinFunctionType,
			Data: uint64(uintptr(unsafe.Pointer(fnPtr))),
		}
	}
}

// getBuiltin returns a built-in function as a Value
func (vm *VM) getBuiltin(index int) Value {
	if index < 0 || index >= len(Builtins) {
		return NilValue()
	}

	// Return the cached value instead of creating a new one each time
	return builtinValueCache[index]
}

// executeBuiltin executes a built-in function
func (vm *VM) executeBuiltin(fn BuiltinFunction, numArgs int) error {
	args := make([]Value, numArgs)
	for i := numArgs - 1; i >= 0; i-- {
		args[i] = vm.pop()
	}

	// Pop the function itself
	vm.pop()

	result := fn(args...)
	return vm.push(result)
}
