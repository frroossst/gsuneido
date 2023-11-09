type ast = 
  | Function of ast list
  | Eq of ast list
  | Node of string * string * ast list

exception TypeError of string

let rec infer_type = function
  | Function(children) -> 
    List.iter infer_type children;
    "Function"
  | Eq(children) -> 
    let types = List.map infer_type children in
    if List.length (List.filter ((=) "Number") types) > 0 then
      if List.exists ((=) "String") types then
        raise (TypeError "Type mismatch: Number and String in Eq")
      else
        "Number"
    else
      "Unknown"
  | Node(value, "Unknown", children) -> 
    List.iter infer_type children;
    "Number"
  | Node(_, typ, children) -> 
    List.iter infer_type children;
    typ

let infer_ast ast =
  try 
    let _ = infer_type ast in
    "Inferred successfully"
  with TypeError msg -> 
    "Error: " ^ msg

