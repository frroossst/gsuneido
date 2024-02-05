def todo(str = ""):
    raise NotImplementedError(f"TODO {str}")

def unimplemented(str = ""):
    raise NotImplementedError(f"Unimplemented: {str}")

def catch_exception(func):
    def wrapper(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            print(f"Exception: {e}")
    return wrapper
