package context

import (
	"github.com/xiam/sexpr/ast"
)

type Function struct {
	name string
	fn   func(*Context) error
}

func NewFunction(fn func(*Context) error) *Function {
	return &Function{
		fn: fn,
	}
}

func NewFunctionWithName(fn func(*Context) error, name string) *Function {
	f := NewFunction(fn)
	f.name = name
	return f
}

func (f Function) Exec(ctx *Context) error {
	return f.fn(ctx)
}

func (f Function) Name() string {
	return f.name
}

type Map map[Value]*Value

type ValueType uint8

const (
	valueTypeNone ValueType = iota

	ValueTypeInt
	ValueTypeFloat
	ValueTypeSymbol
	ValueTypeAtom
	ValueTypeString
	ValueTypeMap
	ValueTypeList
	ValueTypeFunction
)

var (
	Nil   = fromValuer(ast.NewAtomValue(":nil"))
	True  = fromValuer(ast.NewAtomValue(":true"))
	False = fromValuer(ast.NewAtomValue(":false"))
)

type Value struct {
	node      *ast.Node
	valueType ValueType
	v         interface{}

	name string
}

var valuerTypeMap = map[ast.NodeType]ValueType{
	ast.NodeTypeInt:    ValueTypeInt,
	ast.NodeTypeFloat:  ValueTypeFloat,
	ast.NodeTypeSymbol: ValueTypeSymbol,
	ast.NodeTypeAtom:   ValueTypeAtom,
	ast.NodeTypeString: ValueTypeString,
	ast.NodeTypeMap:    ValueTypeMap,
	ast.NodeTypeList:   ValueTypeList,
}

func valuerType(t ast.NodeType) ValueType {
	if v, ok := valuerTypeMap[t]; ok {
		return v
	}
	return valueTypeNone
}

func fromValuer(v ast.Valuer) *Value {
	return &Value{
		valueType: valuerType(v.Type()),
		v:         v,
	}
}

func NewValue(node *ast.Node) (*Value, error) {
	return nil, nil
}

func (v Value) Type() ValueType {
	return v.valueType
}

func (v *Value) List() []*Value {
	return v.v.([]*Value)
}

func (v *Value) Symbol() string {
	return v.v.(string)
}

func (v *Value) String() string {
	return v.v.(string)
}

func (v *Value) Function() *Function {
	fn := func(*Context) error {
		panic("called")
		return nil
	}
	return NewFunction(fn)
}

func (v *Value) Int() int64 {
	switch v.valueType {
	case ValueTypeInt:
		return v.node.Value().(int64)
	case ValueTypeFloat:
		return int64(v.node.Value().(float64))
	}
	return 0
}

func (v *Value) Float() float64 {
	return 0
}

func Eq(a *Value, b *Value) bool {
	return false
}

func NewIntValue(v int64) *Value {
	return &Value{
		v:         v,
		valueType: ValueTypeInt,
	}
}

func NewMapValue(v map[Value]*Value) *Value {
	return &Value{
		v:         v,
		valueType: ValueTypeMap,
	}
}

func NewFunctionValue(fn func(*Context) error) *Value {
	return &Value{
		v:         fn,
		valueType: ValueTypeFunction,
	}
}
