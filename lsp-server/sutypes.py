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

    @staticmethod
    def from_str(str):
        match str:
            case "String":
                return SuTypes.String
            case "Number":
                return SuTypes.Number
            case "Variable":
                return SuTypes.Unknown
            case _:
                raise ValueError(f"Unknown type {str}")

    @staticmethod
    def to_str(t):
        match t:
            case SuTypes.String:
                return "String"
            case SuTypes.Number:
                return "Number"
            case SuTypes.Unknown:
                return "Variable"
            case _:
                raise ValueError(f"Unknown type {t}")


class EnumEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, Enum):
            return obj.name
        return json.JSONEncoder.default(self, obj)
