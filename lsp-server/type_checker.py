import json

from graph import Graph


def load_kv_data():
    with open("type_store.json", "r") as fobj:
        content = json.load(fobj)

    return content

def load_graph_data():
    with open("type_graph.json", "r") as fobj:
        content = fobj.read()

    return content

def main():

    content = load_kv_data()

    print(json.dumps(content, indent=4))
    print("=" * 80)

    graph = Graph().from_json(load_graph_data())

    print(graph)
    print("=" * 80)

    graph.add_edge("Number", "String")

    # check if a path exists between two primitive types
    primitive_types = Graph.get_primitive_type_nodes()
    for i in primitive_types:
        for j in primitive_types:
            if i.value == j.value:
                continue
            if graph.path_exists(i.value, j.value):
                raise TypeError(f"Types {i.value} and {j.value} cannot be equated")



if __name__ == "__main__":
    main()