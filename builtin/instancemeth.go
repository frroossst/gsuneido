package builtin

import . "github.com/apmckinlay/gsuneido/runtime"

func init() {
	InstanceMethods = Methods{
		"Members": method0(func(this Value) Value {
			return this.(*SuInstance).Members()
		}),
		"Member?": memberq,
		"Size": method0(func(this Value) Value {
			return this.(*SuInstance).Size()
		}),
		// TODO more methods
	}
}
