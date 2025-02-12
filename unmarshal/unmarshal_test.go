package unmarshal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/kdehairy/hclquery"
	"github.com/kdehairy/hclquery/unmarshal/json"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type TestCase struct {
	expected interface{}
	test     func(block *hclsyntax.Block, attr string) (bool, error)
	name     string
	fixture  string
	block    string
	attr     string
}

type AttrObj struct {
	A string  `cty:"a"`
	B string  `cty:"b"`
	C string  `cty:"c"`
	D *string `cty:"d"`
}

func TestObjType(t *testing.T) {
	cases := []TestCase{
		{
			name:    "block with object type attribute",
			fixture: "test-1.tf",
			block:   "module",
			attr:    "attr_obj",
			test: func(input *hclsyntax.Block, name string) (bool, error) {
				block := New(input)
				var obj AttrObj
				attr, err := block.GetAttr(name)
				attr.To(&obj, &hcl.EvalContext{
					Functions: map[string]function.Function{
						"upper": stdlib.UpperFunc,
					},
				})
				if err != nil {
					return false, fmt.Errorf("failed to parse obj: %v", err)
				}
				if obj.A != "A" {
					return false, nil
				}
				if obj.B != "b" {
					return false, nil
				}
				if obj.C != "c" {
					return false, nil
				}
				if obj.D != nil {
					return false, nil
				}
				return true, nil
			},
		},
		{
			name:    "attribute with json object",
			fixture: "test-1.tf",
			block:   "module/block1",
			attr:    "jsonAttr",
			test: func(input *hclsyntax.Block, name string) (bool, error) {
				block := New(input)
				var obj struct {
					Name  string `cty:"name"`
					Image string `cty:"image"`
				}
				attr, err := block.GetAttr(name)
				if err != nil {
					return false, err
				}
				attr.To(&obj, &hcl.EvalContext{
					Functions: map[string]function.Function{
						"jsondecode": json.JSONDecodeFunc,
					},
				})
				if obj.Name != "datetime" {
					return false, nil
				}
				if obj.Image != "datetime-image-path" {
					return false, nil
				}
				return true, nil
			},
		},
		{
			name:    "attribute with json array",
			fixture: "test-1.tf",
			block:   "module/block2",
			attr:    "jsonAttr",
			test: func(input *hclsyntax.Block, name string) (bool, error) {
				block := New(input)
				var obj []struct {
					Name  string `cty:"name"`
					Image string `cty:"image"`
				}
				attr, err := block.GetAttr(name)
				if err != nil {
					return false, err
				}
				err = attr.To(&obj, &hcl.EvalContext{
					Functions: map[string]function.Function{
						"jsondecode": json.JSONDecodeFunc,
					},
				})
				if err != nil {
					t.Fatalf("failed to parse into provided obj: %v", err)
				}
				if obj[0].Name != "datetime" {
					return false, nil
				}
				if obj[0].Image != "datetime-image-path" {
					return false, nil
				}
				return true, nil
			},
		},
		{
			name:    "attribute with nested json array",
			fixture: "test-1.tf",
			block:   "module/block3",
			attr:    "jsonAttr",
			test: func(input *hclsyntax.Block, name string) (bool, error) {
				block := New(input)
				var obj struct {
					Name  string `cty:"name"`
					Image string `cty:"image"`
					Env   []struct {
						Name  string `cty:"name"`
						Value string `cty:"value"`
					} `cty:"env"`
				}
				attr, err := block.GetAttr(name)
				if err != nil {
					return false, err
				}
				err = attr.To(&obj, &hcl.EvalContext{
					Functions: map[string]function.Function{
						"jsondecode": json.JSONDecodeFunc,
					},
				})
				if err != nil {
					t.Fatalf("failed to parse into provided obj: %v", err)
				}
				if obj.Name != "datetime" {
					return false, nil
				}
				if obj.Image != "datetime-image-path" {
					return false, nil
				}
				if len(obj.Env) != 1 {
					return false, nil
				}
				return true, nil
			},
		},
	}

	for _, tc := range cases {
		testName := strings.ReplaceAll(tc.name, " ", "_")
		testName = strings.ToLower(testName)
		t.Run(testName, func(t *testing.T) {
			hclParser := hclparse.NewParser()
			hclFile, _ := hclParser.ParseHCLFile("test_cases/test-1.tf")
			if hclFile == nil {
				t.Fatalf("failed to parse hcl file")
			}

			blocks, err := hclquery.QueryFile("test_cases/test-1.tf", tc.block)
			if err != nil {
				t.Fatalf("failed to find block:%v", err)
			}

			if len(blocks) != 1 {
				t.Fatalf("expected 1 but found %v blocks", len(blocks))
			}
			passed, err := tc.test(blocks[0], tc.attr)
			if err != nil {
				t.Fatalf("failed: %v", err)
			}

			if !passed {
				t.Error("Not the expected Object")
			}
		})
	}
}

func TestStringType(t *testing.T) {
	cases := []TestCase{
		{
			name:     "block with both sub-blocks and attrs",
			fixture:  "test-1.tf",
			block:    "module",
			attr:     "attr01",
			expected: "x",
		},
		{
			name:     "sub-block",
			fixture:  "test-1.tf",
			block:    "module/block1",
			attr:     "attr11",
			expected: "a",
		},
	}

	for _, tc := range cases {
		testName := strings.ReplaceAll(tc.name, " ", "_")
		testName = strings.ToLower(testName)
		t.Run(testName, func(t *testing.T) {
			blocks, err := hclquery.QueryFile("test_cases/test-1.tf", tc.block)
			if err != nil {
				t.Fatalf("failed to find block:%v", err)
			}

			if len(blocks) == 0 {
				t.Fatal("expecting at least 1  block, got 0")
			}

			for _, b := range blocks {
				block := New(b)
				var str string
				attr, err := block.GetAttr(tc.attr)
				attr.To(&str, nil)
				if err != nil {
					t.Fatalf("error while reading attribute: %v", err)
				}
				if str != tc.expected.(string) {
					t.Errorf("expected '%v' but found '%v'", tc.expected.(string), str)
				}
			}
		})
	}
}
