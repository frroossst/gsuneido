class {
    Sum(n = 0) {
        if n <= 0 { return 0 }
        return n + .Sum(n - 1)
    }

    Fact(x) {
        if x is 0 { return 1 }
        if x is 1 { return 1 }
        return Fact(x * (x - 1))
    }
}