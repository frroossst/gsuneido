"""
A graph data structure with nodes and edges

The nodes are type variables while each edge represents equality or another constraint

A path finding algorithm is used to find a path from the start node to the end node, if 
a path is found, then the constraints are satisfiable, otherwise they are not
"""
class Graph:

    nodes = None

    def __repr__(self) -> str:
        return f"Graph({self.nodes})"

    def __init__(self):
        self.nodes = []

    def add_node(self, node):
        if self.find_node(node.value) is None:
            self.nodes.append(node)

    def find_node(self, name):
        for node in self.nodes:
            if node.value == name:
                return node
        return None

    def add_edge(self, node1, node2):
        n1 = self.find_node(node1)
        n2 = self.find_node(node2)

        if n1 is not None and n2 is not None:
            n1.add_edge(n2)
            n2.add_edge(n1)
        else:
            raise Exception("Node not found")

    def are_connected(self, node1, node2):
        n1 = self.find_node(node1)
        n2 = self.find_node(node2)

        if n1 is not None and n2 is not None:
            for edge in n1.get_connected_edges():
                if edge == n2:
                    return True
        return False

    # simple BFS to find a path from node1 to node2
    def path_exists(self, node1, node2):
        n1 = self.find_node(node1)
        n2 = self.find_node(node2)

        if n1 is not None and n2 is not None:
            visited = []
            queue = []

            queue.append(n1)

            while len(queue) > 0:
                node = queue.pop(0)
                visited.append(node)

                for edge in node.get_connected_edges():
                    if edge == n2:
                        return True
                    if edge not in visited:
                        queue.append(edge)
                        
        return False


class Node:

    value = None
    # neighbours, what it can see
    edges = None

    def __repr__(self) -> str:
        return f"Node({self.value}, {self.edges})"

    def __init__(self, name):
        self.value = name
        self.edges = []

    def get_connected_edges(self):
        return self.edges
    
    def add_edge(self, edge):
        self.edges.append(edge)


def test_test():

    graph = Graph()

    node_number = Node("Number")
    node_varx = Node("x")
    node_string = Node("String")
    node_vary = Node("y")


    graph.add_node(node_number)
    graph.add_node(node_varx)
    graph.add_node(node_string)
    graph.add_node(node_vary)

    graph.add_edge("Number", "x")
    graph.add_edge("x", "y")

    assert graph.are_connected("Number", "x") is True
    assert graph.path_exists("Number", "y") is True
    assert graph.are_connected("Number", "String") is False

    print("tests passed")

if __name__ == "__main__":
    test_test()