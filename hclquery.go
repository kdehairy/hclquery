/*
The purpose of the hclquery package is simple; given a hcl body and a query
string, it executes the query against the body and returns back a list of
blocks that satisfies the query.

# The Query Grammer

	Expr         ::= Segment ( '/' Segment )*

	Segment      ::= Ident

		| Ident '{' Predicate '}'
		| Ident '{' Predicate '}' '[' NUM ']'
		| Block
		| Block '{' Predicate '}'

	Block        ::= Ident '[' NUM ']'

		| Ident ':' Ident

	Predicate    ::= Ident

		| Ident '=' Literal

	Literal      ::= ”' CHARACTERS ”'

		| '"' CHARACTERS '"'

# Examples

1. `terraform`

find a block of type `terraform`.

	terraform {
	  ...
	}

2. `terraform/required_providers`

find a block of type `provider` that is nested inside a block of type `terraform`.

	terraform {
	  ...
	  required_providers {
	    aws = {
	      source  = "hashicorp/aws"
	      version = ">= 5.11.0"
	    }
	  }
		...
	}

3. `terraform/backend:s3`

find a block of type `backend` and label `s3` that is nested inside a block of type `terraform`

	terraform {
	  backend "s3" {
	    ...
	  }
		...
	}

4. `terraform/backend:s3{region}`

find a block of type `backend` with a label `s3` and has an attribute called `region`.

	terraform {
	  backend "s3" {
			...
	    region = "eu-west-2"
			...
	  }
		...
	}

5. `terraform/backend:s3{region='eu-west-2'}`

find a block of type `backend` with a label `s3` and has an attribute called `region` with a value of `eu-west-2`

	terraform {
	  backend "s3" {
			...
	    region = "eu-west-2"
			...
	  }
		...
	}
*/
package hclquery

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Given a hcl body, and a query string, it returns an array of hcl blocks that
// matches the given query.
func Query(b hcl.Body, path string) (hclsyntax.Blocks, error) {
	body := b.(*hclsyntax.Body)
	logger.Debug("Body received", "body", b, "type", reflect.TypeOf(b))
	compilation, err := Compile(path)
	if err != nil {
		return nil, err
	}
	blocks := body.Blocks
	if len(blocks) == 0 {
		logger.Debug("No blocks in the passed body")
		return hclsyntax.Blocks{}, nil
	}
	return compilation.Exec(blocks)
}
