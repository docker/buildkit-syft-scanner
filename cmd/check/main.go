// Copyright 2023 buildkit-syft-scanner authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// check is a simple script that ensures a target JSON file matches a specified
// schema.
//
// The schema is another JSON file that should match the desired format
// structure of the target - check will ensure that the schema is a subset of
// the target.
//
// check also provides simple variables and comparisons to easily evaluate
// relationships in the target SPDX files. A property set to "=" will assign
// the variable value to the corresponding property from the target, while a
// property set to "==" will check the previously assigned variable to the
// corresponding property from the target.
//
// See the ./examples/*/checks/ directories for usage examples.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

func main() {
	schemaFilename := os.Args[1]
	targetFilename := os.Args[2]

	var schema, target map[string]interface{}
	if dt, err := os.ReadFile(schemaFilename); err != nil {
		panic(err)
	} else {
		if err := json.Unmarshal(dt, &schema); err != nil {
			panic(err)
		}
	}
	if dt, err := os.ReadFile(targetFilename); err != nil {
		panic(err)
	} else {
		if err := json.Unmarshal(dt, &target); err != nil {
			panic(err)
		}
	}

	vars := make(map[string]interface{})
	err := check(schema, target, vars, true, "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = check(schema, target, vars, false, "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func check(schema interface{}, target interface{}, vars map[string]interface{}, assign bool, previous ...string) error {
	schemaType := reflect.TypeOf(schema)
	targetType := reflect.TypeOf(target)

	if schemaType.Kind() == reflect.String {
		if strings.HasPrefix(schema.(string), "==") {
			if !assign {
				key := schema.(string)[2:]
				if _, ok := vars[key]; !ok {
					return fmt.Errorf("variable %s not found", key)
				}
				if target != vars[key] {
					return fmt.Errorf("variable mismatch on %s, expected %v, got %v", strings.Join(previous, "."), vars[key], target)
				}
			}
			return nil
		}
		if strings.HasPrefix(schema.(string), "=") {
			key := schema.(string)[1:]
			if assign {
				vars[key] = target
			}
			return nil
		}
	}

	if schemaType.Kind() != targetType.Kind() {
		return fmt.Errorf("type mismatch on %s, expected %s, got %s", strings.Join(previous, "."), schemaType.Kind(), targetType.Kind())
	}
	switch schemaType.Kind() {
	case reflect.Pointer:
		return check(reflect.ValueOf(schema).Elem().Interface(), reflect.ValueOf(target).Elem().Interface(), vars, assign, previous...)
	case reflect.Map:
		return checkMap(schema.(map[string]interface{}), target.(map[string]interface{}), vars, assign, previous...)
	case reflect.Array, reflect.Slice:
		return checkSlice(schema.([]interface{}), target.([]interface{}), vars, assign, previous...)
	default:
		if !reflect.DeepEqual(schema, target) {
			return fmt.Errorf("value mismatch on %s, expected %v, got %v", strings.Join(previous, "."), schema, target)
		}
		return nil
	}
}

func checkMap(schema map[string]interface{}, target map[string]interface{}, vars map[string]interface{}, assign bool, previous ...string) error {
	for k, v := range schema {
		v2, ok := target[k]
		if !ok {
			return fmt.Errorf("map mismatch on %s, expected %v", strings.Join(previous, ".")+"."+k, schema)
		}
		if err := check(v, v2, vars, assign, append(previous, k)...); err != nil {
			return err
		}
	}
	return nil
}

func checkSlice(schema []interface{}, target []interface{}, vars map[string]interface{}, assign bool, previous ...string) error {
	if len(schema) > len(target) {
		return fmt.Errorf("length mismatch on %s, expected at least %d, got %d", strings.Join(previous, "."), len(schema), len(target))
	}
	for i, v := range schema {
		found := false
		for _, v2 := range target {
			if err := check(v, v2, vars, assign, append(previous, fmt.Sprintf("[%d]", i))...); err == nil {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("slice mismatch on %s, expected %v", strings.Join(previous, ".")+fmt.Sprintf("[%d]", i), schema)
		}
	}
	return nil
}
