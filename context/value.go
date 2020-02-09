package context

import (
	"fmt"
	"log"
	"sort"
	"strings"

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

func (vt ValueType) String() string {
	switch vt {
	case ValueTypeInt:
		return ":int"
	case ValueTypeFloat:
		return ":float"
	case ValueTypeSymbol:
		return ":symbol"
	case ValueTypeAtom:
		return ":atom"
	case ValueTypeString:
		return ":string"
	case ValueTypeMap:
		return ":map"
	case ValueTypeList:
		return ":list"
	case ValueTypeFunction:
		return ":func"
	}

	panic("reached")
}

var (
	Nil   = NewAtomValue(":nil")
	True  = NewAtomValue(":true")
	False = NewAtomValue(":false")
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

func (v *Value) SetNode(n *ast.Node) {
	log.Printf("NODE: %v", n)
	v.node = n
}

func (v *Value) Node() *ast.Node {
	return v.node
}

func NewValue(node *ast.Node) (*Value, error) {
	switch node.Type() {
	case ast.NodeTypeInt:
		return NewIntValue(node.Value().(int64)), nil
	case ast.NodeTypeFloat:
		return NewFloatValue(node.Value().(float64)), nil
	case ast.NodeTypeAtom:
		return NewAtomValue(node.Value().(string)), nil
	case ast.NodeTypeSymbol:
		return NewSymbolValue(node.Value().(string)), nil
	case ast.NodeTypeString:
		return NewStringValue(node.Value().(string)), nil
	}
	log.Fatalf("UNHABDLED VALUE: %v", node.Type())
	panic("NEW VALUE")
	return nil, nil
}

func (v Value) Type() ValueType {
	return v.valueType
}

func (v *Value) List() []*Value {
	return v.v.([]*Value)
}

func (v *Value) Atom() string {
	return v.v.(string)
}

func (v *Value) Symbol() string {
	return v.v.(string)
}

func (v *Value) String() string {
	switch v.Type() {
	case ValueTypeMap:
		return encodeMap(v.Map())
	case ValueTypeAtom:
		return v.v.(string)
	case ValueTypeSymbol:
		return v.v.(string)
	case ValueTypeString:
		return fmt.Sprintf("%q", v.v.(string))
	case ValueTypeInt:
		return fmt.Sprintf("%d", v.v)
	case ValueTypeFloat:
		return fmt.Sprintf("%v", v.v)
	case ValueTypeList:
		return encodeList(v.List())
	case ValueTypeFunction:
		return fmt.Sprintf("<function: %v>", v.v)
	}
	panic(fmt.Sprintf("reached: %v", v.Type()))
	return fmt.Sprintf("%v", v.v)
}

func (v *Value) Function() *Function {
	fn := func(ctx *Context) error {
		return v.v.(func(*Context) error)(ctx)
	}
	return NewFunction(fn)
}

func (v *Value) Map() Map {
	return v.v.(map[Value]*Value)
}

func (v *Value) Int() int64 {
	switch v.Type() {
	case ValueTypeInt:
		return v.v.(int64)
	case ValueTypeFloat:
		return int64(v.v.(float64))
	}
	return 0
}

func (v *Value) Float() float64 {
	return 0
}

func Eq(a *Value, b *Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	if a.String() != b.String() {
		return false
	}
	return true
}

func NewArrayValue(v []*Value) *Value {
	return &Value{
		v:         v,
		valueType: ValueTypeList,
	}
}

func NewStringValue(v string) *Value {
	return &Value{
		v:         v,
		valueType: ValueTypeString,
	}
}

func NewSymbolValue(v string) *Value {
	return &Value{
		v:         v,
		valueType: ValueTypeSymbol,
	}
}

func NewAtomValue(v string) *Value {
	return &Value{
		v:         v,
		valueType: ValueTypeAtom,
	}
}

func NewFloatValue(v float64) *Value {
	return &Value{
		v:         v,
		valueType: ValueTypeFloat,
	}
}

func NewListValue(v []*Value) *Value {
	return &Value{
		v:         v,
		valueType: ValueTypeList,
	}
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

type sortableValue []Value

func (sv sortableValue) Len() int {
	return len(sv)
}

func (sv sortableValue) Less(i, j int) bool {
	a, b := sv[i].String(), sv[j].String()
	return a < b
}

func (sv sortableValue) Swap(i, j int) {
	sv[i], sv[j] = sv[j], sv[i]
}

func encodeMap(value map[Value]*Value) string {
	items := []string{}
	keys := sortableValue{}

	for k := range value {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	for _, k := range keys {
		v := value[k]
		items = append(items, fmt.Sprintf("%s %s", k.String(), v.String()))
	}

	return fmt.Sprintf("{%s}", strings.Join(items, " "))
}

func encodeList(values []*Value) string {
	items := []string{}
	for i := range values {
		items = append(items, values[i].String())
	}
	return fmt.Sprintf("[%s]", strings.Join(items, " "))
}
