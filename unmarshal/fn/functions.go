package fn

import (
	"strings"
	"unicode/utf8"

	"github.com/kdehairy/hclquery/unmarshal/json"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

var JSONDecodeFunc = function.New(&function.Spec{
	Description: `Parses the given string as JSON and returns a value corresponding to what the JSON document describes.`,
	Params: []function.Parameter{
		{
			Name: "str",
			Type: cty.String,
		},
	},
	Type: func(args []cty.Value) (cty.Type, error) {
		str := args[0]
		if !str.IsKnown() {
			// If the string isn't known then we can't fully parse it, but
			// if the value has been refined with a prefix then we may at
			// least be able to reject obviously-invalid syntax and maybe
			// even predict the result type. It's safe to return a specific
			// result type only if parsing a full document with this prefix
			// would return exactly that type or fail with a syntax error.
			rng := str.Range()
			if prefix := strings.TrimSpace(rng.StringPrefix()); prefix != "" {
				// If we know at least one character then it should be one
				// of the few characters that can introduce a JSON value.
				switch r, _ := utf8.DecodeRuneInString(prefix); r {
				case '{', '[':
					// These can start object values and array values
					// respectively, but we can't actually form a full
					// object type constraint or tuple type constraint
					// without knowing all of the attributes, so we
					// will still return DynamicPseudoType in this case.
				case '"':
					// This means that the result will either be a string
					// or parsing will fail.
					return cty.String, nil
				case 't', 'f':
					// Must either be a boolean value or a syntax error.
					return cty.Bool, nil
				case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
					// These characters would all start the "number" production.
					return cty.Number, nil
				case 'n':
					// n is valid to begin the keyword "null" but that doesn't
					// give us any extra type information.
				default:
					// No other characters are valid as the beginning of a
					// JSON value, so we can safely return an early error.
					return cty.NilType, function.NewArgErrorf(0, "a JSON document cannot begin with the character %q", r)
				}
			}
			return cty.DynamicPseudoType, nil
		}

		buf := []byte(str.AsString())
		return json.ImpliedType(buf)
	},
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		buf := []byte(args[0].AsString())
		return ctyjson.Unmarshal(buf, retType)
	},
})
