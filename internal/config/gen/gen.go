// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/config"
)

const license = `// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

`

var durationType = reflect.TypeOf(time.Duration(0))
var stringerType = reflect.TypeOf((*interface{ String() string })(nil)).Elem()
var stringersType = reflect.TypeOf((*interface{ Strings() []string })(nil)).Elem()
var flagSetType = reflect.TypeOf((*interface{ Set(string) error })(nil)).Elem()

func main() {
	var out string

	// Load runtime config flags
	flag.StringVar(&out, "out", "", "Generated file output path")
	flag.Parse()

	// Open output file path
	output, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}

	var confStruct ConfigStruct
	confType := reflect.TypeOf(config.Configuration{})

	// Parse our config type for
	// usable configuration fields.
	confStruct.Type = confType
	confStruct.Load(confType)

	fprintf(output, "// THIS IS A GENERATED FILE, DO NOT EDIT BY HAND\n")
	fprintf(output, license)
	fprintf(output, "package config\n\n")
	fprintf(output, "import (\n")
	fprintf(output, "\t\"fmt\"\n")
	fprintf(output, "\t\"time\"\n\n")
	fprintf(output, "\t\"codeberg.org/gruf/go-bytesize\"\n")
	fprintf(output, "\t\"code.superseriousbusiness.org/gotosocial/internal/language\"\n")
	fprintf(output, "\t\"github.com/spf13/pflag\"\n")
	fprintf(output, "\t\"github.com/spf13/cast\"\n")
	fprintf(output, ")\n")
	fprintf(output, "\n")
	generateFlagConsts(output, &confStruct)
	generateFlagRegistering(output, &confStruct)
	generateMapMarshaler(output, &confStruct)
	generateMapUnmarshaler(output, &confStruct)
	generateGetSetters(output, &confStruct)
	must(output.Close())
	must(exec.Command("goimports", "-w", out).Run())
}

type ConfigStruct struct {
	// The config struct this
	// struct *may* be a member of,
	// this provides prefixes.
	Struct *ConfigStruct

	// ...
	FlagName string

	// The name of struct's field
	// in its containing struct.
	FieldName string

	// The underlying Go type
	// of the config struct.
	Type reflect.Type

	// ...
	Fields []ConfigField

	// ...
	Structs []ConfigStruct
}

type ConfigField struct {
	// The config struct this
	// field is a member of,
	// this provides prefixes.
	Struct *ConfigStruct

	// The config flag
	// name of the field.
	FlagName string

	// The name of this field
	// in containing struct.
	FieldName string

	// Usage string.
	Usage string

	// The underlying Go type
	// of the config field.
	Type reflect.Type

	// Whether to generate
	// CLI flag registering.
	RegisterCLI bool

	// i.e. is this a deprecated field we don't
	// want being used, point to this field instead.
	DeprecatedBy string
}

// Flag ...
func (f *ConfigField) Flag() string {
	flag := f.FlagName
	parent := f.Struct
	for parent != nil && parent.FlagName != "" {
		flag = parent.FlagName + "-" + flag
		parent = parent.Struct
	}
	return flag
}

// Path ...
func (f *ConfigField) Path() string {
	path := f.FieldName
	parent := f.Struct
	for parent != nil && parent.FieldName != "" {
		path = parent.FieldName + "." + path
		parent = parent.Struct
	}
	return path
}

// TypeName ...
func (f *ConfigField) TypeName() string {
	return strings.TrimPrefix(f.Type.String(), "config.")
}

// TypeName ...
func (s *ConfigStruct) TypeName() string {
	return strings.TrimPrefix(s.Type.String(), "config.")
}

