// ArgumentArity.ss
// Every Suneido argument-passing convention in one class, plus the broken-arity
// calls the type checker's ArityCheckPass now flags (in the "broken arity"
// section). The class compiles fine - arity is a runtime/static-check concern,
// not a compile error - so the broken methods just never get called at runtime.
class
	{
	// ----- parameter declaration styles -----

	sum(a, b)                          // required positional
		{ return a + b }

	box(w, h = 10, color = "white")    // trailing params with defaults
		{ return Object(:w, :h, :color) }

	collect(@args)                     // @param: gathers ALL args into one object
		{ return args }

	themed(_color = "grey")            // dynamic param: inherits caller's _color
		{ return color }               // (note: referenced without the underscore)

	applyBlock(block)                  // target for trailing-block sugar
		{ return block() }

	// ----- constructor member-init params -----

	New(.size, .Color = "white")       // .size -> private member, .Color -> public, with default
		{ }
	Describe()
		{ return Object(size: .size, color: .Color) }

	// ----- broken arity: ERRORS the checker now reports -----
	// (each comment is the exact diagnostic produced)

	tooMany()                          // too many arguments to sum: 3 given, takes at most 2
		{ return .sum(1, 2, 3) }

	missingOne()                       // missing argument to sum: b
		{ return .sum(1) }

	missingAll()                       // missing arguments to sum: a, b
		{ return .sum() }

	tooManyWithDefaults()              // too many arguments to box: 4 given, takes at most 3
		{ return .box(1, 2, 3, 4) }

	ctorTooMany()                      // too many arguments to new this: 3 given, takes at most 2
		{ return new this(1, 2, 3) }

	ctorMissing()                      // missing argument to new this: size
		{ return new this() }

	// ----- legal arity: deliberately NOT flagged (soundness) -----

	okDefaults()   { return .box(5) }                 // trailing defaults fill in
	okOmitMiddle() { return .box(5, color: "red") }   // omit h via a named arg
	okExtraNamed() { return .sum(1, 2, bogus: 9) }    // extra named arg is ignored
	okSpread(ob)   { return .sum(@ob) }               // @spread -> arity unknown, silent
	okVariadic()   { return .collect(1, 2, 3, x: 4) } // @args callee takes anything
	okDynamic()    { return .themed() }               // dynamic _color may be omitted

	// ----- legal but WARNED (confusing) -----
	// argument "h" to box is passed both positionally and by name;
	// the named value overrides the positional one (this may be confusing)
	doubleBound()  { return .box(1, 2, h: 99) }       // h given positionally (slot 1) AND by name

	}
