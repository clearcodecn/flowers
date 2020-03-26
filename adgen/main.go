package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
)

type File struct {
	Domain []string `json:"domain"`
}

func main() {
	data, _ := ioutil.ReadFile("adgen/ad.json")

	f := new(File)
	err := json.Unmarshal(data, f)
	if err != nil {
		log.Fatalln(err)
	}
	tpl, _ := template.New("ad").Parse(adTpl)

	buf := bytes.NewBuffer(nil)

	tpl.Execute(buf, map[string]interface{}{
		"Ad": f.Domain,
	})

	ioutil.WriteFile("ad/ad.go", buf.Bytes(), 0777)
}

var (
	adTpl = `package ad
// generate by adgen, DO NOT EDIT.

func FilterAdDomain() func(host string) bool {
	var (
		adMap = make(map[string]struct{})
	)
	{{- range .Ad }}
	adMap["{{ . }}"] = struct{}{}
	{{- end }}
	return func(host string) bool {
		if _,ok := adMap[host]; ok {
			return true
		}
		return false
	}
}
`
)
