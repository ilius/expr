package main

import (
	"fmt"

	"github.com/ilius/expr/docgen"
	"github.com/ilius/expr/test/real_world"
)

func main() {
	doc := docgen.CreateDoc(real_world.NewEnv())

	fmt.Println(doc.Markdown())
}
