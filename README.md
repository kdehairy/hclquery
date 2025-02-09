# Overview
The purpose of the hclquery package is simple; given a hcl body and a query
string, it executes the query against the body and returns back a list of
blocks that satisfies the query.

# The Query
### Grammer

```
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
Literal      ::= ''' CHARACTERS '''
               | '"' CHARACTERS '"'
```

### Precedence
1. `/`, `:`, `[]` and `{}`
2. `=`

### Associativity
- `/`, `:`, `[]` and `{}` are left-associative.
- `=` is right-associative.

### Examples
`terraform`
find a block of type `terraform`.

```
terraform {
  ...
}
```

`terraform/required_providers`
find a block of type `provider` that is nested inside a block of type `terraform`.
```
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
```

`terraform/backend:s3`
find a block of type `backend` and label `s3` that is nested inside a block of type `terraform`
```
terraform {
  backend "s3" {
    ...
  }
	...
}
```

`terraform/backend:s3{region}`
find a block of type `backend` with a label `s3` and has an attribute called `region`.

```
terraform {
  backend "s3" {
		...
    region = "eu-west-2"
		...
  }
	...
}
```

`terraform/backend:s3{region='eu-west-2'}`
find a block of type `backend` with a label `s3` and has an attribute called `region` with a value of `eu-west-2`

```
terraform {
  backend "s3" {
		...
    region = "eu-west-2"
		...
  }
	...
}
```

# Unmarshalling
After landing on the desired block, we can unmarshal the values of any attribute inside such a bock to go native types.

## Examples
In a terraform config like this:
```
terraform {
  backend "s3" {
		...
    region = "eu-west-2"
		...
  }
	...
}
```

Doing a query on `terraform/backend:s3{region}`, we would get the following block:
```
backend "s3" {
  	...
  region = "eu-west-2"
  	...
}
```

To unmarshal the value of the `region` attribute, we would do:
```go
// `block` is the block returned from the query
wrapped_block := unmarshal.New(block)
var str string
attr, err := wrapped_block.GetAttr("region")
attr.To(&str, nil)
```
