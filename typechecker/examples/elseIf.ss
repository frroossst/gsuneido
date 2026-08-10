class {
	record_change_members: false
	Foo() {
		if .isValid() {
			.record_change_members = Object()
		} else if Object?(.record_change_members) {
			.record_change_members.AddUnique(.record_change)
		} else {
			.record_change_members = Object(.members)
		}
	}
}
