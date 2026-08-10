// Flow-sensitive occurrence typing.
//
// Inside an if/else/?: arm guarded by a Suneido type predicate, the local
// being tested gets narrowed to the matching type. Same for `is <literal>`
// and `Type(x) is "..."`. This is the language's idiom: every "checked"
// access is conventionally guarded with one of these forms, so flow
// narrowing turns `TUnknown` into something useful without any annotation.
class {
    bigList: false
    NewList() {
        list = Object() 
        list.Add(1,2,3,4).Add(4,5,6,7)
        .bigList = list
    }
    FindInList(target) {
        for item in .bigList {
            if item is target {
                return Object()
            }
        }
        return false
    }

    // Predicate set per suneidoc/Language/Reference/*Q.md
    StringCase(x) {
        if String?(x) {
            // x is narrowed to String here
            return x
        }
        // No narrowing in the fall-through; x stays Unknown
    }

    NumberCase(x) {
        if Number?(x) { return x }
    }

    DateCase(x) {
        if Date?(x) { return x }
    }

    BoolCase(x) {
        if Boolean?(x) { return x }
    }

    ObjectOrRecordCase(x) {
        // Object? is documented to also return true for records.
        if Object?(x) { return x }
    }

    RecordCase(x) {
        if Record?(x) { return x }
    }

    ClassCase(x) {
        if Class?(x) { return x }
    }

    FunctionCase(x) {
        // Function? covers functions, methods, and blocks.
        if Function?(x) { return x }
    }

    // Negation flips polarity: in the body, x is "not String".
    // For a plain Unknown there is nothing concrete to remove, so this
    // mostly matters for unions.
    NotPredicate(x) {
        if not String?(x) {
            return x
        }
    }

    // `is <literal>` narrows to the literal's type.
    IsLiteral(x) {
        if x is 5 { return x }       // x : Number here
    }

    // `isnt false` removes TFalse from a union; common for nullable result.
    IsntFalse() {
        x = .Maybe()
        if x isnt false {
            return x
        }
        return false
    }

    // Type(x) is "Foo" mirrors the predicate set.
    TypeStringEq(x) {
        if Type(x) is "Number" { return x }
        if Type(x) is "String" { return x }
    }

    // and-chain composes both refinements in the body.
    AndChain(a, b) {
        if Number?(a) and String?(b) {
            return a + 1   // a : Number
        }
    }

    // or-chain narrows in the FALSE branch (compose negations).
    OrChain(x) {
        if Number?(x) or String?(x) {
            return x
        }
        // here, x is neither Number nor String
        return x
    }

    // Else branch sees the negated refinement.
    ElseBranch(x) {
        if Number?(x) {
            return x
        } else {
            return x   // x is "not Number" here (effective only on unions)
        }
    }

    // Ternary arms are guarded the same way as if/else.
    Ternary(x) {
        return Number?(x) ? x : false
    }

    // After the if, the narrowing does NOT escape (we don't merge yet).
    NoEscape(x) {
        if Number?(x) {
            // x : Number here
        }
        return x   // x is still whatever it was before the if
    }

    // Helper that returns Number | False so we have something to narrow.
    Maybe() {
        if .pick is true { return 42 }
        return false
    }

    pick: false
}
