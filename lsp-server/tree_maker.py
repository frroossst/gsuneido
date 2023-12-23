import json

class TreeFactory:

    @classmethod
    def load(self, filename) -> dict:
        with open(filename) as fobj:
            data = json.load(fobj)

        return data['Body']

"""
A node has the following properties:
    - Tag: This is what could be referred to as types of the node (Function, Binary, etc.)
    - Value: What the node actually stores (123, "hello", etc.)
    - Type_t: The type of the node (int, str, etc.)
    - Children: A list of nodes that are children of the node

N.B. It might seem a little ambiguous the difference between Tag and Type_t. 
"""
class Node:

    def __init__(self, tag, value, type_t, children):
        self.tag = tag
        self.value = value
        self.type_t = type_t
        self.children = children



def main():
    data = TreeFactory.load("output.json")
    print(json.dumps(data, indent=4))



if __name__ == "__main__":
    main()
