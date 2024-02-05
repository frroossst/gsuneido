import json

with open("output.json", "r") as fobj:
    content = json.load(fobj)

with open("output.json", "w") as fobj:
    json.dump(content, fobj, indent=4)
