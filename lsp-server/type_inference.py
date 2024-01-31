# Proof Of Concept
# Takes in a json object and dumps type information to a file
# Checks basic type safety

from graph import Graph, Node
from enum import Enum
import json

class SuTypes(Enum):
    Unknown = 0
    String = 1
    Number = 2
    Boolean = 3
    Any = 4
    NotApplicable = 5
    Never = 6
    Function = 7
    Object = 8

class EnumEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, Enum):
            return obj.name
        return json.JSONEncoder.default(self, obj)

def check_type_equivalence(lhs, rhs) -> bool:
    if lhs == SuTypes.Any or rhs == SuTypes.Any:
        return True
    if lhs == SuTypes.Unknown or rhs == SuTypes.Unknown:
        raise TypeError("Unknown type in type equivalence check")
    return lhs == rhs

    
class KVStore:

    db = {}

    def __init__(self):
        pass

    def get(self, var) -> SuTypes | None:
        return self.db.get(var, None)

    # set if it doesn't exist already
    def set_once(self, var, value) ->  bool:
        if self.get(var) is None:
            self.db[var] = value
            return True
        return False

    def set(self, var, value):
        self.db[var] = value

class Identifier:

    def __init__(self, class_name, function_name, variable_name, scope_level):
        return f"{class_name}::{function_name}::{'@' * scope_level}{variable_name}"


def load_data_body() -> dict:

    with open('output.json') as data_file:
        data = json.load(data_file)

    return data['Methods']


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
        case _:
            raise NotImplementedError(f"missed case {stmt['Tag']}")

def infer_unary(stmt, store, graph) -> SuTypes:
    args = stmt["Args"]
    value = stmt["Value"]

    node = Node(args[0]["Value"])
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
        store.set_once(lhs["ID"], lhs_t)
    rhs_t = infer_generic(rhs, store, graph)
    if rhs_t is not None and rhs["Type_t"] != "Operator":
        store.set_once(rhs["ID"], rhs_t)

    # if not check_type_equivalence(lhs_t, rhs_t):
    #     raise TypeError(f"Type mismatch for {lhs_t} and {rhs_t}")
    lhs_n = Node(lhs["ID"])
    rhs_n = Node(rhs["ID"])
    graph.add_node(lhs_n)
    graph.add_node(rhs_n)
    graph.add_edge(lhs_n.value, rhs_n.value)

    store.set_once(lhs["ID"], lhs_t)
    store.set_once(rhs["ID"], rhs_t)

    # if lhs or rhs is of type Any, then the other type is inferred
    if lhs_t == SuTypes.Any:
        store.set(lhs["ID"], rhs_t)
        return rhs_t
    elif rhs_t == SuTypes.Any:
        store.set(rhs["ID"], lhs_t)
        return lhs_t
    

def infer_nary(stmt, store, graph) -> SuTypes:
    value = stmt["Value"]
    args = stmt["Args"]

    # all args should have the same type conforming with value of the operator
    for i in args:
        if not constraint_type_with_operator_value(value, i["Type_t"]):
            raise TypeError(f"Type mismatch for {i['Type_t']} and {value}")

    valid_t = get_valid_type_for_operator(value)

    prev = None
    for i in args:
        if i["Tag"] == "Call":
            i = i["Args"][0]
        n = Node(i["Value"])
        store.set_once(i["ID"], valid_t)
        graph.add_node(n)
        if prev is not None:
            graph.add_edge(prev.value, n.value)
        prev = n

    vn = Node(valid_t.name)
    graph.add_node(vn)
    graph.add_edge(prev.value, vn.value)

    return valid_t

def infer_if(stmt, store, graph):
    cond = stmt["Args"][0]
    cond_t = infer_generic(cond, store, graph)
    store.set(cond["ID"], cond_t)

    then = stmt["Args"][1]
    then_t = infer_generic(then, store, graph)
    store.set(then["ID"], then_t)

    if len(stmt["Args"]) == 3:
        else_t = infer_generic(stmt["Args"][2], store, graph)
        store.set(stmt["Args"][2]["ID"], else_t)

    return SuTypes.NotApplicable


def main():
    graph = Graph()
    store = KVStore()
    methods = load_data_body()

    for stmt in methods:
        infer_generic(stmt[0], store, graph)

    graph.visualise()

    # pretty print the store
    print(json.dumps(store.db, indent=4, cls=EnumEncoder))


if __name__ == "__main__":
    main()