// Load ...
func (s *ConfigStruct) Load(t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get field's tagged name.
		name := field.Tag.Get("name")
		if name == "-" {
			continue
		}

		// Get field's tagged usage.
		usage := field.Tag.Get("usage")

		// Check if another deprecates this one.
		depBy := field.Tag.Get("deprecated-by")

		if ft := field.Type; ft.Kind() == reflect.Struct &&
			depBy == "" && usage == "" {

			// This is a nested struct, load nested fields.
			s.Structs = append(s.Structs, ConfigStruct{})
			nested := &(s.Structs[len(s.Structs)-1])
			nested.Struct = s
			nested.Type = ft
			nested.FlagName = name // can be empty
			nested.FieldName = field.Name
			nested.Load(ft)
			continue
		}

		// We have a config value,
		// ensure is valid field.
		if name == "" {
			panic("field with empty name")
		}

		// Append ConfigField to parent struct.
		s.Fields = append(s.Fields, ConfigField{
			Struct:       s,
			FlagName:     name,
			FieldName:    field.Name,
			Usage:        usage,
			Type:         field.Type,
			RegisterCLI:  field.Tag.Get("nocli") != "yes",
			DeprecatedBy: depBy,
		})
	}
}

// RangeFields ...
func (s *ConfigStruct) RangeFields(yield func(*ConfigField) bool) {
	if yield == nil {
		panic("nil func")
	}
	for i := range s.Structs {
		(&s.Structs[i]).RangeFields(yield)
	}
	for i := range s.Fields {
		if !yield(&s.Fields[i]) {
			return
		}
	}
}

func generateFlagConsts(out io.Writer, config *ConfigStruct) {
	fprintf(out, "const (\n")
	for field := range config.RangeFields {
		name := strings.ReplaceAll(field.Path(), ".", "")
		fprintf(out, "\t%sFlag = \"%s\"\n", name, field.Flag())
	}
	fprintf(out, ")\n\n")
}

func generateFlagRegistering(out io.Writer, config *ConfigStruct) {
	// Before starting to define our
	// own function definition, call
	// generate() for our sub-structs.
	for i := range config.Structs {
		generateFlagRegistering(out, &config.Structs[i])
	}

	// Generate RegisterFlags() function,
	// for the top-level Configuration{}
	// type it takes no flag prefix, but
	// all other configuration types do.
	if config.FieldName == "" {
		fprintf(out, "func (cfg *%s) RegisterFlags(flags *pflag.FlagSet) {\n", config.TypeName())
		for _, strct := range config.Structs {
			fprintf(out, "\tcfg.%s.RegisterFlags(\"%s\", flags)\n", strct.FieldName, strct.FlagName)
		}
	} else {
		fprintf(out, "func (cfg *%s) RegisterFlags(prefix string, flags *pflag.FlagSet) {\n", config.TypeName())
		for _, strct := range config.Structs {
			fprintf(out, "\tcfg.%s.RegisterFlags(joinFlag(prefix, \"%s\"), flags)\n", strct.FieldName, strct.FlagName)
		}
	}

	for _, field := range config.Fields {
		if !field.RegisterCLI {
			// Skip registering flags
			// unpermitted in env / cli.
			continue
		}

		if field.DeprecatedBy != "" {
			// Skip registering
			// deprecated flags.
			continue
		}

		var flag string

		// For top level Configuration{} type
		// without a prefix, each field flag name
		// is as-is (quoted for fmt string below).
		// For others, they need to add prefix.
		if config.FieldName == "" {
			flag = "\"" + field.FlagName + "\""
		} else {
			flag = "joinFlag(prefix, \"" + field.FlagName + "\")"
		}

		// "path" to the struct field
		// member is simple its name.
		path := field.FieldName

		// Check for easy cases of just regular primitive types.
		if field.Type.Kind().String() == field.Type.String() {
			typeName := field.Type.String()
			typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
			fprintf(out, "\tflags.%s(%s, cfg.%s, \"%s\")\n", typeName, flag, path, field.Usage)
			continue
		}

		// Check for easy cases of just
		// regular primitive slice types.
		if field.Type.Kind() == reflect.Slice {
			elem := field.Type.Elem()
			if elem.Kind().String() == elem.String() {
				typeName := elem.String()
				typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
				fprintf(out, "\tflags.%sSlice(%s, cfg.%s, \"%s\")\n", typeName, flag, path, field.Usage)
				continue
			}
		}

		// Durations should get set directly
		// as their types as viper knows how
		// to deal with this type directly.
		if field.Type == durationType {
			fprintf(out, "\tflags.Duration(%s, cfg.%s, \"%s\")\n", flag, path, field.Usage)
			continue
		}

		if field.Type.Kind() == reflect.Slice {
			// Check if the field supports Stringers{}.
			if field.Type.Implements(stringersType) {
				fprintf(out, "\tflags.StringSlice(%s, cfg.%s.Strings(), \"%s\")\n", flag, path, field.Usage)
				continue
			}

			// Or the pointer type of the field value supports Stringers{}.
			if ptr := reflect.PointerTo(field.Type); ptr.Implements(stringersType) {
				fprintf(out, "\tflags.StringSlice(%s, cfg.%s.Strings(), \"%s\")\n", flag, path, field.Usage)
				continue
			}

			fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Flag(), stringersType)
		} else {
			// Check if the field supports Stringer{}.
			if field.Type.Implements(stringerType) {
				fprintf(out, "\tflags.String(%s, cfg.%s.String(), \"%s\")\n", flag, path, field.Usage)
				continue
			}

			// Or the pointer type of the field value supports Stringer{}.
			if ptr := reflect.PointerTo(field.Type); ptr.Implements(stringerType) {
				fprintf(out, "\tflags.String(%s, cfg.%s.String(), \"%s\")\n", flag, path, field.Usage)
				continue
			}

			fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Flag(), stringerType)
		}
	}

	fprintf(out, "}\n\n")
}

