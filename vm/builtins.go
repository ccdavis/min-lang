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
