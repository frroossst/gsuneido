class {
    Foo(.x = false) {

    }

    Bar(y) {
        if .x is false {
            .Bar(0)
        }
    }
}