func generateMapMarshaler(out io.Writer, config *ConfigStruct) {
	// Before starting to define our
	// own function definition, call
	// generate() for our sub-structs.
	for i := range config.Structs {
		generateMapMarshaler(out, &config.Structs[i])
	}

	if config.FieldName == "" {
		var count int
		for range config.RangeFields {
			count++
		}

		// The top-level Configuration{} type gets a special MarshalMap() function,
		// different from the below MarshalIntoMap() functions in that it returns its value.
		fprintf(out, "func (cfg *%s) MarshalMap() map[string]any {\n", config.TypeName())
		fprintf(out, "\tcfgmap := make(map[string]any, %d)\n", count)
		fprintf(out, "\tcfg.MarshalIntoMap(cfgmap)\n")
		fprintf(out, "\treturn cfgmap\n")
		fprintf(out, "}\n\n")
	}

	// Generate MarshalIntoMap() function,
	// for the top-level Configuration{}
	// type it takes no flag prefix, but
	// all other configuration types do.
	if config.FieldName == "" {
		fprintf(out, "func (cfg *%s) MarshalIntoMap(cfgmap map[string]any) {\n", config.TypeName())
		for _, strct := range config.Structs {
			fprintf(out, "\tcfg.%s.MarshalIntoMap(\"%s\", cfgmap)\n", strct.FieldName, strct.FlagName)
		}
	} else {
		fprintf(out, "func (cfg *%s) MarshalIntoMap(prefix string, cfgmap map[string]any) {\n", config.TypeName())
		for _, strct := range config.Structs {
			fprintf(out, "\tcfg.%s.MarshalIntoMap(joinFlag(prefix, \"%s\"), cfgmap)\n", strct.FieldName, strct.FlagName)
		}
	}

	for _, field := range config.Fields {
		// Deprecated fields don't need
		// including in marshaled map.
		if field.DeprecatedBy != "" {
			continue
		}

		var flag string

		// For top level Configuration{} type
		// without a prefix, each field flag name
		// is as-is (quoted for fmt string below).
		// For others, they need to add prefix.
		if config.FieldName == "" {
			flag = "\"" + field.FlagName + "\""
		} else {
			flag = "joinFlag(prefix, \"" + field.FlagName + "\")"
		}

		// "path" to the struct field
		// member is simple its name.
		path := field.FieldName

		// Check for easy cases of just regular primitive types.
		if field.Type.Kind().String() == field.Type.String() {
			fprintf(out, "\tcfgmap[%s] = cfg.%s\n", flag, path)
			continue
		}

		// Check for easy cases of just
		// regular primitive slice types.
		if field.Type.Kind() == reflect.Slice {
			elem := field.Type.Elem()
			if elem.Kind().String() == elem.String() {
				fprintf(out, "\tcfgmap[%s] = cfg.%s\n", flag, path)
				continue
			}
		}

		// Durations should get set directly
		// as their types as viper knows how
		// to deal with this type directly.
		if field.Type == durationType {
			fprintf(out, "\tcfgmap[%s] = cfg.%s\n", flag, path)
			continue
		}

		if field.Type.Kind() == reflect.Slice {
			// Either the field must support Stringers{}.
			if field.Type.Implements(stringersType) {
				fprintf(out, "\tcfgmap[%s] = cfg.%s.Strings()\n", flag, path)
				continue
			}

			// Or the pointer type of the field value must support Stringers{}.
			if ptr := reflect.PointerTo(field.Type); ptr.Implements(stringersType) {
				fprintf(out, "\tcfgmap[%s] = cfg.%s.Strings()\n", flag, path)
				continue
			}

			fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Flag(), stringersType)
		} else {
			// Either the field must support Stringer{}.
			if field.Type.Implements(stringerType) {
				fprintf(out, "\tcfgmap[%s] = cfg.%s.String()\n", flag, path)
				continue
			}

			// Or the pointer type of the field value must support Stringer{}.
			if ptr := reflect.PointerTo(field.Type); ptr.Implements(stringerType) {
				fprintf(out, "\tcfgmap[%s] = cfg.%s.String()\n", flag, path)
				continue
			}

			fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Flag(), stringerType)
		}
	}
	fprintf(out, "}\n\n")
}

