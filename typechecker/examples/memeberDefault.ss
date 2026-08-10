class {
	Foo(x :string, y :number|false, z :object = false) :object|false
		{
		if Type(y) is "Number"
			return false
		return Object()
		}
}
