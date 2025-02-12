package json

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kdehairy/hclquery/logging"
	"github.com/zclconf/go-cty/cty"
)

var logger *logging.Logger = logging.NewDefaultLogger()

// ImpliedType returns the cty Type implied by the structure of the given
// JSON-compliant buffer. This function implements the default type mapping
// behavior used when decoding arbitrary JSON without explicit cty Type
// information.
//
// The rules are as follows:
//
// JSON strings, numbers and bools map to their equivalent primitive type in
// cty.
//
// JSON objects map to cty object types, with the attributes defined by the
// object keys and the types of their values.
//
// JSON arrays map to cty list types, with the elements defined by the
// types of the array members.
//
// Any nulls are typed as DynamicPseudoType, so callers of this function
// must be prepared to deal with this. Callers that do not wish to deal with
// dynamic typing should not use this function and should instead describe
// their required types explicitly with a cty.Type instance when decoding.
//
// Any JSON syntax errors will be returned as an error, and the type will
// be the invalid value cty.NilType.
func ImpliedType(buf []byte) (cty.Type, error) {
	logger.Trace("implied", "received", string(buf))
	r := bytes.NewReader(buf)
	dec := json.NewDecoder(r)
	dec.UseNumber()

	ty, err := impliedType(dec)
	if err != nil {
		return cty.NilType, err
	}
	logger.Trace("implied", "returned", ty.GoString())

	if dec.More() {
		return cty.NilType, fmt.Errorf("extraneous data after JSON object")
	}

	return ty, nil
}

func impliedType(dec *json.Decoder) (cty.Type, error) {
	tok, err := dec.Token()
	logger.Trace("implied", "token", tok)
	if err != nil {
		return cty.NilType, err
	}

	return impliedTypeForTok(tok, dec)
}

func impliedTypeForTok(tok json.Token, dec *json.Decoder) (cty.Type, error) {
	if tok == nil {
		return cty.DynamicPseudoType, nil
	}

	switch ttok := tok.(type) {
	case bool:
		return cty.Bool, nil

	case json.Number:
		return cty.Number, nil

	case string:
		return cty.String, nil

	case json.Delim:

		switch rune(ttok) {
		case '{':
			return impliedObjectType(dec)
		case '[':
			return impliedListType(dec)
		default:
			return cty.NilType, fmt.Errorf("unexpected token %q", ttok)
		}

	default:
		return cty.NilType, fmt.Errorf("unsupported JSON token %#v", tok)
	}
}

func impliedObjectType(dec *json.Decoder) (cty.Type, error) {
	// By the time we get in here, we've already consumed the { delimiter
	// and so our next token should be the first object key.
	logger.Trace("implied", "type", "Object")

	var atys map[string]cty.Type

	for {
		// Read the object key first
		tok, err := dec.Token()
		logger.Trace("implied", "token", tok)
		if err != nil {
			return cty.NilType, err
		}

		if ttok, ok := tok.(json.Delim); ok {
			if rune(ttok) != '}' {
				return cty.NilType, fmt.Errorf("unexpected delimiter %q", ttok)
			}
			break
		}

		key, ok := tok.(string)
		logger.Trace("implied", "key", key)
		if !ok {
			return cty.NilType, fmt.Errorf("expected string but found %T", tok)
		}

		// Now read the value
		tok, err = dec.Token()
		logger.Trace("implied", "token", tok)
		if err != nil {
			return cty.NilType, err
		}

		aty, err := impliedTypeForTok(tok, dec)
		logger.Trace("implied", "type", aty.GoString())
		if err != nil {
			return cty.NilType, err
		}

		if atys == nil {
			atys = make(map[string]cty.Type)
		}
		if existing, exists := atys[key]; exists {
			// We didn't originally have any special treatment for multiple properties
			// of the same name, having the type of the last one "win". But that caused
			// some confusing error messages when the same input was subsequently used
			// with [Unmarshal] using the returned object type, since [Unmarshal] would
			// try to fit all of the property values of that name to whatever type
			// the last one had, and would likely fail in doing so if the earlier
			// properties of the same name had different types.
			//
			// As a compromise to avoid breaking existing successful use of _consistently-typed_
			// redundant properties, we return an error here only if the new type
			// differs from the old type. The error message doesn't mention that subtlety
			// because the equal type carveout is a compatibility concession rather than
			// a feature folks should rely on in new code.
			if !existing.Equals(aty) {
				// This error message is low-quality because ImpliedType doesn't do
				// path tracking while it traverses, unlike Unmarshal. However, this
				// is a rare enough case that we don't want to pay the cost of allocating
				// another path-tracking buffer that would in most cases be ignored,
				// so we just accept a low-context error message. :(
				return cty.NilType, fmt.Errorf("duplicate %q property in JSON object", key)
			}
		}
		atys[key] = aty
	}

	if len(atys) == 0 {
		return cty.EmptyObject, nil
	}

	return cty.Object(atys), nil
}

func impliedListType(dec *json.Decoder) (cty.Type, error) {
	// By the time we get in here, we've already consumed the [ delimiter
	// and so our next token should be the first value.
	logger.Trace("implied", "type", "List")

	var etys []cty.Type

	for {
		tok, err := dec.Token()
		if err != nil {
			return cty.NilType, err
		}
		logger.Trace("implied", "token", tok)

		if ttok, ok := tok.(json.Delim); ok {
			if rune(ttok) == ']' {
				break
			}
		}

		ety, err := impliedTypeForTok(tok, dec)
		logger.Trace("implied", "type", ety.GoString())
		if err != nil {
			return cty.NilType, err
		}
		etys = append(etys, ety)
	}

	if len(etys) == 0 {
		return cty.ListValEmpty(cty.NilType).Type(), nil
	}

	ety := etys[0]
	for _, t := range etys {
		if !t.Equals(ety) {
			// if they are not all same type, then return a tuple
			return cty.Tuple(etys), nil
		}
	}

	return cty.List(ety), nil
}