func generateMapUnmarshaler(out io.Writer, config *ConfigStruct) {
	// Before starting to define our
	// own function definition, call
	// generate() for our sub-structs.
	for i := range config.Structs {
		generateMapUnmarshaler(out, &config.Structs[i])
	}

	// Generate UnmarshalMap() function,
	// for the top-level Configuration{}
	// type it takes no flag prefix, but
	// all other configuration types do.
	if config.FieldName == "" {
		fprintf(out, "func (cfg *%s) UnmarshalMap(cfgmap map[string]any) error {\n", config.TypeName())
		for _, strct := range config.Structs {
			fprintf(out, "\tif err := cfg.%s.UnmarshalMap(\"%s\", cfgmap); err != nil {\n", strct.FieldName, strct.FlagName)
			fprintf(out, "\t\treturn err\n")
			fprintf(out, "\t}\n\n")
		}
	} else {
		fprintf(out, "func (cfg *%s) UnmarshalMap(prefix string, cfgmap map[string]any) error {\n", config.TypeName())
		for _, strct := range config.Structs {
			fprintf(out, "\tif err := cfg.%s.UnmarshalMap(joinFlag(prefix, \"%s\"), cfgmap); err != nil {\n", strct.FieldName, strct.FlagName)
			fprintf(out, "\t\treturn err\n")
			fprintf(out, "\t}\n\n")
		}
	}

	for i := range config.Fields {
		field := &(config.Fields[i])

		var flag string

		// For top level Configuration{} type
		// without a prefix, each field flag name
		// is as-is (quoted for fmt string below).
		// For others, they need to add prefix.
		if config.FieldName == "" {
			flag = "\"" + field.FlagName + "\""
		} else {
			flag = "joinFlag(prefix, \"" + field.FlagName + "\")"
		}

		// "path" to the struct field
		// member is simple its name.
		path := field.FieldName

		// Check for case of deprecated.
		if field.DeprecatedBy != "" {
			generateUnmarshalerDeprecated(out, flag, path, field)
			continue
		}

		// Check for easy cases of just regular primitive types.
		if field.Type.Kind().String() == field.Type.String() {
			generateUnmarshalerPrimitive(out, flag, path, field)
			continue
		}

		// Check for easy cases of just
		// regular primitive slice types.
		if field.Type.Kind() == reflect.Slice {
			elem := field.Type.Elem()
			if elem.Kind().String() == elem.String() {
				generateUnmarshalerPrimitive(out, flag, path, field)
				continue
			}
		}

		// Durations should get set directly
		// as their types as viper knows how
		// to deal with this type directly.
		if field.Type == durationType {
			generateUnmarshalerPrimitive(out, flag, path, field)
			continue
		}

		// Either the field must support flag.Value{}.
		if field.Type.Implements(flagSetType) {
			generateUnmarshalerFlagType(out, flag, path, field)
			continue
		}

		// Or the pointer type of the field value must support flag.Value{}.
		if ptr := reflect.PointerTo(field.Type); ptr.Implements(flagSetType) {
			generateUnmarshalerFlagType(out, flag, path, field)
			continue
		}

		fprintf(os.Stderr, "field %s doesn't implement %s!\n", field.Flag(), flagSetType)
	}
	fprintf(out, "\treturn nil\n")
	fprintf(out, "}\n\n")
}

