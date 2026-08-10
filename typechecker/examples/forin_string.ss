class
    {
    LiteralString()
        {
        for i in "adfdasf"
            Print(Type(i))
        }

    StringCall()
        {
        for i in String(789342)
            Print(Type(i))
        }

    DisplayCall()
        {
        for i in Display(123, 0)
            Print(Type(i))
        }

    MethodReturnsString()
        {
        for i in .greeting()
            Print(Type(i))
        }

    greeting()
        {
        return "hello"
        }

    LocalAssigned()
        {
        s = "abc"
        for i in s
            Print(Type(i))
        }

    UnknownIterable(ob)
        {
        // ob is Unknown; loop var must stay Unknown
        for i in ob
            Print(i)
        }

    NotIteratingString()
        {
        // negative case: iterable is Object, var stays Unknown
        for i in #(1, 2, 3)
            Print(i)
        }
    }
