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

```terraform
terraform {
  ...
}
```

`terraform/required_providers`
find a block of type `provider` that is nested inside a block of type `terraform`.
```terraform
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
```terraform
terraform {
  backend "s3" {
    ...
  }
	...
}
```

`terraform/backend:s3{region}`
find a block of type `backend` with a label `s3` and has an attribute called `region`.

```terraform
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

```terraform
terraform {
  backend "s3" {
		...
    region = "eu-west-2"
		...
  }
	...
}
```

# Usage
Assuming the followning HCL content:

```terraform
resource "aws_vpc_endpoint" "this" {
  vpc_id            = "123"
  dynamic "dns_options" {
    for_each = try([each.value.dns_options], [])

    content {
      dns_record_ip_type                             = try(dns_options.value.dns_options.dns_record_ip_type, null)
      private_dns_only_for_inbound_resolver_endpoint = try(dns_options.value.private_dns_only_for_inbound_resolver_endpoint, null)
    }
  }
}
```
Then...

```go
// getting the 'backend "S3"' block
blocks, err := QueryFile("path/to/hcl/file.tf", "resource:aws_vpc_endpoint{vpc_id='123'}")
if err != nil {
	t.Fatalf("failed to find block: %v", err)
}

// unmarshalling the "region" attribute
wrapped_block := unmarshal.New(block)
var str string
attr, err := wrapped_block.GetAttr("vpc_id")
attr.To(&str, nil)
```

# Unmarshalling
After landing on the desired block, we can unmarshal the values of any attribute inside such a bock to go native types.

## Examples
In a terraform config like this:
```terraform
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
```terraform
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