func generateUnmarshalerDeprecated(out io.Writer, flag, path string, field *ConfigField) {
	fprintf(out, "\tif ival, ok := cfgmap[%s]; ok && ival != \"\" {\n", flag)
	fprintf(out, "\t\treturn fmt.Errorf(\"value received for deprecated field '%%s', please use '%%s' instead\", %s, %q)\n", flag, field.DeprecatedBy)
	fprintf(out, "\t}\n")
	fprintf(out, "\n")
}

func generateUnmarshalerPrimitive(out io.Writer, flag, path string, field *ConfigField) {
	fprintf(out, "\tif ival, ok := cfgmap[%s]; ok {\n", flag)
	if field.Type.Kind() == reflect.Slice {
		elem := field.Type.Elem()
		typeName := elem.String()
		if i := strings.IndexRune(typeName, '.'); i >= 0 {
			typeName = typeName[i+1:]
		}
		typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
		fprintf(out, "\t\tvar err error\n")
		// note we specifically handle slice types ourselves to split by comma
		fprintf(out, "\t\tcfg.%s, err = to%sSlice(ival)\n", path, typeName)
		fprintf(out, "\t\tif err != nil {\n")
		fprintf(out, "\t\t\treturn fmt.Errorf(\"error casting %%#v -> %%s for '%%s': %%w\", ival, \"[]%s\", %s, err)\n", elem.String(), flag)
		fprintf(out, "\t\t}\n")
	} else {
		typeName := field.Type.String()
		if i := strings.IndexRune(typeName, '.'); i >= 0 {
			typeName = typeName[i+1:]
		}
		typeName = strings.ToUpper(typeName[:1]) + typeName[1:]
		fprintf(out, "\t\tvar err error\n")
		fprintf(out, "\t\tcfg.%s, err = cast.To%sE(ival)\n", path, typeName)
		fprintf(out, "\t\tif err != nil {\n")
		fprintf(out, "\t\t\treturn fmt.Errorf(\"error casting %%#v -> %%s for '%%s': %%w\", ival, \"%s\", %s, err)\n", field.TypeName(), flag)
		fprintf(out, "\t\t}\n")
	}
	fprintf(out, "\t}\n")
	fprintf(out, "\n")
}

