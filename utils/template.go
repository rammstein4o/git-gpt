// ---
// Copyright © 2023 Radoslav Salov <rado.salov@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
// ---

package utils

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"strings"
)

// Data defines a custom type for the template data.
type Data map[string]interface{}

var (
	templates    map[string]*template.Template
	templatesDir = "templates"
)

func NewTemplateByString(format string, data map[string]interface{}) (string, error) {
	t, err := template.New("message").Parse(format)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

// processTemplate processes the template with the given name and data.
func processTemplate(name string, data map[string]interface{}) (*bytes.Buffer, error) {
	t, ok := templates[name]
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}

	var tpl bytes.Buffer

	if err := t.Execute(&tpl, data); err != nil {
		return nil, err
	}

	return &tpl, nil
}

// GetTemplateByString returns the parsed template as a string.
func GetTemplateByString(name string, data map[string]interface{}) (string, error) {
	tpl, err := processTemplate(name, data)
	return strings.TrimSpace(tpl.String()), err
}

// GetTemplateByBytes returns the parsed template as a byte.
func GetTemplateByBytes(name string, data map[string]interface{}) ([]byte, error) {
	tpl, err := processTemplate(name, data)
	return bytes.TrimSpace(tpl.Bytes()), err
}

// LoadTemplates loads all the templates found in the templates directory.
func LoadTemplates(files embed.FS) error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(files, templatesDir)
	if err != nil {
		return err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(files, templatesDir+"/"+tmpl.Name())
		if err != nil {
			return err
		}

		templates[tmpl.Name()] = pt
	}
	return nil
}
