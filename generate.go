package goapi

import (
	"embed"
	"os"
	"text/template"
)

//go:embed template.go.tmpl
var templateFS embed.FS

func generate(api *API) error {

	errorName := func() string {
		return api.errorScheme
	}

	functions := template.FuncMap{
		"schemePrefix": schemePrefix,
		"errorName":    errorName,
	}

	tmpl, err := template.New("template.go.tmpl").
		Funcs(functions).
		ParseFS(templateFS, "template.go.tmpl")

	if err != nil {
		return err
	}

	file, err := os.Create("openapi.yaml")

	if err != nil {
		return err
	}

	defer file.Close()

	err = tmpl.Execute(file, *api)

	if err != nil {
		return err
	}

	return nil
}
