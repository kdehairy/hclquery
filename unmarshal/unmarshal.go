package unmarshal

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/kdehairy/hclquery/logging"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

var logger = logging.NewDefaultLogger()

type Block struct {
	block *hclsyntax.Block
}

type Attr struct {
	attr *hcl.Attribute
}

func New(block *hclsyntax.Block) Block {
	return Block{
		block: block,
	}
}

func (b Block) GetAttr(name string) (*Attr, error) {
	attrs, _ := b.block.Body.JustAttributes()
	attr, ok := attrs[name]
	if !ok {
		return nil, fmt.Errorf("No attribute with name '%v' found", name)
	}

	return &Attr{attr}, nil
}

func trueType(val reflect.Value) reflect.Value {
	if val.Kind() != reflect.Ptr {
		return val
	}

	return val.Elem()
}

func tubleToList(val cty.Value) cty.Value {
	var vals []cty.Value
	it := val.ElementIterator()
	for it.Next() {
		_, elm := it.Element()
		vals = append(vals, elm)
	}
	return cty.ListVal(vals)
}

func (a *Attr) To(obj interface{}, ctx *hcl.EvalContext) error {
	val, _ := a.attr.Expr.Value(ctx)
	logger.Debug(fmt.Sprintf("attribute value is: '%v'", val))

	target := trueType(reflect.ValueOf(obj))
	logger.Debug(fmt.Sprintf("target type is: '%v'", target.Type()))
	if val.Type().IsTupleType() &&
		(target.Kind() == reflect.Array || target.Kind() == reflect.Slice) {
		logger.Debug("attempting to convert parsed value from tuble to array...")
		elmIt := val.ElementIterator()
		elmIt.Next()
		_, elm := elmIt.Element()
		elmType := elm.Type()
		for elmIt.Next() {
			_, elm := elmIt.Element()
			if elmType != elm.Type() {
				logger.Error("Incompatible types", "value", val.Type(), "target", target.Type())
				return errors.New("parsed value is of Tuble type, but target is of an Array or Slice type")
			}
		}
		val = tubleToList(val)
		logger.Debug("parsed value is converted.", "new type", val.Type())
	}

	err := gocty.FromCtyValue(val, obj)
	if err != nil {
		return fmt.Errorf("failed to parse value into %v: %v", reflect.TypeOf(obj), err)
	}

	return nil
}
