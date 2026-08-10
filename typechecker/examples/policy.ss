class {
    StrictStringConcat() {
        x = 123
        return x $ " <- "
    }
    CrossTypeCompare() {
        x = "string"
        y = false
        return x < y
    }
}

