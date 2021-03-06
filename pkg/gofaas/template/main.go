package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"os"

	"github.com/schollz/faas/pkg/gofaas"
	log "github.com/schollz/logger"

	// start generated code
	{{with .ImportPath}}"{{.}}"{{end}}
	// end generated code
)

// start generated code
const functionNameToRun = "{{.FunctionName}}"
var paramNames = []string{ {{range $index, $element := .InputParams }}{{if $index}}, {{end}}"{{$element.Name}}"{{ end }} }
type Input struct {
	{{- range .InputParams }}
	{{title .Name }} {{.Type }}  `json:"{{.Name}}"`{{end}}
}
const userCors = true
// end generated code
var startTime time.Time


func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()
	if debug {
		log.SetLevel("debug")
	} else {
		log.SetLevel("info")
	}

		startTime = time.Now()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			log.Debugf("checking time: %2.1f > 3000?", time.Since(startTime).Seconds())
			if time.Since(startTime).Seconds() > 3000 {
				os.Exit(0)
			}
		}
	}()
	log.Infof("running on port %s", "8080")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	timeStart := time.Now()
	defer func() {
		log.Infof("%s?%s %s", r.URL.Path, r.URL.RawQuery, time.Since(timeStart))
	}()

	var response []byte
	var err error
	var input Input
	bUsage, _ := json.Marshal(input)
	response, _ = json.Marshal(struct {
		Usage   string `json:"usage"`
		Message string `json:"message"`
		Success bool   `json:"success"`
	}{
		string(bUsage),
		"usage",
		true,
	})
	if r.Method == "GET" {
		response, err = handleGet(w, r)
	} else if r.Method == "POST" {
		response, err = handlePost(w, r)
	}
	if err != nil {
		response, _ = json.Marshal(struct {
			Usage   string `json:"usage"`
			Message string `json:"message"`
			Success bool   `json:"success"`
		}{
			string(bUsage),
			err.Error(),
			false,
		})
	}
	if userCors {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	{{ if eq 1 (len .OutputParams) }}
	w.Header().Set("Content-Type", "text/plain")
	{{else}}
	w.Header().Set("Content-Type", "text/javascript")
	{{end}}
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.Write(response)
}

func handlePost(w http.ResponseWriter, r *http.Request) (response []byte, err error) {
	decoder := json.NewDecoder(r.Body)
	var input Input
	err = decoder.Decode(&input)
	if err != nil {
		log.Error(err)
		return
	}
	return getResponse(input)
}

func handleGet(w http.ResponseWriter, r *http.Request) (response []byte, err error) {
	log.Debug(r.URL.RawQuery)
	funcString, ok := r.URL.Query()["func"]
	if !ok {
		err = fmt.Errorf("no func string")
		log.Error(err)
		return
	}

	log.Debug(funcString)
	functionName, jsonBytes, err := gofaas.ParseFunctionString(paramNames, funcString[0])
	if err != nil {
		log.Error(err)
		return
	}

	if functionNameToRun != functionName {
		err = fmt.Errorf("mismatched functions")
		log.Error(err)
		return
	}

	var input Input
	err = json.Unmarshal(jsonBytes, &input)
	if err != nil {
		log.Error(err)
		return
	}

	return getResponse(input)
}

func getResponse(input Input) (response []byte, err error) {
	// start generated code
	{{range $index, $element := .OutputParams }}{{if $index}}, {{end}}out{{$index}}{{ end }} := {{with .PackageName}}{{.}}.{{end}}{{.FunctionName}}({{range $index, $element := .InputParams }}{{if $index}}, {{end}}input.{{title $element.Name}}{{ end }})

// if only one output parm, then just marshal it and print it
{{ if eq 1 (len .OutputParams) }}
	{{range $index, $element := .OutputParams }}
response= []byte(fmt.Sprintf("%+v",(out{{$index}})))
	{{end}}
{{else }}
	var b []byte
	responseString := ""
	{{range $index, $element := .OutputParams }}
	{{if $index}}responseString += ","{{end}}
	b, err = json.Marshal(out{{$index}})
	if err != nil {
		log.Error(err)
		return
	}
	responseString += `"` + "{{if $element.Name}}{{$element.Name}}{{else}}{{$index}}{{end}}" + `"` + ": " + string(b)
	{{end}}
	responseString = "{" + responseString + "}"
	response = []byte(responseString)
{{end}}
	// end generated code

	return
}
