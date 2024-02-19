from sutypes import SuTypes, TypeRepr



class StoreValue:

    def __init__(self, value, actual: TypeRepr, inferred: TypeRepr) -> None:
        """
        @param value: The name of the variable
        @param actual: The actual type of the variable
        @param inferred: The inferred type of the variable
        """
        if not isinstance(actual, TypeRepr) or not isinstance(inferred, TypeRepr):
            raise ValueError("actual and inferred should be of type TypeRepr")

        self.value = value
        self.actual = actual
        self.inferred = inferred

    def __repr__(self) -> str:
        return f"Value(value = {self.value}, actual = {self.actual}, inferred = {self.inferred})"

    def to_json(self) -> dict:
        return {
            "value": self.value,
            "actual": str(self.actual),
            "inferred": str(self.inferred)
        }
        
class SymbolTable:
    """
    This stores the mapping of one variable to multiple UUIDs 
    To help KVStore lookup by variable key
    It is a two way mapping
    """



class KVStore:

    db = {}

    def __init__(self):
        pass

    def to_json(self) -> str:
        json_data = {}
        for k, v in self.db.items():
            json_data[k] = v.to_json()
        return json_data

    def get(self, var) -> SuTypes | None:
        return self.db.get(var, None)

    def set(self, var_id, value) ->  bool:
        if not isinstance(value, StoreValue):
            raise ValueError("Value should be of type StoreVale")

        if (curr_val := self.get(var_id)) is None:
            self.db[var_id] = value
        elif curr_val is not None:
            # check if the new type is a subtype of the existing type
            # or if it is an equivalent type
            # if not check_type_equal_or_subtype(value.inferred, curr_val.inferred):
            if curr_val.inferred == TypeRepr(TypeRepr.construct_definition_from_primitive(SuTypes.Any)):
                # if current is Any then it should be overwritable
                pass
            elif not (curr_val.inferred <= value.inferred):
                raise TypeError(f"Conflicting inferred types for variable {var_id}\nexisting: {curr_val.inferred}, got: {value.inferred}") 
            else:
                self.db[var_id] = value
        else:
            raise ValueError(f"Variable already exists in the store\nexists: {self.get(var_id)},\ngot: {value}")

    def set_on_type_equivalence(self, var_id, value, check=False):
        if (val := self.get(var_id)) is None:
            self.set(var_id, value)
            return

        if not (val.inferred <= value.inferred):
            raise TypeError(f"Conflicting inferred types for variable {var_id}\nexisting: {val.inferred}, \ngot: {value.inferred}")

        if val.actual == value.inferred:
            self.set(var_id, value)
            return
        else:
            # this error is only thrown when type checking not during type inference
            if check:
                raise TypeError(f"For type node {var_id} type is inferred as {value.actual} but declared as {val.actual} \n{val}")



    @classmethod
    def from_json(cls, json_data: dict):
        kv_instance = cls()

        for k, v in json_data.items():
            value = v.get("value")
            actual = SuTypes.from_str(v.get("actual"))
            inferred = SuTypes.from_str(v.get("inferred"))
            store_value = StoreValue(value, actual, inferred)
            kv_instance.set(k, store_value)

        return kv_instance