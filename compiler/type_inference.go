package compiler

import (
	"fmt"
	"minlang/ast"
	"minlang/vm"
)

// convertToValueType converts compiler.Type to vm.ValueType
// This is used for type-specialized opcode generation
func convertToValueType(t Type) vm.ValueType {
	switch typ := t.(type) {
	case *BasicType:
		switch typ.Name {
		case "int":
			return vm.IntType
		case "float":
			return vm.FloatType
		case "bool":
			return vm.BoolType
		case "string":
			return vm.StringType
		default:
			// For struct types defined as BasicType with custom names
			// we don't know the exact type, so default to IntType
			return vm.IntType
		}
	case *ArrayType:
		return vm.ArrayType
	case *MapType:
		return vm.MapType
	case *FunctionType:
		return vm.FunctionType
	}
	// Default to IntType for unknown types
	return vm.IntType
}

// typeAnnotationToValueType converts AST TypeAnnotation to vm.ValueType
func typeAnnotationToValueType(ta *ast.TypeAnnotation) vm.ValueType {
	if ta == nil {
		return vm.IntType // Default
	}

	switch ta.Name {
	case "int":
		return vm.IntType
	case "float":
		return vm.FloatType
	case "bool":
		return vm.BoolType
	case "string":
		return vm.StringType
	}

	// Check if it's an array type
	if ta.ElementType != nil {
		return vm.ArrayType
	}

	// Check if it's a map type
	if ta.KeyType != nil && ta.ValueType != nil {
		return vm.MapType
	}

	// Default to IntType for unknown types (including struct types)
	return vm.IntType
}

// inferExpressionType determines the type of an expression at compile time
// This allows us to emit type-specialized opcodes
func (c *Compiler) inferExpressionType(node ast.Expression) vm.ValueType {
	switch n := node.(type) {
	case *ast.IntegerLiteral:
		return vm.IntType

	case *ast.FloatLiteral:
		return vm.FloatType

	case *ast.BooleanLiteral:
		return vm.BoolType

	case *ast.StringLiteral:
		return vm.StringType

	case *ast.Identifier:
		// Check if we have type information from our type tracking
		if t, ok := c.varTypes[n.Value]; ok {
			return t
		}

		// Check for enum values (they're integers)
		for _, enumType := range c.enumTypes {
			if _, ok := enumType.Variants[n.Value]; ok {
				return vm.IntType
			}
		}

		// Default: we don't know, return IntType as a safe default
		return vm.IntType

	case *ast.InfixExpression:
		return c.inferInfixType(n)

	case *ast.PrefixExpression:
		// -x has the same type as x
		// !x is always bool
		if n.Operator == "!" {
			return vm.BoolType
		}
		return c.inferExpressionType(n.Right)

	case *ast.CallExpression:
		// Function calls - we'd need to track return types
		// For now, default to int
		return vm.IntType

	case *ast.ArrayLiteral:
		return vm.ArrayType

	case *ast.MapLiteral:
		return vm.MapType

	case *ast.StructLiteral:
		return vm.StructType

	default:
		// Unknown type - default to int
		return vm.IntType
	}
}

// inferInfixType determines the result type of an infix expression
func (c *Compiler) inferInfixType(node *ast.InfixExpression) vm.ValueType {
	leftType := c.inferExpressionType(node.Left)
	rightType := c.inferExpressionType(node.Right)

	switch node.Operator {
	case "+":
		// String concatenation takes precedence
		if leftType == vm.StringType || rightType == vm.StringType {
			return vm.StringType
		}
		// Float promotion
		if leftType == vm.FloatType || rightType == vm.FloatType {
			return vm.FloatType
		}
		return vm.IntType

	case "-", "*", "/":
		// Float promotion
		if leftType == vm.FloatType || rightType == vm.FloatType {
			return vm.FloatType
		}
		return vm.IntType

	case "%":
		// Modulo is integer-only
		return vm.IntType

	case "==", "!=", "<", ">", "<=", ">=":
		// Comparisons always return bool
		return vm.BoolType

	case "&&", "||":
		// Logical operations return bool
		return vm.BoolType

	default:
		return vm.IntType
	}
}

