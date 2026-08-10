class {
    IfBranchesSame() {
        newdir = false
        if .x < 0
            newdir = -1
        else if .x is 0
            newdir = 0
        else if .x > 0
            newdir = 1

        x = newdir
        return x
    }

    IfBranchesDiffer() {
        newdir = false
        if .x < 0
            newdir = false
        else if .x is 0
            newdir = 'invalid'
        else if .x > 0
            newdir = 1

        x = newdir
        return x
    }

    IfBranchesDiffer2() {
        newdir = false
        if .x < 0
            newdir = false
        else if .x is 0
            newdir = 'invalid'
        else if .x > 0
            newdir = 1

        if Type(newdir) is "Number" {
            return newdir
        } else {
            return false
        }
    }

}
