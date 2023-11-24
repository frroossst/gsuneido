# Proof Of Concept
# Takes in a json object and dumps type information to a file
# Checks basic type safety

import json



class TypeInference:

    def __init__(self):
        self.type_store = {}

    def to_json(self):
        return json.dumps(self.type_store, indent=6)

    def set_type(self, key, value):
        self.type_store[key] = value

    def get_type(self, key):
        try:
            return self.type_store[key]
        except KeyError:
            return "Unknown"
    

def load_json(file_name):
    with open(file_name, "r") as fobj:
        content = json.load(fobj)
    return content

# infers constant like numbers, strings, booleans
# adds to global type store to be later associated
# with variables
def infer_constants(content, store):
    for k, v in content.items():
        if check_number(k):
            print("Number: " + k)
            store.set_type(k, "Number")
        elif check_string(k):
            print("String: " + k)
            store.set_type(k, "String")
        elif check_alnum(k):
            print("Variable: " + k)
            store.set_type(k, "Variable")
        elif "(" in k and ")" in k:
            if is_solveable(k):
                store.set_type(k, "Solveable")
            else:
                store.set_type(k, "Operator")
            print("Solveable or Operator: " + k)
        else:
            print("Unknown: " + k)
            raise TypeError("Unknown type: " + k)


def check_number(value):
    try:
        int(value)
        return True
    except ValueError:
        return False

def check_string(value):
    if value[0] == '"' and value[-1] == '"':
        return True
    else:
        return False

# checks if value is either a mix of letters and numbers
# returns false if contains parenthesis 
def check_alnum(value):
    if "(" in value or ")" in value:
        return False
    elif value.isalnum():
        return True
    return False


# ==============================================================================
def run_inference(content, store):
    for k, v in content.items():
        if store.get_type(k) == "Solveable":
            solve_solveables(store, k)
        elif store.get_type(k) == "Operator":
            solve_operators(k, [], store)


# ==============================================================================
# Solveables
# ==============================================================================

# mark solveable if starts with Unary, Binary, or Nary
def is_solveable(value):
    if value.startswith("Unary") or value.startswith("Binary") or value.startswith("Nary"):  # noqa: E501
        return True
    return False

def solve_solveables(store, operator):
    if operator.startswith("Binary"):
        return solve_binary(operator, store)
    elif operator.startswith("Unary"):
        return solve_unary(operator, store)
    elif operator.startswith("Nary"):
        return solve_nary(operator, store)
    else:
        raise TypeError("Unknown operator: " + operator)

def solve_binary(operator, store):
    # parse binary 
    operator = operator[7:]
    operator = operator[:-1]
    print("Solving Binary: " + operator)
    # Eq num Nary(Add x "123")
    # left = num
    # right = Nary(Add x "123")
    # operator = Eq
    operator = operator.split(" ")
    ops = operator[0]
    lhs = operator[1]
    rhs = " ".join(operator[2:])

    print(f"Ops: {ops}, LHS: {lhs}, RHS: {rhs}")

    if is_solveable(rhs):
        rhs_type = solve_solveables(store, rhs)
    else:
        raise TypeError("Not solveable: " + rhs)

    store.set_type(lhs, rhs_type)


def solve_unary(operator, store):
    pass

def solve_nary(operator, store):
    # parse nary
    operator = operator[5:]
    operator = operator[:-1]

    operator = operator.split(" ")
    ops = operator[0]
    args = operator[1:]


    print(f"Ops: {ops}, Args: {args}")

    return solve_operators(ops, args, store)

# ==============================================================================
# Operators
# ==============================================================================
def solve_operators(operator, args,  store):
    if operator == "Eq":
        pass
    elif operator == "Add":
        return solve_add(args, store)
    elif operator.startswith("Call"):
        args = operator[5:-1:]
        return solve_call(args, store)

# all args must be of type Number
def solve_add(args, store):
    for i in args:
        if store.get_type(i) == "Variable":
            store.set_type(i, "Number")
        elif store.get_type(i) != "Number":
            err = f"For operator Add, {i} is not of type Number, it is of type {store.get_type(i)}"  # noqa: E501
            raise TypeError(err)

    return "Number"

# all args must be of type Function
def solve_call(args, store):
    if store.get_type(args) != "Function":
        err = f"For operator Call, {args} is not of type Function, it is of type {store.get_type(args)}" # noqa: E501
        raise TypeError(err)


def main():
    content = load_json("output.json")

    store = TypeInference()

    infer_constants(content, store)

    run_inference(content, store)

    print(store.to_json())

if __name__ == "__main__":
    main()
