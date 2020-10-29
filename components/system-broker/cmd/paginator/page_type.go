/*
 * Copyright 2018 The Service Manager Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"html/template"
	"os"
	"strings"
)

const StorageTypesDirectory = "github.com/Peripli/service-manager/storage/postgres"

type PageType struct {
	PackageName string
	Type        string
	OutputType  string
	DataPath    string
}

func GeneratePageEntityFile(fileDir, typeName, packageName, outputType, dataPath string) error {
	t := template.Must(template.New("generate-page-type").Parse(PageTypeTemplate))
	entityTemplate := PageType{
		Type:        typeName,
		PackageName: packageName,
		OutputType:  outputType,
		DataPath:    dataPath,
	}
	file, err := os.Create(fmt.Sprintf("%s/%s_gen.go", fileDir, strings.ToLower(typeName)))
	if err != nil {
		return err
	}
	if err = t.Execute(file, entityTemplate); err != nil {
		return err
	}
	return nil
}

// func toPlural(typeName string) string {
// 	typeNamePlural := fmt.Sprintf("%ss", typeName)
// 	if strings.HasSuffix(typeName, "y") {
// 		typeNamePlural = fmt.Sprintf("%sies", typeName[:len(typeName)-1])
// 	}
// 	return typeNamePlural
// }

// func toLowerSnakeCase(str string) string {
// 	builder := strings.Builder{}
// 	for i, char := range str {
// 		if unicode.IsUpper(char) && i > 0 {
// 			builder.WriteRune('_')
// 		}
// 		builder.WriteRune(unicode.ToLower(char))
// 	}
// 	return builder.String()
// }
