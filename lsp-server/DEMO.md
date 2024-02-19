# Undeclared Class Variable
```
class()
    {
    MissingAttribute(x) 
        {
        .x = x
        }
    }
```
```
TypeError: Error in MissingAttribute at line 1:  Attribute x not found
```

# Variable reassignment
```
class()
    {
    Reassign()
        {
        x = "hello"
        x = 123
        }
    }
```
```
TypeError: Error in Reassign at line 2:  Conflicting inferred types for variable e3c08d1050474cbfbf13c23ff18b8761
existing: SuTypes.String, got: SuTypes.Number
```

# Parameter mismatch
```
// typdefinition for IncorrectParam in a separate file
type IncorrectParam >>= fn(x: Number) -> None
```
```
class()
    {
    IncorrectParam(x)
        {
        x = "IAmAString"
        }
    }
```

```
TypeError: Conflicting inferred types for variable 17650a1119d644b3817651625465b494
existing: SuTypes.String, got: SuTypes.Number
```
