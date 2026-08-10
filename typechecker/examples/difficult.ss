class {
    WhatIsSize(x) {
        // x can be anything that remotely implements the
        // could be String or Object or any random ass method 
        sz = x.Size()
    } 

    GoOnADate(x) {
        y = Date()
        // we dont know if x is a valid date or not
        // if it is not, we shouldnt be allowed to call this
        // method on it, but we literally dont know anuthing about x
        z = x.MinusDays(y)
        // we cannot infer what z is because we dont know if the MinusDays call is valid or not
    }
    
}