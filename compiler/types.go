package compiler

import (
	"fmt"
	"minlang/ast"
)

// Type represents a type in the type system
type Type interface {
	String() string
	Equals(other Type) bool
}

// BasicType represents basic types
type BasicType struct {
	Name string
}

func (t *BasicType) String() string {
	return t.Name
}

func (t *BasicType) Equals(other Type) bool {
	if ot, ok := other.(*BasicType); ok {
		return t.Name == ot.Name
	}
	return false
}

// ArrayType represents array types
type ArrayType struct {
	ElementType Type
}

func (t *ArrayType) String() string {
	return "[]" + t.ElementType.String()
}

func (t *ArrayType) Equals(other Type) bool {
	if ot, ok := other.(*ArrayType); ok {
		return t.ElementType.Equals(ot.ElementType)
	}
	return false
}

// MapType represents map types
type MapType struct {
	KeyType   Type
	ValueType Type
}

func (t *MapType) String() string {
	return "map[" + t.KeyType.String() + "]" + t.ValueType.String()
}

func (t *MapType) Equals(other Type) bool {
	if ot, ok := other.(*MapType); ok {
		return t.KeyType.Equals(ot.KeyType) && t.ValueType.Equals(ot.ValueType)
	}
	return false
}

// FunctionType represents function types
type FunctionType struct {
	ParamTypes []Type
	ReturnType Type
}

func (t *FunctionType) String() string {
	params := ""
	for i, p := range t.ParamTypes {
		if i > 0 {
			params += ", "
		}
		params += p.String()
	}
	ret := "void"
	if t.ReturnType != nil {
		ret = t.ReturnType.String()
	}
	return "func(" + params + ") " + ret
}

func (t *FunctionType) Equals(other Type) bool {
	if ot, ok := other.(*FunctionType); ok {
		if len(t.ParamTypes) != len(ot.ParamTypes) {
			return false
		}
		for i := range t.ParamTypes {
			if !t.ParamTypes[i].Equals(ot.ParamTypes[i]) {
				return false
			}
		}
		if t.ReturnType == nil && ot.ReturnType == nil {
			return true
		}
		if t.ReturnType == nil || ot.ReturnType == nil {
			return false
		}
		return t.ReturnType.Equals(ot.ReturnType)
	}
	return false
}

// AnyType represents unknown/any type
type AnyType struct{}

func (t *AnyType) String() string {
	return "any"
}

func (t *AnyType) Equals(other Type) bool {
	_, ok := other.(*AnyType)
	return ok
}

// Common types
var (
	IntType    = &BasicType{Name: "int"}
	FloatType  = &BasicType{Name: "float"}
	BoolType   = &BasicType{Name: "bool"}
	StringType = &BasicType{Name: "string"}
	NilType    = &BasicType{Name: "nil"}
	AnyTypeVal = &AnyType{}
)

// ConvertASTType converts an AST type annotation to a compiler type
func ConvertASTType(astType *ast.TypeAnnotation) Type {
	if astType == nil {
		return AnyTypeVal
	}

	if astType.IsArray {
		return &ArrayType{ElementType: ConvertASTType(astType.ElementType)}
	}

	if astType.IsMap {
		return &MapType{
			KeyType:   ConvertASTType(astType.KeyType),
			ValueType: ConvertASTType(astType.ValueType),
		}
	}

	if astType.IsFunction {
		params := make([]Type, len(astType.ParamTypes))
		for i, p := range astType.ParamTypes {
			params[i] = ConvertASTType(p)
		}
		return &FunctionType{
			ParamTypes: params,
			ReturnType: ConvertASTType(astType.ValueType),
		}
	}

	// Basic type
	switch astType.Name {
	case "int":
		return IntType
	case "float":
		return FloatType
	case "bool":
		return BoolType
	case "string":
		return StringType
	default:
		// Unknown type, treat as any
		return AnyTypeVal
	}
}

// IsAssignableTo checks if a value of type 'from' can be assigned to 'to'
func IsAssignableTo(from, to Type) bool {
	// Any type can be assigned to any
	if _, ok := to.(*AnyType); ok {
		return true
	}
	if _, ok := from.(*AnyType); ok {
		return true
	}

	// Nil can be assigned to any reference type
	if fromBasic, ok := from.(*BasicType); ok {
		if fromBasic.Name == "nil" {
			// Can assign nil to arrays, maps, etc
			return true
		}
	}

	// Int can be promoted to float
	if fromBasic, ok := from.(*BasicType); ok {
		if toBasic, ok2 := to.(*BasicType); ok2 {
			if fromBasic.Name == "int" && toBasic.Name == "float" {
				return true
			}
		}
	}

	return from.Equals(to)
}

