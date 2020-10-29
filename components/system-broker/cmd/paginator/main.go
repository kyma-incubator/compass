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
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 3 {
		panic("Usage is <type> <output-type> <data-path>")
	}
	typeName := args[0]
	outputTypeName := args[1]
	dataPath := args[2]
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	lastIndexOfSlash := strings.LastIndex(dir, "/")
	packageName := dir
	if lastIndexOfSlash > 0 {
		packageName = dir[lastIndexOfSlash+1:]
	}
	fmt.Println(dir)
	fmt.Println(packageName)

	if err := GeneratePageEntityFile(dir, typeName, packageName, outputTypeName, dataPath); err != nil {
		panic(err)
	}
}
