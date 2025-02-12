package unmarshal

import (
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
		return nil, fmt.Errorf("no attribute with name '%v' found", name)
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
		logger.Debug("elm", "type", elm.Type().GoString())
		vals = append(vals, elm)
	}
	return cty.ListVal(vals)
}

func (a *Attr) To(obj interface{}, ctx *hcl.EvalContext) error {
	val, _ := a.attr.Expr.Value(ctx)
	logger.Debug(fmt.Sprintf("attribute type is: '%v'", val.Type().GoString()))

	target := trueType(reflect.ValueOf(obj))
	logger.Debug(fmt.Sprintf("target type is: '%v'", target.Type()))
	err := gocty.FromCtyValue(val, obj)
	if err != nil {
		return fmt.Errorf("failed to parse value into %v: %v", reflect.TypeOf(obj), err)
	}

	return nil
}
