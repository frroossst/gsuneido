from typing import Any
import json

from sutypes import SuTypes


class StoreValue:

    def __init__(self, value, actual: SuTypes, inferred: SuTypes) -> None:
        """
        @param value: The value of the variable
        @param actual: The actual type of the variable
        @param inferred: The inferred type of the variable
        """
        self.value = value
        self.actual = actual
        self.inferred = inferred

    def __repr__(self) -> str:
        return f"Value(value = {self.value}, actual = {self.actual}, inferred = {self.inferred})"

    def to_json(self) -> dict:
        if isinstance(self.actual, SuTypes):
            actual = self.actual.name
        else:
            actual = self.actual

        if isinstance(self.inferred, SuTypes):
            inferred = self.inferred.name
        else:
            inferred = self.inferred
        
        return {
            "value": self.value,
            "actual": actual,
            "inferred": inferred,
        }


class KVStore:

    db = {}

    def __init__(self):
        pass

    def to_json(self) -> str:
        json_data = {}
        for k, v in self.db.items():
            json_data[k] = v.to_json()
        return json.dumps(json_data, indent=4)

    def get(self, var) -> SuTypes | None:
        return self.db.get(var, None)

    # set if it doesn't exist already
    def set(self, var_id, value) ->  bool:
        if not isinstance(value, StoreValue):
            raise TypeError("Value should be of type Value")

        if self.get(var_id) is None:
            self.db[var_id] = value
        elif (val := self.get(var_id)) is not None:
            if value.actual == val.actual and value.inferred == val.inferred and value.value == val.value:
                pass
        else:
            raise ValueError(f"Variable already exists in the store\nexists: {self.get(var_id)},\ngot: {value}")
