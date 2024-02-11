# Proof Of Concept
# Takes in a json object and dumps type information to a file
# Checks basic type safety
import argparse
import json

from graph import Graph, Node
from kvstore import KVStore, StoreValue
from sutypes import SuTypes
from utils import DebugInfo

def check_type_equivalence(lhs, rhs) -> bool:
    if lhs == SuTypes.Any or rhs == SuTypes.Any:
        return True
    if lhs == SuTypes.Unknown or rhs == SuTypes.Unknown:
        raise TypeError("Unknown type in type equivalence check")
    return lhs == rhs

def load_data_body() -> dict:

    with open('ast.json') as data_file:
        data = json.load(data_file)

    return data['Methods']

def load_data_attributes() -> dict:

    with open('ast.json') as data_file:
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
    if valid_types is None:
        raise NotImplementedError("constraint and operator not implemented")
    return type in valid_types

def get_valid_type_for_operator(value) -> SuTypes:
    valid_types = {
        "Add": SuTypes.Number,
        "PostInc": SuTypes.Number,
        "And": SuTypes.Boolean,
        "Cat": SuTypes.String,
    }

    if (x:= valid_types.get(value, None) )is None:
        raise NotImplementedError("valid operator type not implemented")
    return x

def get_type_assertion_functions() -> list[str]:
    return [
        "String?",
        "Number?",
        "Boolean?",
        "Object?",
        "Class?",
        "Function?",
        "Date?",
    ]


def infer_generic(stmt, store, graph, attributes) -> SuTypes:
    print(f"[LOG] {store.get('e13f80ba58bf46138a09eef589eb0c76')}")
    match stmt["Tag"]:
        case "Unary":
            return infer_unary(stmt, store, graph, attributes)
        case "Binary":
            return infer_binary(stmt, store, graph, attributes)
        case "Nary":
            return infer_nary(stmt, store, graph, attributes)
        case "Identifier":
            if store.get(stmt["ID"]) is not None:
                return store.get(stmt["ID"]).inferred
            return SuTypes.Any
        case "If":
            return infer_if(stmt, store, graph, attributes)
        case "Call" | "Compound":
            return infer_generic(stmt["Args"][0], store, graph, attributes)
        case "Return":
            return infer_generic(stmt["Args"][0], store, graph, attributes)
        case "Object":
            return infer_object(stmt["Args"], store, graph, attributes)
        case "Member":
            return infer_attribute(stmt, store, graph, attributes)
        case "Constant":
            return SuTypes.from_str(stmt["Type_t"])
        case _:
            raise NotImplementedError(f"missed case {stmt['Tag']}")

def infer_unary(stmt, store, graph, attributes) -> SuTypes:
    args = stmt["Args"]
    value = stmt["Value"]

    node = Node(args[0]["ID"])
    graph.add_node(node)

    valid_t = get_valid_type_for_operator(value)
    vn = Node(valid_t.name)
    graph.add_node(vn)
    graph.add_edge(node.value, vn.value)

    return valid_t


def infer_binary(stmt, store, graph, attributes) -> SuTypes:
    args = stmt["Args"]
    lhs = args[0]
    rhs = args[1::][0]

    # inferred types of lhs and rhs should be the same
    lhs_t = infer_generic(lhs, store, graph, attributes)
    if lhs_t is not None and lhs["Type_t"] != "Operator":
        v = StoreValue(lhs["Value"], lhs["Type_t"], lhs_t)
        store.set(lhs["ID"], v)
    rhs_t = infer_generic(rhs, store, graph, attributes)
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

    # v = StoreValue(stmt["Value"], stmt["Type_t"], lhs_t)
    # store.set(lhs["ID"], v)
    # v = StoreValue(stmt["Value"], stmt["Type_t"], rhs_t)
    # store.set(rhs["ID"], v)

    # if lhs or rhs is of type Any, then the other type is inferred
    if lhs_t == SuTypes.Any:
        store.set(lhs["ID"], StoreValue(rhs["Value"], rhs["Type_t"], rhs_t))
        return rhs_t
    elif rhs_t == SuTypes.Any:
        store.set(rhs["ID"], StoreValue(lhs["Value"], lhs["Type_t"], lhs_t))
        return lhs_t
    else:
        return SuTypes.NotApplicable
    

def infer_nary(stmt, store, graph, attributes) -> SuTypes:
    value = stmt["Value"]
    args = stmt["Args"]

    valid_t = get_valid_type_for_operator(value)

    prev = None
    for i in args:
        if i["Tag"] == "Call":
            i = i["Args"][0]
            if i["Value"] in get_type_assertion_functions():
                typed_check_t = SuTypes.from_str(i["Value"].removesuffix("?"))

                type_checked_var = i["Args"][0]

                n = Node(type_checked_var["ID"])
                v = StoreValue(type_checked_var["Value"], SuTypes.from_str(type_checked_var["Type_t"]), typed_check_t)
                store.set(type_checked_var["ID"], v)
                graph.add_node(n)

                primitive_type_node = graph.find_node(SuTypes.to_str(typed_check_t))
                graph.add_edge(n.value, primitive_type_node.value)

        """
        NOTE: Infer Generic cause Args might not always be constants and variables,
                so infer a generic somewhere here to infer further
        """

        n = Node(i["ID"])
        v = StoreValue(i["Value"], SuTypes.from_str(i["Type_t"]), valid_t)
        store.set(i["ID"], v)
        graph.add_node(n)
        if prev is not None:
            graph.add_edge(prev.value, n.value)
        prev = n

    valid_str = SuTypes.to_str(valid_t)
    n = graph.find_node(valid_str)
    # n.add_edge(prev)
    graph.add_edge(prev.value, n.value)

    return valid_t

