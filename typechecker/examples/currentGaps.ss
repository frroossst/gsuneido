class {
    x: 42

    Init() {
        // seeing class memeber for the first time as it has no default value
        .count = 5
        .label = "hi"
    }

    // what happens if we type this as return TUknown before we see Init() ?
    GetCount() {
        return .count
    }

    // inline init
    Set_H(.x) {}

    // need an example of such use cases
    Symbol_1() {
        return #foo
    }

    Iterable(ob) {
        for x in ob {
            return x
        }
    }

    Compound() {
        x = false
        x += 1
        return x
    }

    Ternary(c) {
        return c is 1 ? "yes" : 42
    }

    
}
