from sutypes import SuTypes



class Parser:

    src = None

    def __init__(self, src: str) -> None:
        self.src = src

    def parse(self) -> SuTypes:
        pass


def get_test_parameter_type_values():
    return {
        "originalTestFunc":  {"x": SuTypes.String, "y": SuTypes.Number, "z": SuTypes.Number},
        "SameVarID": {"x": SuTypes.String},
    }

