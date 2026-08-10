class {
    accounts: #(1,2,3,4,5)
    Make(name, age, account) {
        .name = name
        if age < 18 {
            return false
        }
        if .accounts.Has?(account) {
        	return "account already exists!!!"
        }
		return Object(:name, :age, :account)
    }
}