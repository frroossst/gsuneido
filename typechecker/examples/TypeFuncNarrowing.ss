// Gains from narrowing on `Type(x) is "Foo"` (and `isnt`).
//
// The engine already supports this form (flow_narrowing.go: refineTypeStringEq).
// This file is a forcing-function demo: every method below has a parameter
// that starts as TUnknown, then is narrowed by a `Type(x) is "..."` (or its
// negation) so subsequent reads of x carry a concrete type. //
// Why this matters: the predicate idiom `String?(x)` is one stylistic choice;
// the explicit `Type(x) is "String"` is the other, and Suneido stdlib uses
// both. Without this rule, half the codebase's type guards are invisible to
// the inference engine, so parameters and locals stay TUnknown after the
// guard and every downstream operation that depends on the type goes
// unresolved.
class
	{
	// One narrowed branch per Type-string. Inside the if, x has the concrete
	// primitive type and Report() observes its inferred return.
	StringBranch(x)		{
		if Type(x) is "String"
			return x  // x is narrowed to TString
		return ""
		}

	NumberBranch(x)		{
		if Type(x) is "Number"
			return x  // TNumber
		return 0
		}

	DateBranch(x)		{
		if Type(x) is "Date"
			return x  // TDate
		return Date()
		}

	BooleanBranch(x)		{
		if Type(x) is "Boolean"
			return x  // TBoolean
		return false
		}

	ObjectBranch(x)		{
		if Type(x) is "Object"
			return x  // TObject
		return Object()
		}

	RecordBranch(x)		{
		if Type(x) is "Record"
			return x  // TRecord
		return [empty: true]
		}

	// Negation flips polarity: inside the body x is "not String".
	// On a bare TUnknown there is nothing concrete to remove, so the only
	// observable effect is on unions - see UnionDrop below.
	NotString(x)
		{
		if Type(x) isnt "String"
			return x
		return x  // also TUnknown here; demo of "no spurious narrowing"
		}

	// Reverse-order operands work too: "Foo" is Type(x) narrows the same way.
	ReversedOperand(x)		{
		if "Number" is Type(x)
			return x  // TNumber
		return 0
		}

	// Else branch sees the negated guard - meaningful when x carries a union.
	// Helper Maybe() returns Number | False, so the else-branch drops Number,
	// leaving x as TFalse.
	UnionElseBranch()		{
		x = .Maybe()  // Number | False
		if Type(x) is "Number"
			return x + 1     // x : Number here
		else
			return x         // TFalse here - the whole "wasn't Number" branch
		}

	// And-chain composes both refinements. After both guards are true, a is
	// String and b is Number, so a $ Display(b) is well-typed.
	AndChain(a, b)		{
		if Type(a) is "String" and Type(b) is "Number"
			return a $ Display(b)  // explicit Display - no coercion reliance
		return ""
		}

	// Ternary arms narrow per arm just like if/else.
	Ternary(x)		{
		return Type(x) is "Number" ? x + 0 : 0
		}

	// Realistic shape: a callsite returns "Number | False" and we want the
	// Number out. Concrete code path is what HttpClient.socketClient does
	// after parsing a port, just rewritten to use the comparison form
	// rather than Assert's named-arg form.
	PortFromHeaders(headers)		{
		port = headers.GetDefault('port', 80)  // unknown statically
		if Type(port) is "Number"
			return port + 0   // narrowed to TNumber, arithmetic is well-typed
		return 80
		}

	// Maybe-style helper for UnionElseBranch: returns Number | False.
	Maybe()		{
		if .pick is true
			return 42
		return false
		}

	pick: false


    foo: class {} 
    DoesDetectInstance() {
        if Type(.foo) is "string" {
            return ""
        }
        if Type(.foo) is "Instance" {
            return ""
        }
    }
	}
