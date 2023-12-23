import ast
import astor

src= """
def stringQ(x):
    return True

def numberQ(x):
    return True

def qux():
    pass

def function(x, y, z):
    num = x + "123"
    num += 1
		
    if stringQ(x) and numberQ(y):
        abc = x + y + z + num
    else:
        num()
    qux()
"""

ast = ast.parse(src)
print(astor.dump_tree(ast, indentation="  "))
