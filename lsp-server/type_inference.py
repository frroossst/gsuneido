# Proof Of Concept
# Takes in a json object and dumps type information to a file
# Checks basic type safety

from graph import Graph, Node
from kvstore import KVStore, StoreValue
from sutypes import SuTypes
import json

from utils import catch_exception, todo

def check_type_equivalence(lhs, rhs) -> bool:
    if lhs == SuTypes.Any or rhs == SuTypes.Any:
        return True
    if lhs == SuTypes.Unknown or rhs == SuTypes.Unknown:
        raise TypeError("Unknown type in type equivalence check")
    return lhs == rhs

    
"""
! It assumes a singular class
! It does not handle scope level
"""
class Identifier:
    @classmethod
    def __init__(self, function_name: str, variable_name: str, scope_level: int = 1) -> str:
        return f"{function_name}::{'@' * scope_level}{variable_name}"

def load_data_body() -> dict:

    with open('output.json') as data_file:
        data = json.load(data_file)

    return data['Methods']

def load_data_attributes() -> dict:

    with open('output.json') as data_file:
        data = json.load(data_file)

    return data['Attributes']


def constraint_type_with_operator_value(value, type) -> bool:
    valid_constraints = {
        "Add": ["Number"],
    }

    if type == "Operator":
        return True

    if type == "Variable":
        return True

    valid_types = valid_constraints.get(value, None)
    return type in valid_types

def get_valid_type_for_operator(value) -> SuTypes:
    valid_types = {
        "Add": SuTypes.Number,
        "PostInc": SuTypes.Number,
        "And": SuTypes.Boolean,
    }

    return valid_types[value]


def infer_generic(stmt, store, graph) -> SuTypes:
    match stmt["Tag"]:
        case "Unary":
            return infer_unary(stmt, store, graph)
        case "Binary":
            return infer_binary(stmt, store, graph)
        case "Nary":
            return infer_nary(stmt, store, graph)
        case "Identifier":
            return SuTypes.Any
        case "If":
            return infer_if(stmt, store, graph)
        case "Call" | "Compound":
            return infer_generic(stmt["Args"][0], store, graph)
        case "Return":
            return infer_generic(stmt["Args"][0], store, graph)
        case "Object":
            return infer_object(stmt["Args"], store, graph)
        case "Member":
            return infer_attribute(stmt, store, graph)
        case "Constant":
            return SuTypes.from_str(stmt["Type_t"])
        case _:
            raise NotImplementedError(f"missed case {stmt['Tag']}")

def infer_unary(stmt, store, graph) -> SuTypes:
    args = stmt["Args"]
    value = stmt["Value"]

    node = Node(args[0]["ID"])
    graph.add_node(node)

    valid_t = get_valid_type_for_operator(value)
    vn = Node(valid_t.name)
    graph.add_node(vn)
    graph.add_edge(node.value, vn.value)

    return valid_t


def infer_binary(stmt, store, graph) -> SuTypes:
    args = stmt["Args"]
    lhs = args[0]
    rhs = args[1::][0]

    # inferred types of lhs and rhs should be the same
    lhs_t = infer_generic(lhs, store, graph)
    if lhs_t is not None and lhs["Type_t"] != "Operator":
        v = StoreValue(lhs["Value"], lhs["Type_t"], lhs_t)
        store.set(lhs["ID"], v)
    rhs_t = infer_generic(rhs, store, graph)
    if rhs_t is not None and rhs["Type_t"] != "Operator":
        v = StoreValue(rhs["Value"], rhs["Type_t"], rhs_t)
        store.set(rhs["ID"], v)

    # if not check_type_equivalence(lhs_t, rhs_t):
    #     raise TypeError(f"Type mismatch for {lhs_t} and {rhs_t}")
    lhs_n = Node(lhs["ID"])
    rhs_n = Node(rhs["ID"])
    graph.add_node(lhs_n)
    graph.add_node(rhs_n)
    graph.add_edge(lhs_n.value, rhs_n.value)

    v = StoreValue(stmt["Value"], stmt["Type_t"], lhs_t)
    store.set(lhs["ID"], v)
    v = StoreValue(stmt["Value"], stmt["Type_t"], rhs_t)
    store.set(rhs["ID"], v)

    # if lhs or rhs is of type Any, then the other type is inferred
    if lhs_t == SuTypes.Any:
        store.set(lhs["ID"], StoreValue(rhs["Value"], rhs["Type_t"], rhs_t))
        return rhs_t
    elif rhs_t == SuTypes.Any:
        store.set(rhs["ID"], StoreValue(lhs["Value"], lhs["Type_t"], lhs_t))
        return lhs_t
    

