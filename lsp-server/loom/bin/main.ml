(* Import the Yojson library *)
open Yojson.Basic.Util

(* Define a JSON string *)
let json_str = "{\"name\":\"John\",\"age\":30,\"city\":\"New York\"}"

(* Parse the JSON string into a Yojson value *)
let json_value = Yojson.Basic.from_string json_str

(* Access fields from the JSON value *)
let name = json_value |> member "name" |> to_string
let age = json_value |> member "age" |> to_int
let city = json_value |> member "city" |> to_string

(* Print the extracted values *)
let () =
  print_endline ("Name: " ^ name);
  print_endline ("Age: " ^ string_of_int age);
  print_endline ("City: " ^ city)


(* custom types struct to represent types in a dynamically typed language *)
type BoxedType = 
  | kind of TypeKind
  | value of Value
  | uuid of String

