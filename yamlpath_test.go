package yamlpath

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

func pokus() *yaml.Node {
	doc := createDoc()
	maap := createMap()
	key := createStr("key")
	value := createStr("value")
	seq := createSeq()
	first := createStr("first")
	doc.Content = append(doc.Content, maap)
	maap.Content = append(maap.Content, key, value)
	seq.Content = append(seq.Content, first)
	return doc
}

func print(root *yaml.Node, num int) {
	fmt.Printf("%v: %#v\n", num, root)
	for _, child := range root.Content {
		print(child, num+1)
	}
}

func main() {
	in, err := os.Open("example.yml")
	if err != nil {
		panic(err)
	}
	defer in.Close()

	doc, err := Parse(in)
	if err != nil {
		panic(err)
	}

	out, err := os.Create("example1.yml")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	err = Save(doc, out)
	if err != nil {
		panic(err)
	}

	var d yaml.Node
	err = yaml.Unmarshal([]byte(data), &d)
	if err != nil {
		panic(err)
	}


	fmt.Println(isArraySet.MatchString("[0]"))
	fmt.Println(isArraySet.MatchString("asdf[0]"))
	fmt.Println(isArraySet.MatchString("asdf[+]"))
	fmt.Println(isArraySet.MatchString("[+]"))

	e := pokus()
//	print(e, 0)
//	Set(e, "key", "a")
	fmt.Println("aaaaaaaaaaaaaaaaaaaaa")
	err = Set(e, "asdf[0].martin.position", "devops")
	fmt.Println("first",err)

	err = Set(e, "asdf[3].martin.position", "devops")
	fmt.Println("first",err)

	err = Set(e, "qwer.[0].[+].position", "devops")
	fmt.Println("second",err)

	err = Set(&d, "asdf[0].martin.position", "devops")
	fmt.Println("third",err)
	err = Set(&d, "[0].martin.job", "devops")
	fmt.Println("third",err)
	err = Set(&d, "[0].martin.skills[5]", "devops")
	fmt.Println("third",err)
	err = Set(&d, "[0].martin.skills[2]", "kubernetes")
	fmt.Println("third",err)
	fmt.Println("third",err)
	err = Set(&d, "[0].martin.skills[6].asdf[3].c", "qwer")

	err = Set(&d, "qwer.[0].position", "devops")
	fmt.Println("fourth", err)
	err = Set(&d, "[0].martin.position", "devops")
	if err != nil {
		panic(err)
	}

	err = Delete(e, "qwer.[0].[0]")

	err = Delete(&d, "[0].martin.position")
	fmt.Println(err)

//	print(&d, 0)
	output, err := yaml.Marshal(&d)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))

	output, err = yaml.Marshal(e)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
	err = Delete(e, "asdf[1].martin.position")
	err = Delete(e, "asdf[0]")
	fmt.Println(err)
	output, err = yaml.Marshal(e)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}
