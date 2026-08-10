class {
    IsEven(x) {
        return (x % 2) is 0
    }
    IsEvenWrapper(y) {
        x = .IsEven()
        return .IsEven() 
    }
    A() {
        return .B()
    }
    B() {
        return .C()
    }
    C() {
        return .D()
    }
    D() {
        return 42
    }
}
