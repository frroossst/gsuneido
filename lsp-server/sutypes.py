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

    @staticmethod
    def from_str(str):
        match str:
            case "String":
                return SuTypes.String
            case "Number":
                return SuTypes.Number
            case "Variable" | "Member":
                return SuTypes.Unknown
            case "Return" | "Operator" | "PostInc" | "Callable" | "Compound" | "If":
                return SuTypes.InBuiltOperator
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
            case _:
                raise ValueError(f"Unknown type {t} converting to string")


class EnumEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, Enum):
            return obj.name
        return json.JSONEncoder.default(self, obj)
