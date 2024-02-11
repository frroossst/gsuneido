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
Exception: Error in MissingAttribute at line 1:  Attribute x not found
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
Exception: Error in Reassign at line 2:  Conflicting inferred types for variable e3c08d1050474cbfbf13c23ff18b8761
existing: SuTypes.String, got: SuTypes.Number
```
