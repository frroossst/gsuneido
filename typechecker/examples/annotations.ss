class {
    // Simplest case: lowercase primitive, single alternative.
    Add(a: number, b: number) : number {
        return a + b
    }

    // Union annotation: parameter accepts either type. The `$`
    // operator is string concat (use `+` and you get a Number, which
    // the verify pass would catch since the return is declared string).
    Stringify(x: number | string) : string {
        if String?(x) { return x }
        return "n=" $ x
    }

    // Annotation seeds the param when there's no default value to
    // derive from. Without the annotation, x would be TUnknown.
    Identity(x: object) {
        return x   // returns TObject thanks to the annotation
    }

    // Annotation overrides the default-value type. The default below
    // is a Number literal but the user has declared String. The
    // checker honours the declaration and emits a stderr diagnostic
    // for the inconsistent default.
    OverrideDefault(x: string = 0) {
        return x
    }

    // Capitalized primitive name. Both casings work; `Number` is the
    // form Suneido uses elsewhere (predicates, `Type(x)` strings).
    Sum(xs: Object) : Number {
        return xs.Sum()
    }

    // User class name. We don't yet track per-class types, so `Point`
    // resolves to TObject - good enough for method-call inference.
    Origin() : Point {
        return Object(x: 0, y: 0)
    }

    // Return annotation pinning. The body always returns a Number, but
    // the declared return type is `number | string` so callers must
    // handle both branches. The annotation is the contract.
    LookupName(id: number) : string | number {
        return id
    }

    // Return annotation works through narrowing. With `(x: number | string)`,
    // a guarded body returns x of the narrowed type, and the function's
    // public return is the annotated `number` (callers see the contract).
    Round(x: number | string) : number {
        if Number?(x) { return x }
        return 0
    }

    // Inline-init param with annotation seeds the matching class member
    // from the annotation. No boolean-widening here: when the type is
    // explicitly `false`, the user means exactly TFalse.
    Configure(.disabled: false) {
        return .disabled
    }

    // Variadic `@args` parameter. The parser implicitly annotates it
    // as `object`, so args is typed as TObject.
    Sum2(@args) {
        return args.Sum()
    }
}
