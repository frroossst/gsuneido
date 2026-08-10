class {
	Foo(x) {
		if String?(x) and Number?(x) { return x }
	}
	Bar(x) {
		if String?(x) or Number?(x) { return x }
	}
}