func generateUnmarshalerFlagType(out io.Writer, flag, path string, field *ConfigField) {
	fprintf(out, "\t\tif ival, ok := cfgmap[%s]; ok {\n", flag)
	if field.Type.Kind() == reflect.Slice {
		// same as above re: slice types and splitting on comma
		fprintf(out, "\t\tt, err := toStringSlice(ival)\n")
		fprintf(out, "\t\tif err != nil {\n")
		fprintf(out, "\t\t\treturn fmt.Errorf(\"error casting %%#v -> []string for '%%s': %%w\", ival, %s, err)\n", flag)
		fprintf(out, "\t\t}\n")
		fprintf(out, "\t\tcfg.%s = %s{}\n", path, strings.TrimPrefix(field.Type.String(), "config."))
		fprintf(out, "\t\tfor _, in := range t {\n")
		fprintf(out, "\t\t\tif err := cfg.%s.Set(in); err != nil {\n", path)
		fprintf(out, "\t\t\t\treturn fmt.Errorf(\"error parsing %%#v for '%%s': %%w\", ival, %s, err)\n", flag)
		fprintf(out, "\t\t\t}\n")
		fprintf(out, "\t\t}\n")
	} else {
		fprintf(out, "\t\tt, err := cast.ToStringE(ival)\n")
		fprintf(out, "\t\tif err != nil {\n")
		fprintf(out, "\t\t\treturn fmt.Errorf(\"error casting %%#v -> string for '%%s': %%w\", ival, %s, err)\n", flag)
		fprintf(out, "\t\t}\n")
		fprintf(out, "\t\tcfg.%s = %s\n", path, strings.TrimPrefix(zeroValueStr(field.Type), "config."))
		fprintf(out, "\t\tif err := cfg.%s.Set(t); err != nil {\n", path)
		fprintf(out, "\t\t\treturn fmt.Errorf(\"error parsing %%#v for '%%s': %%w\", ival, %s, err)\n", flag)
		fprintf(out, "\t\t}\n")
	}
	fprintf(out, "\t}\n")
	fprintf(out, "\n")
}

func zeroValueStr(t reflect.Type) string {
	return fmt.Sprintf("%#v", reflect.New(t).Elem().Interface())
}

func generateGetSetters(out io.Writer, config *ConfigStruct) {
	for field := range config.RangeFields {
		// Get name from struct path, without periods.
		name := strings.ReplaceAll(field.Path(), ".", "")

		// Type without "config." prefix.
		fieldType := field.TypeName()

		// ConfigState structure helper methods
		fprintf(out, "// Get%s safely fetches the Configuration value for state's '%s' field\n", name, field.Path())
		fprintf(out, "func (st *ConfigState) Get%s() (v %s) {\n", name, fieldType)
		fprintf(out, "\treturn st.config.%s\n", field.Path())
		fprintf(out, "}\n\n")
		fprintf(out, "// Set%s safely sets the Configuration value for state's '%s' field\n", name, field.Path())
		fprintf(out, "func (st *ConfigState) Set%s(v %s) {\n", name, fieldType)
		fprintf(out, "\tst.config.%s = v\n", field.Path())
		fprintf(out, "\tst.reloadToViper()\n")
		fprintf(out, "}\n\n")

		// Global ConfigState helper methods
		fprintf(out, "// Get%s safely fetches the value for global configuration '%s' field\n", name, field.Path())
		fprintf(out, "func Get%[1]s() %[2]s { return global.Get%[1]s() }\n\n", name, fieldType)
		fprintf(out, "// Set%s safely sets the value for global configuration '%s' field\n", name, field.Path())
		fprintf(out, "func Set%[1]s(v %[2]s) { global.Set%[1]s(v) }\n\n", name, fieldType)
	}

	// Separate out the config fields
	// to get only the `mem-ratio` ones.
	var memFields []*ConfigField
	for field := range config.RangeFields {
		if strings.Contains(field.FieldName, "MemRatio") {
			memFields = append(memFields, field)
		}
	}

	fprintf(out, "// GetTotalOfMemRatios safely fetches the combined value for all the state's mem ratio fields\n")
	fprintf(out, "func (st *ConfigState) GetTotalOfMemRatios() (total float64) {\n")
	for _, field := range memFields {
		fprintf(out, "\ttotal += st.config.%s\n", field.Path())
	}
	fprintf(out, "\treturn\n")
	fprintf(out, "}\n\n")

	fprintf(out, "// GetTotalOfMemRatios safely fetches the combined value for all the global state's mem ratio fields\n")
	fprintf(out, "func GetTotalOfMemRatios() (total float64) { return global.GetTotalOfMemRatios() }\n\n")
}

func fprintf(out io.Writer, format string, args ...any) {
	_, err := fmt.Fprintf(out, format, args...)
	must(err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}

}
