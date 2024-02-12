from graph import Graph
from kvstore import KVStore
from type_inference import parse_class, process_parameters, process_methods, process_custom_types, propogate_infer


def should_fail(func):
    def wrapper(*args, **kwargs):
        try:
            func(*args, **kwargs)
        except TypeError:
            pass
        else:
            print(f"[FAILED] {func.__name__} should raise a TypeError")

    return wrapper

def should_pass(func):
    def wrapper(*args, **kwargs):
        try:
            func(*args, **kwargs)
        except Exception as e:
            print(f"[FAILED] {func.__name__} should NOT raise a TypeError")
        else:
            pass

    return wrapper



@should_fail
def single_line_type_mismatch():
    global current_function
    current_function = "single_line_type_mismatch"
    src = """
    class 
    	{
    	foo() 
    		{
    		x + "123"
    		}
    	}
    """
    ast = {
        "Tag": "Class",
        "Value": "nil",
        "Type_t": "Class",
        "Args": None,
        "Name": "",
        "Base": "class",
        "ID": "1f9d089323e0446ba607abd71291c4d5",
        "Methods": {
            "foo": [
                {
                    "Tag": "Function",
                    "Value": "",
                    "Type_t": "Function",
                    "Args": None,
                    "Name": "foo",
                    "ID": "bd1962804624407fb157795292e3c609",
                    "Parameters": [],
                    "Body": [
                        [
                            {
                                "Tag": "Nary",
                                "Value": "Add",
                                "Type_t": "Operator",
                                "Args": [
                                    {
                                        "Tag": "Identifier",
                                        "Value": "x",
                                        "Type_t": "Variable",
                                        "Args": None,
                                        "ID": "2931153bb41349c08bffbe72bba204e3"
                                    },
                                    {
                                        "Tag": "Constant",
                                        "Value": "\"123\"",
                                        "Type_t": "String",
                                        "Args": None,
                                        "ID": "41643df55d98458bbf2ec0c9a99ffc26"
                                    }
                                ],
                                "ID": "94ff1f1db7644c729132bde8458aefe1"
                            }
                        ]
                    ]
                }
            ]
        },
        "Attributes": {}
    }

    attributes = parse_class(ast["Attributes"])
    methods = parse_class(ast["Methods"])
    param_t, typedefs, bindings = {}, {}, {}

    type_inference_test(methods, attributes, param_t, typedefs, bindings)


@should_fail
def single_variable_reassignment():
    global current_function
    current_function = "single_variable_reassignment"
    src = """
    class 
    	{
    	foo() 
    		{
    		x = 123
    		x = "123"
    		}
    	}
    """
    ast = {
        "Tag": "Class",
        "Value": "nil",
        "Type_t": "Class",
        "Args": None,
        "Name": "",
        "Base": "class",
        "ID": "1f9d089323e0446ba607abd71291c4d5",
        "Methods": {
            "Reassignment": [
                {
                    "Tag": "Function",
                    "Value": "",
                    "Type_t": "Function",
                    "Args": None,
                    "Name": "Reassignment",
                    "ID": "adbb73e36cf2488cb665edc41a584d69",
                    "Parameters": [],
                    "Body": [
                        [
                            {
                                "Tag": "Binary",
                                "Value": "Eq",
                                "Type_t": "Operator",
                                "Args": [
                                    {
                                        "Tag": "Identifier",
                                        "Value": "x",
                                        "Type_t": "Variable",
                                        "Args": None,
                                        "ID": "5bdd9c52ce3547d0a7542fef3604d0e1"
                                    },
                                    {
                                        "Tag": "Constant",
                                        "Value": "123",
                                        "Type_t": "Number",
                                        "Args": None,
                                        "ID": "cd071ba4c1d24664b0067ac9037a7a33"
                                    }
                                ],
                                "ID": "a1860b50572a43dbaa1720cae4d0193b"
                            }
                        ],
                        [
                            {
                                "Tag": "Binary",
                                "Value": "Eq",
                                "Type_t": "Operator",
                                "Args": [
                                    {
                                        "Tag": "Identifier",
                                        "Value": "x",
                                        "Type_t": "Variable",
                                        "Args": None,
                                        "ID": "5bdd9c52ce3547d0a7542fef3604d0e1"
                                    },
                                    {
                                        "Tag": "Constant",
                                        "Value": "\"hello\"",
                                        "Type_t": "String",
                                        "Args": None,
                                        "ID": "97f42451401b4600abff70a4fcad98fd"
                                    }
                                ],
                                "ID": "5196eea5132840249d26b50abadca825"
                            }
                        ]
                    ]
                }
            ],
        },
        "Attributes": {}
    }

    methods = parse_class(ast["Methods"])
    attributes = parse_class(ast["Attributes"])
    param_t, typedefs, bindings = {}, {}, {}

    type_inference_test(methods, attributes, param_t, typedefs, bindings)


def type_inference_test(methods, attributes, param_t, typedefs, bindings):
    graph = Graph()
    store = KVStore()

    process_parameters(methods, param_t, store, graph, attributes)
    process_custom_types(methods, typedefs, bindings, store, graph, attributes)
    process_methods(methods, store, graph, attributes)
    propogate_infer(store, graph, check=True)




def main():
    single_line_type_mismatch()
    single_variable_reassignment()
    

if __name__ == "__main__":
    main()
