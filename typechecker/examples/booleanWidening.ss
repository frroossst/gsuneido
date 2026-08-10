// Boolean-widening rules for class-level state vs. method-local state.
//
// Class members are non-local: any method may toggle them. A literal `false`
// (or `true`) at the class level therefore widens to Boolean so callers
// reading the member do not get a tighter type than the writes can produce.
//
// Inline-init params (`.x = false`) implicitly create a class member from
// the default value, so the same widening must apply.
//
// Method locals (`x = false`) are the opposite: they keep the precise
// TFalse / TTrue so constraint checks stay tight inside the method.
class {
    // [class member, literal bool] -> widens to Boolean
    flagA: false

    // [class member, literal bool, also written elsewhere] -> Boolean
    flagB: false
    SetFlagB() { .flagB = true }

    // [non-bool member] -> stays at the concrete type (no spurious widening)
    counter: 0
    BumpCounter() { .counter += 1 }

    // [inline-init param creates a member] -> member must widen to Boolean
    // even when there is no separate class-level declaration for it.
    Configure(.allowOverride? = false) {
        // local read of the inline-init param sees the widened member type
        return .allowOverride?
    }

    // [method local] -> stays as TFalse (does NOT widen).
    // Locals are tight on purpose so constraint checks inside the method
    // stay precise.
    LocalFalse() {
        flag = false
        return flag
    }

    LocalTrue() {
        flag = true
        return flag
    }

    // [method local, mixed via union of returns] -> the function's return
    // type folds True | False to Boolean; the local x itself is whichever
    // branch wrote last, but the *return union* widens via Fold.
    LocalReturnedBool(c) {
        if c is 0 { return true }
        return false
    }
}
