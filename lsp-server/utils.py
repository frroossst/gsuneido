import argparse

def todo(str = ""):
    raise NotImplementedError(f"TODO {str}")

def unimplemented(str = ""):
    raise NotImplementedError(f"Unimplemented: {str}")

def catch_exception(func):
    def wrapper(*args, **kwargs):
        try:
            return func(*args, **kwargs)
        except Exception as e:
            p = argparse.ArgumentParser("Type Inference")
            p.add_argument("-t", action="store_true")
            args = p.parse_args()
            if not args.t:
                print(f"Exception: {e}")
            else:
                raise RuntimeError(e) from e
            return
    return wrapper
