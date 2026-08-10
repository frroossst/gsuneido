class {
    // OK. number <: number - silent.
    AddOk(a: number, b: number) : number {
        return a + b
    }

    // OK. number <: number|string and string <: number|string - silent
    // at both return sites.
    EitherOk(flag: boolean) : number | string {
        if flag { return 1 }
        return "zero"
    }

    // ERROR pinned at `return x + 1`. Param-as-ground-truth makes x: number
    // authoritative; the `+` operator forces the result to number; declared
    // return is string, so number </: string is a provable violation.
    BadAdd(x: number) : string {
        return x + 1
    }

    // ERROR pinned at `return 42`. Literal 42 is TNumber, declared is
    // string, definite mismatch; no narrowing or coercion can save it.
    BadLiteral() : string {
        return 42
    }

    // ONE ERROR + ONE OK. The two return sites are diagnosed independently,
    // so the bad branch is pinned and the good branch stays silent. The
    // function still publishes `string` as its public return type.
    Mixed(flag: boolean) : string {
        if flag { return 42 }   // ERROR - 42 not subset of string
        return "hi"             // OK
    }

    // WARNING pinned at `return GetMystery()`. The callee is unknown to
    // the type system, so its result type is TUnknown we cannot prove
    // that it fits `string|number`, but we also cannot prove it doesn't.
    Ambiguous() : string | number {
        return GetMystery()
    }

    // WARNING pinned at `return z`. One branch produces a string, the
    // other an unknown call. The folded type is `string | ?`. The known
    // member fits, but the dirty tail might not warn.
    DirtyUnion(flag: boolean) : string | number {
        z = flag ? "hi" : GetMystery()
        return z
    }

    // ERROR + WARNING are independent: a provable boolean leak gets an
    // error pinned at its own site even though the other branch is only
    // warnable. Both diagnostics are emitted.
    LeakAndAmbiguous(flag: boolean) : string | number {
        if flag { return true }            // ERROR - boolean not subset of string|number
        return GetMystery()                // WARNING
    }
}
