// the purpose of this file is to test the in and in range operators
class {
    list: (1,2,3,4,5,6,7,8,9,0)
    InA(x) {
        lower = 0
        upper = 100
        mid = 50
        range = lower < mid < upper
    }
    InB(x) {
        x in (1, 2, 3)
    }
    SC(x) {
        v = x is 1 or x is 2 or x is 3
    }
}
