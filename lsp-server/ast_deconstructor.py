import re



class AST:

    src = None

    def __init__(self, src) -> None:
        self.src = re.sub(r'\s+', ' ', src).strip()

    def parse(self):
        for i in range(len(self.src)):
            if self.src[i] == '(':
                pass
            elif self.src[i] == ')':
                pass
            else:
                pass


class BinaryNode:
    lhs = None
    rhs = None
    current = None
    node_t = None

    def __init__(self, current):
        self.current = current

    def set_lhs(self, lhs):
        self.lhs = lhs

    def set_rhs(self, rhs):
        self.rhs = rhs

    def get_type(self):
        return self.node_t
    
    def set_type(self, t):
        self.node_t = t

class UnaryNode:
    child = None
    current = None
    node_t = None
    
    def __init__(self, current):
        self.current = current
    
    def set_child(self, child):
        self.child = child
    
    def get_type(self):
        return self.node_t
    
    def set_type(self, t):
        self.node_t = t

class NaryNode:
    children = None
    current = None
    node_t = None
    
    def __init__(self, current):
        self.current = current
        self.children = []
    
    def append_child(self, child):
        self.children.append(child)
    
    def get_type(self):
        return self.node_t
    
    def set_type(self, t):
        self.node_t = t


# construct ast from the following string
# everything inside () is a node, Node is a class
ast_src = """Function(x
        Binary(Eq notLogic Unary(Not true))
        Binary(Eq constF Nary(Add 1 2 3 4 5))
        If(Nary(Or Unary(Not notLogic) notLogic) Return(Nary(Mul x 2)))
        Return(Nary(Add x 1 b)))"""

ast = AST.generate(ast_src)
