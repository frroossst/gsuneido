import json

from graph import Graph
from kvstore import KVStore
from sutypes import check_type_equality
from type_inference import propogate_infer



def load_kv_data():
    with open("type_store.json", "r") as fobj:
        content = json.load(fobj)

    return content

def load_graph_data():
    with open("type_graph.json", "r") as fobj:
        content = fobj.read()

    return content

def main():

    store = KVStore().from_json(load_kv_data())
    ascii_store = """
     ____ _____ ___  ____  _____ 
    / ___|_   _/ _ \|  _ \| ____|
    \___ \ | || | | | |_) |  _|  
     ___) || || |_| |  _ <| |___ 
    |____/ |_| \___/|_| \_\_____|

    """
    print(ascii_store)
    print("=" * 80)
    print(json.dumps(store.to_json(), indent=4))


    graph = Graph().from_json(load_graph_data())
    ascii_graph = """
      ____ ____      _    ____  _   _ 
     / ___|  _ \    / \  |  _ \| | | |
    | |  _| |_) |  / _ \ | |_) | |_| |
    | |_| |  _ <  / ___ \|  __/|  _  |
     \____|_| \_\/_/   \_\_|   |_| |_|

    """
    print(ascii_graph)
    print("=" * 80)
    print(json.dumps(graph.to_json(), indent=4))

    for k, v in store.db.items():
        print(f"[DEBUG] Type: {k}, Value: {v}")
        if not check_type_equality(v.actual, v.inferred):
            str_fmt = f"[ERROR] type node {k} expected type {v.inferred} but got {v.actual} instead. Value = {store.get(k)}"
            # raise TypeError(str_fmt)
            print(str_fmt)

    propogate_infer(store, graph, check=True)

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