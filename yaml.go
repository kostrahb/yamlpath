// yamlpath is package that deals with yaml tree manipulation. It does this in a limited way and certainly does not cover all peculiarities of yaml specification (mainly composed keys, but it is probable that there are other features missing).
// This implementation of yamlpath has following intuitive and simple format and cannot do much more. However less is sometimes more.
//  path[2].key.arrayofarrays.[0].[3]
// Structure of yaml path is following: Each level is separated by dot and on each level there are few possibilities
// Map key has following regexp, so there can be only alphanumeric characters
//  [[:word:]]+
// Sequence on the other hand is a little bit more interesting. It can be either number or plus or minus signs denoted by parentheses []. Minus sign is used only in Delete function and denotes that last element of array should be deleted, plus sign on the other hand is used only in Set function and denotes that the element should be added at the end of the array.
//  ([[:word:]]*)\\[([[:digit:]+-]+)\\]
// If the number is overflowing the array size, element is added on the end in Set or nothing is deleted in case of Delete.
package yamlpath

import (
	"io"
	"fmt"
	"strconv"
	"strings"
	"regexp"
	"gopkg.in/yaml.v3"
)

// Parse is helper function that decodes yaml from io.Reader
func Parse(reader io.Reader) (yaml.Node, error) {
	d := yaml.NewDecoder(reader)
	var n yaml.Node
	err := d.Decode(&n)
	if err != nil {
		return n, err
	}
	return n, nil
}

// Save is helper function that writes yaml tree to io.Writer
func Save(doc yaml.Node, writer io.Writer) error {
	e := yaml.NewEncoder(writer)
	err := e.Encode(&doc)
	if err != nil {
		return err
	}
	return e.Close()
}

// Helper function that creates string yaml node
func createStr(value string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str"}
	n.SetString(value)
	return n
}

// Helper function that creates sequence yaml node
func createSeq() *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
}

// Helper function that creates map yaml node
func createMap() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}

// Helper function that creates document yaml node
func createDoc() *yaml.Node {
	return &yaml.Node{Kind: yaml.DocumentNode, Tag: ""}
}


// Merge tries to merge two yaml nodes a and b into a. If it encounters incompatible nodes along the way, it returns error.
func Merge(a, b *yaml.Node) error {
	// Document nodes are simple, just merge first children
	if (a.Kind == yaml.DocumentNode && b.Kind == yaml.DocumentNode) {
		return Merge(a.Content[0], b.Content[0])
	}

	// Scalar nodes are also simple, just replace them
	if (a.Kind == yaml.ScalarNode || b.Kind == yaml.ScalarNode) {
		*a = *b
	}

	// Sequences are trickier. We need to check whether the index is in the array. If it is not or if it is negative, we append new element, otherwise merge recursively
	// Little hack, we use spare variable in yaml.Node to signalize on which position in sequence we shall merge it
	if (a.Kind == yaml.SequenceNode && b.Kind == yaml.SequenceNode) {
		// if the index is out of bounds, just add the element 
		if b.Line < 0 || len(a.Content) < b.Line {
			a.Content = append(a.Content, b.Content...)
		} else {
			Merge(a.Content[b.Line], b.Content[0])
		}
		return nil
	}

	// We need to go through all content of a map and check whether the key is there. If yes, we merge it recursively, otherwise just append new key
	if (a.Kind == yaml.MappingNode && b.Kind == yaml.MappingNode) {
		for i := 0; i< len(b.Content); i+=2 {
			for j := 0; j< len(a.Content); j+=2 {
				if b.Content[i].Value == a.Content[j].Value {
					return Merge(a.Content[j+1], b.Content[i+1])
				}
			}
		}
		a.Content = append(a.Content, b.Content...)
		return nil
	}

	return fmt.Errorf("Cannot merge yaml nodes: They contain incompatible nodes")
}


var isArraySet = regexp.MustCompile("([[:word:]]*)\\[([[:digit:]\\+]+)\\]")
var isArrayDel = regexp.MustCompile("([[:word:]]*)\\[([[:digit:]-]+)\\]")
var isSimple = regexp.MustCompile("[[:word:]]+")


