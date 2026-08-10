class {
    // SOUNDNESS GUARD: no diverging guard at all, so idx must stay False | ?
    // and the += must still error. Pins that narrowing is guard-conditional.
    NoGuard(name)
        {
        names = #('Line Items')
        idx = names.Find(name)
        while names.Member?(idx)
            idx += 0.1
        return idx
        }
    // SOUNDNESS GUARD: guard proves non-false at entry, but the body reassigns
    // idx to a possibly-false value. False must survive in the loop -> the +=
    // must STILL error. Pins "narrow the seed, never the head".
    ReassignFalseInLoop(name)
        {
        names = #('Line Items')
        if false is idx = names.Find(name)
            throw 'nope'
        while names.Member?(idx)
            {
            idx = names.Find(name)
            idx += 0.1
            }
        return idx
        }
    // CAPABILITY: return is a diverging guard too (branchAlwaysExits). Should
    // flip from error to no-error after the fix.
    ReturnGuard(name)
        {
        names = #('Line Items')
        if false is idx = names.Find(name)
            return false
        while names.Member?(idx)
            idx += 0.1
        return idx
        }
    // CAPABILITY: same shape but a for-loop carries idx through its body.
    ForLoopCarried(name)
        {
        names = #('Line Items')
        if false is idx = names.Find(name)
            throw 'nope'
        for (j = 0; j < 3; j++)
            idx += 0.1
        return idx
        }
    }