def infer_if(stmt, store, graph, attributes):
    cond = stmt["Args"][0]
    cond_t = infer_generic(cond, store, graph, attributes)
    v = StoreValue(cond["Value"], SuTypes.from_str(cond["Type_t"]), cond_t)
    store.set(cond["ID"], v)

    then = stmt["Args"][1]
    then_t = infer_generic(then, store, graph, attributes)
    v = StoreValue(then["Value"], SuTypes.from_str(then["Type_t"]), then_t)
    store.set(then["ID"], v)

    if len(stmt["Args"]) == 3:
        else_t = infer_generic(stmt["Args"][2], store, graph, attributes)
        v = StoreValue(stmt["Args"][2]["Value"], SuTypes.from_str(stmt["Args"][2]["Type_t"]), else_t)
        store.set(stmt["Args"][2]["ID"], v)

    return SuTypes.NotApplicable

def infer_attribute(stmt, store, graph, attributes):
    value = stmt["Value"]
    attrb_t = attributes.get(value, None)

    if attrb_t is None:
        raise TypeError(f"Attribute {value} not found")
    
    valid_t = SuTypes.from_str(attrb_t["Type_t"])
    v = StoreValue(stmt["Value"], SuTypes.from_str(stmt["Type_t"]), valid_t)
    store.set(stmt["ID"], v)

    n = Node(stmt["ID"])
    graph.add_node(n)
    t = graph.find_node(valid_t.name)
    # t.add_edge(n)
    graph.add_edge(n.value, t.value)

    return valid_t

def infer_object(stmt, store, graph, attributes):

    for i in stmt:
        t = infer_generic(i["Args"][0], store, graph, attributes)
        v = StoreValue(i["Args"][0]["Value"], SuTypes.from_str(i["Args"][0]["Type_t"]), t)
        store.set(i["Args"][0]["ID"], v)
        n = Node(i["Args"][0]["ID"])
        graph.add_node(n)
        graph.add_edge(n.value, graph.find_node(t.name).value)

    return SuTypes.Object

def propogate_infer(store, graph, check=False):
    primitives = graph.get_primitive_type_nodes()

    # dfs through each primitive and assign the same sutype to connecting nodes
    for p in primitives:
        p = graph.find_node(p.value)
        p.propogate_type(store, new_type=p.sutype, check=check)

def parse_class(clss):
    members = {}

    for k, v in clss.items():
        members[k] = v[0]

    return members

def process_methods(methods, store, graph, attributes):
    for k, v in methods.items():
        debug_info.set_func(k)
        if debug_info.func_name == "SameVarID":
            pass
        print(f"{k}: {json.dumps(v, indent=4)}")
        for x, i in enumerate(v["Body"]):
            debug_info.set_line(x + 1)
            valid_t = infer_generic(i[0], store, graph, attributes)
            if valid_t == SuTypes.NotApplicable or valid_t is None:
                continue
            v = StoreValue(i[0]["Value"], SuTypes.from_str(i[0]["Type_t"]), valid_t)
            store.set(i[0]["ID"], v)
            n = Node(i[0]["ID"])
            graph.add_node(n)
            graph.add_edge(n.value, graph.find_node(valid_t.name).value)

def main():
    global debug_info
    debug_info = DebugInfo()

    p = argparse.ArgumentParser("Type Inference")
    p.add_argument("-t", action="store_true")
    args = p.parse_args()

    graph = Graph()
    store = KVStore()
    attributes = parse_class(load_data_attributes())
    methods = parse_class(load_data_body())

    print("=" * 80)
    ascii_blocks = """
     ____  _     ___   ____ _  ______  
    | __ )| |   / _ \ / ___| |/ / ___| 
    |  _ \| |  | | | | |   | ' /\___ \ 
    | |_) | |__| |_| | |___| . \ ___) |
    |____/|_____\___/ \____|_|\_\____/ 
    """
    print(ascii_blocks)
    print("=" * 80)

    try:
        process_methods(methods, store, graph, attributes)
        propogate_infer(store, graph, attributes)
    except Exception as e:
        if not not args.t:
            print(f"Exception: {e}")
        else:
            debug_info.trigger(e)

    # graph.visualise()
    
    print("=" * 80)
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
    print("=" * 80)
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


    with open("type_store.json", "w") as fobj:
        json.dump(store.to_json(), fobj, indent=4)

    with open("type_graph.json", "w") as fobj:
        json.dump(graph.to_json(), fobj, indent=4)

if __name__ == "__main__":
    main()