func set(root *yaml.Node, xpath string, value *yaml.Node) error {
	steps := strings.Split(xpath, ".")
	if len(steps) < 1 {
		root = value
	}
	current := value.Content[0]
	// To ensure correctness and easy algorithm, we start at the end of an array and go through it in reverse order
	// Srsly, this is much easier, just imagine how many cases there could be if we go the other way. Blah!
	for i := len(steps)-1; i >= 0; i-- {
		step := steps[i]
		if isArraySet.MatchString(step) {
			substrings := isArraySet.FindAllStringSubmatch(step, -1)
			key := substrings[0][1]
			pos := substrings[0][2]

			// Prepare sequence node
			// Little hack, we use spare variable in yaml.Node to signalize on which position in sequence we shall merge it
			seq := createSeq()
			if pos == "+" {
				seq.Line = -1
			} else {
				i, err := strconv.Atoi(pos)
				if err != nil {
					return err
				}
				seq.Line = i
			}
			seq.Content = []*yaml.Node{current}

			// Create needed structure - it differs if we are in root and no key is specified
			var tmp *yaml.Node
			if key == "" {
				tmp = seq
			} else {
				keynode := createStr(key)
				tmp = createMap()
				tmp.Content = []*yaml.Node{keynode, seq}
			}
			current = tmp
		} else if isSimple.MatchString(step) {
			// maps are straightforward
			if step == "" {
				return fmt.Errorf("Empty path section is allowed only in root for sequence")
			}
			tmp := createMap()
			key := createStr(step)
			tmp.Content = []*yaml.Node{key, current}
			current = tmp
		} else {
			return fmt.Errorf("Xpath %v is not valid", xpath)
		}
	}
	doc := createDoc()
	doc.Content = append(doc.Content, current)
	return Merge(root, doc)
}

// Set uses equivalent of xpath and sets the element to yaml/json value. All needed nodes are created as needed.
func Set(root *yaml.Node, xpath, value string) error {
	var v yaml.Node
	err := yaml.Unmarshal([]byte(value), &v)
	if err != nil {
		return err
	}
	return set(root, xpath, &v)
}

/*
func getKey(n *yaml.Node, key string) (*yaml.Node, error) {
	if n.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("Node is not a map")
	}
	for i := 0; i < len(n.Content); i+=2 {
		if n.Content[i].Value == key {
			return n.Content[i+1], nil
		}
	}
	return nil, nil
}

func getNthNode(n *yaml.Node, i int) (*yaml.Node, error) {
	if n.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("Node is not a sequence")
	}
	if i > len(n) {
		return nil, nil
	}
	return n.Content[i], nil
}*/

// Delete deletes single node from yaml file according to equivalent of xpath
func Delete(root *yaml.Node, xpath string) error {
	var err error

	// If document node is passed, just delete recursively for its child
	if root.Kind == yaml.DocumentNode {
		return Delete(root.Content[0], xpath)
	}

	// Get first part of xpath
	steps := strings.SplitN(xpath, ".", 2)

	// We are dealing with array expression -> we have either map or sequence node
	if isArrayDel.MatchString(steps[0]) {
		substrings := isArrayDel.FindAllStringSubmatch(steps[0], -1)
		key := substrings[0][1]
		pos := substrings[0][2]

		// Check for mismatches
		if key == "" && root.Kind != yaml.SequenceNode {
			return fmt.Errorf("Tried accessing sequence but node is not sequence")
		}

		if key != "" && root.Kind != yaml.MappingNode {
			return fmt.Errorf("Tried accessing map but node is not map")
		}

		// Get position in array
		i := 0
		if pos == "-" {
			i = len(root.Content)-1
		} else {
			i, err = strconv.Atoi(pos)
			if err != nil {
				return err
			}
			if i > len(root.Content)-1 {
				return nil
			}
		}

		// Heavy lifting. Either the key is empty or not and either we are at the end of the xpath so we shall delete or not
		if key == "" {
			if len(steps) < 2 {
				root.Content = append(root.Content[:i], root.Content[i+1:]...)
				return nil
			} else {
				return Delete(root.Content[i], steps[1])
			}
		} else {
			// Find required key inside map
			for j := 0; j<len(root.Content); j+=2 {
				if root.Content[j].Value == key {
					if len(steps) < 2 {
						root.Content[j+1].Content = append(root.Content[j+1].Content[:i], root.Content[j+1].Content[i+1:]...)
						return nil
					}
					return Delete(root.Content[j+1].Content[i], steps[1])
				}
			}
			// If not found, there is nothing to delete
			return nil
		}
		// Just in case
		return nil
	}

	// We are dealing with simple key expression -> we have map
	if isSimple.MatchString(steps[0]) {
		// Quick check
		if root.Kind != yaml.MappingNode {
			return fmt.Errorf("Cannot access non map node with key")
		}

		// Let'S find the key
		for j := 0; j<len(root.Content); j+=2 {
			if root.Content[j].Value == steps[0] {
				// If we are at the end of xpath, we delete, otherwise delete recursively
				if len(steps) < 2 {
					root.Content = append(root.Content[:j], root.Content[j+2:]...)
					break
				}
				return Delete(root.Content[j+1], steps[1])
			}
		}
		return nil
	}

	// xpath invalid
	return fmt.Errorf("Xpath %v is not valid", xpath)
}
