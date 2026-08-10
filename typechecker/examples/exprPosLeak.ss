// gsuneido wraps if/while/dowhile conditions, trinary cond/then/else,
// for-in expressions, and switch case exprs in *ast.ExprPos for position
// tracking. ExprPos embeds the ast.Expr interface and has no Children of
// its own, so visitors that call n.Children(fn) dispatch through the
// embedded interface to the wrapped expression's Children visiting
// its grandchildren and skipping the wrapped node entirely. The wrapped
// node never has its type set, and never appears in the AST dump.
class {
    IfCond(x) {
        if x is 5 {
            return true
        }
        return false
    }
    TrinaryLits(x) {
        return x is 0 ? "yes" : 42
    }
    WhileCond() {
        n = 0
        while n is 0 {
            n = 1
        }
        return n
    }
}
