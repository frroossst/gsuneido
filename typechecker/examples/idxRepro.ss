class {
    InLoop(name)
        {
        names = #('Line Items')
        if false is idx = names.Find(name)
            throw 'nope'
        while names.Member?(idx)
            idx += 0.1
        return idx
        }
    StraightLine(name)
        {
        names = #('Line Items')
        if false is idx = names.Find(name)
            throw 'nope'
        idx += 0.1
        return idx
        }
}
