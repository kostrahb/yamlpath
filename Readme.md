# Yamlpath

Yamlpath is package that deals with yaml tree manipulation. It does this in a limited way and certainly does not cover all peculiarities of yaml specification (mainly composed keys, but it is probable that there are other features missing).

This implementation of yamlpath has following intuitive and simple format that lets you set and delete values in yaml tree and cannot do much more. However less is sometimes more.
```
path[2].key.arrayofarrays.[0].[3]
```

This library was written to ease yaml manipulation in golang. As of today there is no reasonable solution that would let me do simple set or delete of value in yaml tree and preserve comments. So there is a solution.

## Usage

```
package main

import (
	"fmt"
	"strings"
	"gopkg.in/yaml.v3"
	"github.com/kostrahb/yamlpath"
)

var data string = ` # Employee records
-  martin:
    name: Martin D'vloper
    job: Developer
    skills:
      - python
      - perl
      - pascal
-  tabitha:
    name: Tabitha Bitumen
    job: Developer
    skills:
      - lisp
      - fortran
      - erlang
`

func main() {
	r := strings.NewReader(data)
	root, err := yamlpath.Parse(r)
	if err != nil {
		panic(err)
	}
	err = yamlpath.Set(&root, "[+].tobi.position", "devops")
	if err != nil {
		panic(err)
	}
	err = yamlpath.Set(&root, "[2].tobi.skills", "[kubernetes, golang, python]") // Any yaml string can be passed as value
	if err != nil {
		panic(err)
	}
	err = yamlpath.Delete(&root, "[0].martin.skills[2]")
	if err != nil {
		panic(err)
	}

	output, err := yaml.Marshal(&root)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}
```

## Yamlpath structure

Structure of yaml path is following, each level is separated by dot and on each level there are few possibilities
* Map key has following regexp `[[:word:]]+`, so there can be only alphanumeric characters
* Sequence on the other hand is a little bit more interesting. It can be either number or plus or minus signs denoted by parentheses []. Minus sign is used only in Delete function and denotes that last element of array should be deleted, plus sign on the other hand is used only in Set function and denotes that the element should be added at the end of the array. `([[:word:]]*)\\[([[:digit:]+-]+)\\]`
* You can merge exactly one key for map with one array index. If you need to access array of arrays, you need to use dot syntax: `key[1].[2]`

## Notes

* If the number is overflowing the array size, element is added on the end in Set or nothing is deleted in case of Delete.

