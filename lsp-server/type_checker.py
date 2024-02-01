import json


def load_graph_data():
    with open("type_store.json", "r") as fobj:
        content = json.load(fobj)

    return content

def main():

    content = load_graph_data()

    print(json.dumps(content, indent=4))


if __name__ == "__main__":
    main()