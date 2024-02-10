def single_line_type_mismatch():
    src = """
    class 
    	{
    	foo() 
    		{
    		x + "123"
    		}
    	}
    """
    ast = """
    {
        "Tag": "Class",
        "Value": "nil",
        "Type_t": "Class",
        "Args": null,
        "Name": "",
        "Base": "class",
        "ID": "1f9d089323e0446ba607abd71291c4d5",
        "Methods": {
            "foo": [
                {
                    "Tag": "Function",
                    "Value": "",
                    "Type_t": "Function",
                    "Args": null,
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
                                        "Args": null,
                                        "ID": "2931153bb41349c08bffbe72bba204e3"
                                    },
                                    {
                                        "Tag": "Constant",
                                        "Value": "\"123\"",
                                        "Type_t": "String",
                                        "Args": null,
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
    """

class TestFailedError(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

def main():
    single_line_type_mismatch()

if __name__ == "__main__":
    main()
