import json


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

    graph = load_graph_data()
    print(graph)



if __name__ == "__main__":
    main()