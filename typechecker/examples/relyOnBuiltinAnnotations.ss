class {
    Tier1Hit(obj :object) {
        exists? = obj.Has?(#foo)
    }

    ReceiverMismatch(s :string) {
        n = s.Has?(#foo)
    }

    Tier3Single(obj) {
        exists? = obj.Has?(#foo)
    }

    Tier3None(obj) {
        result = obj.Frobnicate(#foo)
    }

    UnionMixed(x :string|object) {
        result = x.Has?(#foo)
    }

    UnionAllOk(x :record|object) {
        exists? = x.Has?(#foo)
    }

    ThisFallback() {
        // .members is an :object literal, but we deliberately use a
        // this-method call to exercise the this-branch of dispatch.
        return this.Has?(#anything)
    }

    DirtyUnion(flag :boolean) {
        x = flag ? #() : Mystery()
        result = x.Has?(#foo)
    }
}
