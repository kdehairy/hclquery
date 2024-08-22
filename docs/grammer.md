# Grammer

## BNF

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
