type ast =
  | Function of ast
  | Binary of string * ast * ast
  | Variable of string
  | Integer of int

type inferred_type =
  | Unknown
  | Integer
  | Function of inferred_type * inferred_type

let rec infer_type = function
  | Function(body) -> Function(Unknown, infer_type body)
  | Binary(op, left, right) ->
    let left_type = infer_type left in
    let right_type = infer_type right in
    if left_type = right_type then
      match op with
      | "Eq" -> Integer
      | _ -> failwith ("Unknown operator: " ^ op)
    else
      failwith "Type error: operands must be of the same type"
  | Variable(_) -> Unknown
  | Integer(_) -> Integer