// TypeChecker performs type checking
type TypeChecker struct {
	symbolTable *SymbolTable
	errors      []string
	typeMap     map[string]Type // Maps variable names to their types
}

// NewTypeChecker creates a new type checker
func NewTypeChecker() *TypeChecker {
	return &TypeChecker{
		symbolTable: NewSymbolTable(),
		errors:      []string{},
		typeMap:     make(map[string]Type),
	}
}

// AddError adds a type error
func (tc *TypeChecker) AddError(msg string) {
	tc.errors = append(tc.errors, msg)
}

// Errors returns all type errors
func (tc *TypeChecker) Errors() []string {
	return tc.errors
}

// InferType infers the type of an expression
func (tc *TypeChecker) InferType(node ast.Expression) Type {
	switch node := node.(type) {
	case *ast.IntegerLiteral:
		return IntType

	case *ast.FloatLiteral:
		return FloatType

	case *ast.BooleanLiteral:
		return BoolType

	case *ast.StringLiteral:
		return StringType

	case *ast.NilLiteral:
		return NilType

	case *ast.Identifier:
		if t, ok := tc.typeMap[node.Value]; ok {
			return t
		}
		return AnyTypeVal

	case *ast.ArrayLiteral:
		if len(node.Elements) == 0 {
			return &ArrayType{ElementType: AnyTypeVal}
		}
		// Infer from first element
		elemType := tc.InferType(node.Elements[0])
		return &ArrayType{ElementType: elemType}

	case *ast.InfixExpression:
		left := tc.InferType(node.Left)
		right := tc.InferType(node.Right)

		switch node.Operator {
		case "+":
			// String concatenation
			if left.Equals(StringType) || right.Equals(StringType) {
				return StringType
			}
			// Float promotion
			if left.Equals(FloatType) || right.Equals(FloatType) {
				return FloatType
			}
			return IntType

		case "-", "*", "/", "%":
			if left.Equals(FloatType) || right.Equals(FloatType) {
				return FloatType
			}
			return IntType

		case "==", "!=", "<", ">", "<=", ">=":
			return BoolType

		case "&&", "||":
			return BoolType

		default:
			return AnyTypeVal
		}

	case *ast.PrefixExpression:
		switch node.Operator {
		case "!":
			return BoolType
		case "-":
			operand := tc.InferType(node.Right)
			return operand
		}

	case *ast.IndexExpression:
		containerType := tc.InferType(node.Left)
		if arrayType, ok := containerType.(*ArrayType); ok {
			return arrayType.ElementType
		}
		if mapType, ok := containerType.(*MapType); ok {
			return mapType.ValueType
		}
		if containerType.Equals(StringType) {
			return StringType
		}
		return AnyTypeVal

	case *ast.CallExpression:
		// For now, assume functions return any type
		// Would need to track function signatures
		return AnyTypeVal
	}

	return AnyTypeVal
}

// CheckVarStatement checks a variable statement
func (tc *TypeChecker) CheckVarStatement(stmt *ast.VarStatement) {
	var declaredType Type
	if stmt.Type != nil {
		declaredType = ConvertASTType(stmt.Type)
	} else {
		declaredType = AnyTypeVal
	}

	if stmt.Value != nil {
		valueType := tc.InferType(stmt.Value)

		// Check if value type is assignable to declared type
		if !IsAssignableTo(valueType, declaredType) && !declaredType.Equals(AnyTypeVal) {
			tc.AddError(fmt.Sprintf("cannot assign value of type %s to variable %s of type %s",
				valueType.String(), stmt.Name.Value, declaredType.String()))
		}

		// Store the actual type (or declared if more specific)
		if declaredType.Equals(AnyTypeVal) {
			tc.typeMap[stmt.Name.Value] = valueType
		} else {
			tc.typeMap[stmt.Name.Value] = declaredType
		}
	} else {
		tc.typeMap[stmt.Name.Value] = declaredType
	}
}

// CheckAssignment checks an assignment
func (tc *TypeChecker) CheckAssignment(stmt *ast.AssignmentStatement) {
	valueType := tc.InferType(stmt.Value)

	if ident, ok := stmt.Left.(*ast.Identifier); ok {
		if varType, exists := tc.typeMap[ident.Value]; exists {
			if !IsAssignableTo(valueType, varType) {
				tc.AddError(fmt.Sprintf("cannot assign value of type %s to variable %s of type %s",
					valueType.String(), ident.Value, varType.String()))
			}
		}
	}
}
