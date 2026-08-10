class {
    twice(x) {
        return [x, x]
    }
    Foo(ob) {
        ob.FlatMap(twice)
    }
    R() {
        #(1,2,3,4,5).Reduce({|x,y| x + y }) 
    }
    M() {
        Object(1, 2, 3).Map!({|x| x * 2 })
    }
    Blck() {
        for_each = function (ob, block)
        {
        for (x in ob)
            block(x)
        }
        for_each(#(1, 2, 3, 4), {|x| sum += x; })
        for_each(#(1, 2, 3, 4))
            {|x| sum += x; };
    }
}