


class Constraint:

    def __init__(self, left, right):
        self.left = left
        self.right = right


class TypeInference:

    unification_table = None

