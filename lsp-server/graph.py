from matplotlib import pyplot as plt
import networkx as nx
import json

from sutypes import SuTypes

"""
A graph data structure with nodes and edges

The nodes are type variables while each edge represents equality or another constraint

A path finding algorithm is used to find a path from the start node to the end node, if 
a path is found, then the constraints are satisfiable, otherwise they are not
"""
class Graph:

    nodes = None

    def __repr__(self) -> str:
        return f"Graph(\n\t{self.nodes}\n)"

    def __init__(self):
        self.nodes = []

        # add primitive types
        for i in self.get_primitive_type_nodes():
            self.add_node(i)

    @classmethod
    def get_primitive_type_nodes(cls):
        return [
            Node("String", SuTypes.String),
            Node("Number", SuTypes.Number),
            Node("Boolean", SuTypes.Boolean),
            Node("Object", SuTypes.Object),
            Node("Function", SuTypes.Function)
        ]

    def find_node(self, name):
        for node in self.nodes:
            if node.value == name:
                return node
        return None

    def add_node(self, node):
        if self.find_node(node.value) is None:
            self.nodes.append(node)

    def add_edge(self, node1, node2):
        n1 = self.find_node(node1)
        n2 = self.find_node(node2)

        if n1 is not None and n1 is not None:
            n1.add_edge(n1)
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

    """
    simple BFS to find a path from node1 to node2
    """
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

    
    def visualise(self):
        G = nx.Graph()

        for node in self.nodes:
            G.add_node(node.value)
            for edge in node.get_connected_edges():
                G.add_edge(node.value, edge.value)

        pos = nx.spring_layout(G)
        labels = {node.value: node.value for node in self.nodes}

        nx.draw(G, pos, with_labels=True, labels=labels)
        plt.show()

    def to_json(self):
        graph_data = {
            'nodes': [
                {'value': node.value, 'sutype': node.sutype.name, 'edges': [edge.value for edge in node.edges]}
                for node in self.nodes
            ]
        }
        return json.dumps(graph_data, indent=2)

    @classmethod
    def from_json(cls, json_data):
        graph_instance = cls()
        graph_data = json.loads(json_data)

        for node_data in graph_data.get('nodes', []):
            value = node_data.get('value')
            sutype = SuTypes[node_data.get('sutype', 'Unknown')]
            node = Node(value, sutype)
            graph_instance.add_node(node)

            edges = node_data.get('edges', [])
            for edge_value in edges:
                edge_node = graph_instance.find_node(edge_value)
                if edge_node:
                    graph_instance.add_edge(value, edge_value)

        return graph_instance


class Node:

    # this is the value of the type i.e. "hello", 12.43
    value = None

    # this is the type of the value i.e. String, Number
    sutype = None

    # neighbours, what it can see
    edges = None

    def __repr__(self) -> str:
        return f"Node(type = {self.sutype}, value = {self.value}, edges = {self.edges})"

    def __init__(self, value, sutype = SuTypes.Unknown):
        self.value = value
        self.edges = []
        self.sutype = sutype

    def get_connected_edges(self):
        return self.edges
    
    # type edge = Node
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