def infer_nary(stmt, store, graph) -> SuTypes:
    value = stmt["Value"]
    args = stmt["Args"]

    valid_t = get_valid_type_for_operator(value)

    prev = None
    for i in args:
        if i["Tag"] == "Call":
            i = i["Args"][0]
        n = Node(i["ID"])
        v = StoreValue(i["Value"], SuTypes.from_str(i["Type_t"]), valid_t)
        store.set(i["ID"], v)
        graph.add_node(n)
        if prev is not None:
            graph.add_edge(prev.value, n.value)
        prev = n

    valid_str = SuTypes.to_str(valid_t)
    n = graph.find_node(valid_str)
    n.add_edge(prev)

    return valid_t

def infer_if(stmt, store, graph):
    cond = stmt["Args"][0]
    cond_t = infer_generic(cond, store, graph)
    v = StoreValue(cond["Value"], SuTypes.from_str(cond["Type_t"]), cond_t)
    store.set(cond["ID"], v)

    then = stmt["Args"][1]
    then_t = infer_generic(then, store, graph)
    v = StoreValue(then["Value"], SuTypes.from_str(then["Type_t"]), then_t)
    store.set(then["ID"], v)

    if len(stmt["Args"]) == 3:
        else_t = infer_generic(stmt["Args"][2], store, graph)
        v = StoreValue(stmt["Args"][2]["Value"], SuTypes.from_str(stmt["Args"][2]["Type_t"]), else_t)
        store.set(stmt["Args"][2]["ID"], v)

    return SuTypes.NotApplicable

def infer_attribute(stmt, store, graph):
    value = stmt["Value"]
    attrb_t = attributes.get(value, None)

    if attrb_t is None:
        raise TypeError(f"Attribute {value} not found")
    
    valid_t = SuTypes.from_str(attrb_t["Type_t"])
    v = StoreValue(stmt["Value"], SuTypes.from_str(stmt["Type_t"]), valid_t)
    store.set(stmt["ID"], v)

    n = Node(stmt["ID"])
    graph.add_node(n)
    graph.find_node(valid_t.name).add_edge(n)

    return valid_t

def infer_object(stmt, store, graph):

    for i in stmt:
        t = infer_generic(i["Args"][0], store, graph)
        v = StoreValue(i["Args"][0]["Value"], SuTypes.from_str(i["Args"][0]["Type_t"]), t)
        store.set(i["Args"][0]["ID"], v)
        n = Node(i["Args"][0]["ID"])
        graph.add_node(n)
        graph.add_edge(n.value, graph.find_node(t.name).value)

    return SuTypes.Object


def parse_class(clss):
    members = {}

    for k, v in clss.items():
        members[k] = v[0]

    return members

@catch_exception
def process_methods(methods, store, graph):
    for k, v in methods.items():
        # NOTE: only for debugging
        global current_function
        current_function = k
        print(f"{k}: {json.dumps(v, indent=4)}")
        for i in v["Body"]:
            valid_t = infer_generic(i[0], store, graph)
            if valid_t == SuTypes.NotApplicable:
                continue
            v = StoreValue(i[0]["Value"], SuTypes.from_str(i[0]["Type_t"]), valid_t)
            store.set(i[0]["ID"], v)
            n = Node(i[0]["ID"])
            graph.add_node(n)
            graph.add_edge(n.value, graph.find_node(valid_t.name).value)

def main():
    graph = Graph()
    store = KVStore()
    global attributes
    attributes = parse_class(load_data_attributes())
    methods = parse_class(load_data_body())

    process_methods(methods, store, graph)

    graph.visualise()
    
    print("=" * 80)
    print(json.dumps(store.to_json(), indent=4))
    print("=" * 80)
    print(json.dumps(graph.to_json(), indent=4))


    with open("type_store.json", "w") as fobj:
        json.dump(store.to_json(), fobj, indent=4)

    with open("type_graph.json", "w") as fobj:
        json.dump(graph.to_json(), fobj, indent=4)

if __name__ == "__main__":
    main()
