.PHONY: *

help:
	@echo "Please use 'make <target>' where <target> is one of"
	@echo "    run      to run compile, infer, err_check in sequence"
	@echo "    compile  to run the main.go file and generate the AST output"
	@echo "    infer    to run the type_inference.py file and generate the store and graph output"
	@echo "    check    to run the type_checker.py file to perform type checking"
	@echo "    test      to run the unit_test.py file to perform unit testing"
	@echo "	   todo      to display the todo list"
	@echo "    err_check    similar to check but only prints the errors"

run: compile infer err_check

compile:
	go run main.go
	python3 prettify.py

infer:
	python3 type_inference.py -t

check:
	python3 type_checker.py

todo:
	bat todo.list

test:
	python3 unit_test.py

err_check:
	python3 type_checker.py | grep ERROR || true