// getOperandTypes returns the types of both operands in an infix expression
// This is used to determine which specialized opcode to emit
func (c *Compiler) getOperandTypes(node *ast.InfixExpression) (vm.ValueType, vm.ValueType) {
	leftType := c.inferExpressionType(node.Left)
	rightType := c.inferExpressionType(node.Right)
	return leftType, rightType
}

// inferDetailedType infers the detailed Type (not just vm.ValueType) of an expression
// This is used for type checking
func (c *Compiler) inferDetailedType(node ast.Expression) Type {
	switch n := node.(type) {
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
		// Check if we have detailed type information
		if t, ok := c.typeInfo[n.Value]; ok {
			return t
		}
		return AnyTypeVal

	case *ast.ArrayLiteral:
		if len(n.Elements) == 0 {
			return &ArrayType{ElementType: AnyTypeVal}
		}
		// Infer element type from first element
		elemType := c.inferDetailedType(n.Elements[0])
		return &ArrayType{ElementType: elemType}

	case *ast.MapLiteral:
		if len(n.Pairs) == 0 {
			return &MapType{KeyType: AnyTypeVal, ValueType: AnyTypeVal}
		}
		// Infer key and value types from first pair
		var firstKey ast.Expression
		var firstValue ast.Expression
		for k, v := range n.Pairs {
			firstKey = k
			firstValue = v
			break
		}
		keyType := c.inferDetailedType(firstKey)
		valueType := c.inferDetailedType(firstValue)
		return &MapType{KeyType: keyType, ValueType: valueType}

	case *ast.InfixExpression:
		leftType := c.inferDetailedType(n.Left)
		rightType := c.inferDetailedType(n.Right)

		switch n.Operator {
		case "+":
			// String concatenation
			if leftType.Equals(StringType) || rightType.Equals(StringType) {
				return StringType
			}
			// Float promotion
			if leftType.Equals(FloatType) || rightType.Equals(FloatType) {
				return FloatType
			}
			return IntType

		case "-", "*", "/", "%":
			if leftType.Equals(FloatType) || rightType.Equals(FloatType) {
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
		switch n.Operator {
		case "!":
			return BoolType
		case "-":
			operand := c.inferDetailedType(n.Right)
			return operand
		}

	case *ast.IndexExpression:
		containerType := c.inferDetailedType(n.Left)
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
		// For now, return AnyTypeVal for function calls
		// Would need to track function return types
		return AnyTypeVal
	}

	return AnyTypeVal
}

// checkValueType performs deep type checking for a value against an expected type
func (c *Compiler) checkValueType(node ast.Expression, expectedType Type) error {
	// Check array literals
	if arrLit, ok := node.(*ast.ArrayLiteral); ok {
		if arrType, ok := expectedType.(*ArrayType); ok {
			// Check each element recursively
			for i, elem := range arrLit.Elements {
				// Recursively check if element itself is an array or map
				if err := c.checkValueType(elem, arrType.ElementType); err != nil {
					return fmt.Errorf("array element %d: %v", i, err)
				}
			}
			return nil
		}
	}

	// Check map literals
	if mapLit, ok := node.(*ast.MapLiteral); ok {
		if mapType, ok := expectedType.(*MapType); ok {
			// Check each key-value pair
			for key, value := range mapLit.Pairs {
				keyType := c.inferDetailedType(key)
				if !IsAssignableTo(keyType, mapType.KeyType) {
					return fmt.Errorf("map key has type %s, expected %s",
						keyType.String(), mapType.KeyType.String())
				}

				valueType := c.inferDetailedType(value)
				if !IsAssignableTo(valueType, mapType.ValueType) {
					return fmt.Errorf("map value has type %s, expected %s",
						valueType.String(), mapType.ValueType.String())
				}
			}
			return nil
		}
	}

	// For other expressions, check basic type compatibility
	valueType := c.inferDetailedType(node)
	if !IsAssignableTo(valueType, expectedType) {
		return fmt.Errorf("cannot assign value of type %s to type %s",
			valueType.String(), expectedType.String())
	}

	return nil
}
