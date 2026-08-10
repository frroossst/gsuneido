class {
    NoReturn() {
        return
    }
    OneLiteralReturn() {
        return false
    }
    OneKnownReturn() {
        x = 123
        return x
    }
    SimpleReturnUnion(x) {
        if x is false {
            return true
        } else {
            return false
        }
    }
    FindControlTypeReturn() {
        x = 1234
        if x % 2 is 0 {
            return x
        } else {
            return false
        }
    } 
    TwoBranchReturns(x) {
        if x is false {
            return "Fkdasjf"
        }
        return -1
    }
    UnknownReturn(x) {
        return x
    }
    UnionUnknownReturn(x) {
        if x is false {
            return 123
        }
        return x
    }
}