// Per-operand contract checking for operators (TypeCheckPass).
//
// Same per-site philosophy as the return-contract checker: every operand
// is decomposed into union members and classified independently.
//
//   - proven non-fit                 -> ERROR
//   - TUnknown / dirty union member  -> WARNING
//   - proven fit                     -> silent
//
// Scope today (intentionally tight):
//
//     numeric:  +  -  *  /  %  &  |  ^  <<  >>  unary -  ~
//               all *Eq compound forms; ++ --
//     string:   $   $=
//
// Loose operators (`is`/`isnt`, `<`/`>`, `and`/`or`/`not`) are NOT
// checked - they accept too many shapes for a useful contract.
class {
    // OK. Both operands proven Number - silent.
    AddOk(a: number, b: number) : number {
        return a + b
    }

    // OK. Both operands proven String - silent.
    ConcatOk(a: string, b: string) : string {
        return a $ b
    }

    // ERROR pinned at `s`. `+` requires Number on both sides; param
    // annotation makes `s: string` authoritative, so this is a provable
    // mismatch.
    BadAddString(s: string, n: number) : number {
        return s + n
    }

    // TWO ERRORS, one per offending operand. The `1` literal is fine,
    // but both `s` and `t` are proven String against a numeric `*`.
    TwoBad(s: string, t: string) : number {
        return s * t + 1
    }

    // ERROR pinned at the numeric operand of `$`. String concat requires
    // String on both sides; n: number is a provable mismatch.
    BadConcat(s: string, n: number) : string {
        return s $ n
    }

    // ERROR. Unary `-` requires Number; flag: boolean does not fit.
    BadUnary(flag: boolean) : number {
        return -flag
    }

    // ERROR pinned at `"x"`. Compound `+=` is numeric. The RHS literal
    // is a String, which is a provable mismatch. The LHS `s` does NOT
    // fire because existing inference rewrites its type to the operator
    // result (Number) - a pre-existing quirk of compound-assign
    // handling, out of scope for this pass.
    BadCompound(s: string) {
        s += "x"
    }

    // WARNING. `n` resolves to TUnknown (no annotation, no default), so
    // we cannot prove either operand of `+` is Number - emit warnings
    // on the unknown side. The literal `1` is silent (proven Number).
    AmbiguousAdd(n) : number {
        return n + 1
    }

    // WARNING. `?` dirty union: x is "hi" or a result from an unknown
    // call. The known String member fails against numeric `+`, so this
    // is actually an ERROR for the known member AND the call result is
    // unknown - error dominates.
    DirtyAdd(flag: boolean) : number {
        x = flag ? "hi" : GetMystery()
        return x + 1
    }

    // OK after narrowing. The guard proves x is Number inside the
    // branch, so `x + 1` operates on a clean Number - silent.
    NarrowedOk(x: number | string) : number {
        if Number?(x) { return x + 1 }
        return 0
    }

    // OK. `is`/`isnt` and comparison operators are intentionally NOT
    // checked. This stays silent even though we're comparing a string
    // and a number - Suneido's runtime is permissive here, and a
    // diagnostic would be more noise than signal.
    LooseComparison(x: string, y: number) : boolean {
        return x is y
    }
}
