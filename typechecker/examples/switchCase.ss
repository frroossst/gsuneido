class {
    Classify(x) {
        switch {
            case Number?(x):  return x
            case String?(x):  return x
            case Object?(x):  return x
            case Boolean?(x): return x
            case Date?(x):    return x
        }
    }

    ClassifyByTypeString(x) {
        switch {
            case Type(x) is "Number": return x
            case Type(x) is "String": return x
        }
    }

    NotPredicateOnUnion() {
        v = .Maybe()
        switch {
            case not Number?(v): return v   // v : False here
        }
        return v
    }
    Maybe() {
        if .pick is true { return 42 }
        return false
    }
    pick: false

    ByValue(x) {
        switch x {
            case 1:      return x   // x : Number
            case "hi":   return x   // x : String
            case false:  return x   // x : False
        }
    }

    MultiLiteralArm(x) {
        switch x {
            case 1, "hi": return x   // x : Number | String
        }
    }

    DefaultStripsCases(x) {
        switch x {
            case 1:    return "num"
            case "hi": return "str"
            default:   return x
        }
    }

    CondModeDefault(x) {
        switch {
            case String?(x): return false
            default:         return x + 1
        }
    }

    MergeAssignments(x) {
        v = false
        switch x {
            case 1: v = "one"
            case 2: v = 2
        }
        return v
    }

    MergeAssignmentsWithDefault(x) {
        v = false
        switch x {
            case 1: v = "one"
            case 2: v = 2
            default: v = "fallback"
        }
        return v
    }

    PostSwitchNarrow(x) {
        v = false
        switch x {
            case 1: v = "one"
            case 2: v = 2
        }
        if Type(v) is "Number"
            return v        // v : Number here
        return false
    }

    foo: 0
    MemberNotNarrowed() {
        switch {
            case Number?(.foo): return .foo
        }
    }

    AssignmentInArmInvalidates(x) {
        switch {
            case Number?(x): x = "now a string"; return x   // returns String
        }
    }

    GlobalScrutinee() {
        switch X {
            case 5: return X
        }
    }

    NestedSwitch(x, y) {
        v = false
        switch x {
            case 1:
                switch y {
                    case 10: v = "ten"
                    case 20: v = 20
                }
            case 2:
                v = #other
        }
        return v
    }

    EmptyKeepsEntry(x) {
        v = 42
        switch x {
        }
        return v       // Number
    }
}
