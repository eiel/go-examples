package main

import (
	"fmt"
	"os"
	"text/template"
)

func main() {
	filename := "new_and_init_test.go"
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	funcs := template.FuncMap{
		"Items": func(n int) []int {
			array := make([]int, n)
			for i := range array {
				array[i] = i+1
			}
			return array
		},
	}
	// Templateの名前をファイル名と一致させる必要がある。が、違うディレクトリにある場合ParseFilesはディレクトリ名を含める必要がある
	name := "initial_template.gotmpl"
	t, err := template.New(name).Funcs(funcs).ParseFiles(fmt.Sprintf("gen/%s", name))
	if err != nil {
		panic(err)
	}
	type Data struct {
		Length int
	}
	err = t.Execute(f, Data{Length: 60})

	if err != nil {
		panic(err)
	}
}
