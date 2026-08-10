// the purpose of this file is to show the types inferred from the class members are and can be 
// propogated downstream
class {
    num: 123
    msg: "hello, "
    isValid?: false
    Bar() {
        stdout = .msg
        isInferredValid = .isValid?
    }
}