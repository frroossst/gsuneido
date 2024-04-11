import json

def add_custom_types_to_db():
    pass

def add_func_signatures_to_db():
    pass

def add_predefined_var_types_to_db():
    pass



def main():

    ast = json.load(open("ast.json"))
    methods = {k: v[0] for k, v in ast["Methods"].items()}
    attributes = {k: v[0] for k, v in ast["Attributes"].items()}

    add_custom_types_to_db()
    add_func_signatures_to_db()
    add_predefined_var_types_to_db()





if __name__ == "__main__":
    main()