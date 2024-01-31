# 23 November 2023

- basic python type inference
- did this in python as a proof of concept, future version in OCaml

```
function(x)
	{
	num = x + "123"
	num()
	}`
```
The above now throws a compile time error, as the type of num is inferred to be a string, and strings are not callable.
```
TypeError: For operator Add, "123" is not of type Number, it is of type String
```
Before it'd throw a runtime error,
```
ERROR uncaught in repl: can't convert String to number
```

```
function(x)
    {
    num = x + 123
    num()
    }`
```

```
TypeError: For operator Call, num is not of type Function, it is of type Number
```
Before it'd throw a runtime error,
```
ERROR uncaught in repl: can't call Number
```

Proof of concept shows that basic type inference is possible on Suneido without extensive changes to the language.

## Next steps
- [ ] parse and infer all primitive types  
- [ ] better error messages with line numbers and possible fixes
- [ ] all primitive parsing should be done in Go to avoid the need for yet another language and parser impl  
- [ ] start writing an OCaml inference engine  
- [ ] work on proof of soundness and completeness in Isabelle or Idris  
- [ ] start writing the white papaer  
- [ ] impl basic lsp in either vscode or nvim

# 22 December 2023

## Architecture

Every class has a key value store, this key value store stores the member identifier name has a key and the value is another
key value store that stores the relevant type information of it's scoped members/variables.

When the key value stores are built, they contain the primitive inferred types of the members/variables. 

Type inference uses the key value stores to infer the types of the members/variables, no type checking is done at this stage.

Type cheking is then later done to constriant solve the inferred types

## TODO

- [ ] Implement typeFunction() parsing
- [ ] Implement typeClass() parsing
- [ ] Implement type key-value store

# 30 January 2024

Completed complete parsing of a class, including attributes and methods.

## TODO

- [ ] Implement type inference and graph construction
- [ ] Implement type checking on graph constraints
- [ ] Implement object structure parsing
