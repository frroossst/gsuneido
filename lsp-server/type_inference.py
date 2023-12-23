# Proof Of Concept
# Takes in a json object and dumps type information to a file
# Checks basic type safety

from enum import Enum
import json

class SuTypes(Enum):
    Unknown = 0
    String = 1
    Number = 2
    Boolean = 3
    

class KVStore:

    db = {}

    def __init__(self):
        pass

    def get(self, var):
        return self.db.get(var, None)


def load_data_body() -> dict:

    with open('output.json') as data_file:
        data = json.load(data_file)

    return data['Body']


def constraint_type_with_operator_value(value, type) -> bool:
    valid_constraints = {
        "Add": ["Number"],
    }

    if type == "Variable":
        return True

    valid_types = valid_constraints.get(value, None)
    return type in valid_types


def infer_generic(stmt, store):
    match stmt["Tag"]:
        case "Binary":
            infer_binary(stmt, store)
        case "Nary":
            infer_nary(stmt, store)
        case "Identifier":
            pass
        case _:
            raise NotImplementedError(f"missed case {stmt['Tag']}")

def infer_binary(stmt, store):
    args = stmt["Args"]
    lhs = args[0]
    rhs = args[1::][0]

    # inferred types of lhs and rhs should be the same

    lhs_t = infer_generic(lhs, store)
    rhs_t = infer_generic(rhs, store)

def infer_nary(stmt, store):
    print(json.dumps(stmt, indent=4))
    value = stmt["Value"]
    args = stmt["Args"]

    # all args should have the same type conforming with value of the operator
    for i in args:
        if not constraint_type_with_operator_value(value, i["Type_t"]):
            raise TypeError(f"Type mismatch for {i['Type_t']} and {value}")





def main():
    store = KVStore()
    body = load_data_body()

    for stmt in body:
        infer_generic(stmt[0], store)


if __name__ == "__main__":
    main()
