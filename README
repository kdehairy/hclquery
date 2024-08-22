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
1. `=`
2. `/`, `:`, `[]` and `{}`

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
