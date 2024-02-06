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
    InBuiltOperator = 9
    Union = 10
    Intersect = 11

    @staticmethod
    def from_str(str):
        match str:
            case "String":
                return SuTypes.String
            case "Number":
                return SuTypes.Number
            case "Unknown":
                return SuTypes.Unknown
            case "Boolean":
                return SuTypes.Boolean
            case "Member":
                return SuTypes.Unknown
            case "Any" | "Variable":
                return SuTypes.Any
            case "Object":
                return SuTypes.Object
            case "Return":
                return SuTypes.Unknown
            case "Operator" | "PostInc" | "Callable" | "Compound" | "If" | "InBuiltOperator":
                return SuTypes.InBuiltOperator
            case "Union":
                return SuTypes.Union
            case "Intersect":
                return SuTypes.Intersect
            case _:
                raise ValueError(f"Unknown type {str} converting to SuTypes enum")

    @staticmethod
    def to_str(t):
        match t:
            case SuTypes.String:
                return "String"
            case SuTypes.Number:
                return "Number"
            case SuTypes.Unknown:
                return "Variable"
            case SuTypes.InBuiltOperator:
                return "Operator"
            case SuTypes.Boolean:
                return "Boolean"
            case SuTypes.Any:
                return "Any"
            case SuTypes.Union:
                return "Union"
            case SuTypes.Intersect:
                return "Intersect"
            case _:
                raise ValueError(f"Unknown type {t} converting to string")


class EnumEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, Enum):
            return obj.name
        return json.JSONEncoder.default(self, obj)


class UnionSuType():

    # types is a set
    types = None

    def __init__(self, args: list):
        self.types = set(args)

    def __str__(self):
        return f"UnionSuType({', '.join([str(x) for x in self.types])})"

    def __repr__(self):
        return f"UnionSuType({', '.join([str(x) for x in self.types])})"

    def add_type(self, t: SuTypes):
        self.types.add(t)

    def remove_type(self, t: SuTypes):
        self.type.remove(t)

    def unify(self):
        pass


class IntersectSuType():

    types = None

    def __init__(self) -> None:
        pass



def check_type_equality(lhs, rhs) -> bool:
    if lhs == SuTypes.Any or rhs == SuTypes.Any:
        return True
    # does not matter if one of the types is an inbuilt operator
    if lhs == SuTypes.InBuiltOperator or rhs == SuTypes.InBuiltOperator:
        return True
    if lhs == SuTypes.Unknown or rhs == SuTypes.Unknown:
        print("Unknown type not handled in type equivalence check")
        # ! remove this line
        return True
    return lhs == rhs



