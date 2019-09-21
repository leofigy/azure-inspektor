package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s  <file>", os.Args[1])
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Panic(err)
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		customGen(line)
	}

}

func customGen(name string) {
	input := struct {
		Name string
	}{
		name,
	}

	os.MkdirAll("properties", os.ModePerm)

	filename := filepath.Join("properties", strcase.ToSnake(toCamel(name))+".go")
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	masterTemplate.Execute(file, input)

}

/*

{{ $constructor := constructorName .Name }}
{{ $privateConstructor := privateConstructorName .Name }}
{{ $inputAttributes := inputAttributeName .Name }}
{{ $resource := resourceName .Name }}

*/

func toCamel(name string) string {
	cname := strings.ReplaceAll(name, ".", "")
	names := strings.Split(cname, "/")
	// no boundaries check
	var camelName strings.Builder
	for _, word := range names {
		camelName.WriteString(strings.ToTitle(string(word[0])) + word[1:])
	}
	return camelName.String()
}

var structsFuncMap = template.FuncMap{
	"constructorName": func(name string) string {
		return "New" + toCamel(name)
	},
	"privateConstructorName": func(name string) string {
		return "new" + toCamel(name)
	},
	"resourceName": func(name string) string {
		capsName := toCamel(name)
		// no boundaries check
		return strings.ToLower(string(capsName[0])) + capsName[1:]
	},
	"inputAttributeName": func(name string) string {
		return toCamel(name) + "Properties"
	},
}

var masterTemplate = template.Must(template.New("").Funcs(structsFuncMap).Parse(`// This file is the glue between azurerm and hcl
package properties 

import (
	"encoding/json"

	"github.com/thetonymaster/orbital-module-azure/azurerm/properties/hcl"
)

{{ $constructor := constructorName .Name }} {{ $privateConstructor := privateConstructorName .Name }} {{ $inputAttributes := inputAttributeName .Name }} {{ $resource := resourceName .Name }}
// {{ $constructor }} returns an interface
func {{ $constructor }} () ARMProperty {
	resource := {{ $privateConstructor }}()
	return &resource
}

// for nested resources
func {{ $privateConstructor }}() {{ $resource }} {
	return {{ $resource }} {
		Input:  &{{ $inputAttributes }} {},
		// Note the below resource must be updated by you 
		Output: &hcl.AzurermLb{},
	}
}

type {{ $resource }} struct {
	Input *{{ $inputAttributes }}
	// << Modify the output please >>
	Output *hcl.AzurermLb
}

// {{ $inputAttributes }} input fiels from ARM
type {{ $inputAttributes }} struct { }

// ARMProperty Interface ---------------------------------------------------
//UnmarshalJSON of the expected input
func (r *{{ $resource }}) UnmarshalJSON(b []byte) error {
	// Populating input properties different names
	if err := json.Unmarshal(b, r.Input); err != nil {
		return err
	}
	return nil
}

//IsDataSource check if the type is a data source
func (r *{{ $resource }}) IsDataSource() bool {
	return false
}

//ARMType provider type and resource type
func (r *{{ $resource }}) ARMType() string {
	return "{{ .Name }}"
}

// END ARMProperty interface --------------------------------------------------
// TerraformRqx interface -----------------------------------

//TerraformType : terraform name
func (r *{{ $resource }}) TerraformType() string {
	return r.Output.TerraformType()
}

//Computed : types populated on compute type (after deployment)
func (r *{{ $resource }}) Computed() []string {
	return r.Output.Computed()
}

//Required : attributes that are required by terraform
func (r *{{ $resource }}) Required() []string {
	return r.Output.Required()
}

// END TerraformRqx --------------------------------

// Translator ----------------------------------

//GetReferences - get attributes that complience with the format from azure
func (r *{{ $resource }}) GetReferences() (refs map[string]string) {
	return
}

// Terraform - translate the names to the HCL syntax key method
func (r *{{ $resource }}) Translate(topObject map[string]interface{}) (map[string]interface{}, error) {
	// < Any transformation/modification to the type MicrosoftNetworkVirtualNetworksSubnets
	//  Should be done here before marshal twice to get a map>
	// initial map from input (might be good to copy Marin's function here)
	inputAsMAP := make(map[string]interface{})
	raw, _ := json.Marshal(r.Input)
	if err := json.Unmarshal(raw, &inputAsMAP); err != nil {
		return nil, err
	}

	if dependencies, ok := topObject["dependsOn"]; ok && dependencies != nil && isSlice(dependencies) {
	}

	// topObject map if comes something from input, just in case
	for key, val := range topObject {
		value, found := inputAsMAP[key]

		// avoid to topObject if already came from the properties
		if found && value != nil {
			continue
		}

		inputAsMAP[key] = val
	}
	// Any rename must be done here for any field
	// end manual part
	b, err := json.Marshal(inputAsMAP)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, r.Output)

	if err != nil {
		return nil, err
	}
	out, err := r.Output.Terraform()

	// Post marshal for all of them
	if err == nil {
		for item, val := range out {
			if val == nil {
				delete(out, item)
			}
		}

		out = mapKeysToSnake(out)

	}

	return out, err
}
`))
