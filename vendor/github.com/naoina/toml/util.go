// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package toml

import (
	"fmt"
	"reflect"
	"strings"
)

const fieldTagName = "toml"

// fieldCache maps normalized field names to their position in a struct.
type fieldCache struct {
	named map[string]fieldInfo // fields with an explicit name in tag
	auto  map[string]fieldInfo // fields with auto-assigned normalized names
}

type fieldInfo struct {
	index   []int
	name    string
	ignored bool
}

func makeFieldCache(cfg *Config, rt reflect.Type) fieldCache {
	named, auto := make(map[string]fieldInfo), make(map[string]fieldInfo)
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		// skip unexported fields
		if ft.PkgPath != "" && !ft.Anonymous {
			continue
		}
		col, _ := extractTag(ft.Tag.Get(fieldTagName))
		info := fieldInfo{index: ft.Index, name: ft.Name, ignored: col == "-"}
		if col == "" || col == "-" {
			auto[cfg.NormFieldName(rt, ft.Name)] = info
		} else {
			named[col] = info
		}
	}
	return fieldCache{named, auto}
}

func (fc fieldCache) findField(cfg *Config, rv reflect.Value, name string) (reflect.Value, string, error) {
	info, found := fc.named[name]
	if !found {
		info, found = fc.auto[cfg.NormFieldName(rv.Type(), name)]
	}
	if !found {
		if cfg.MissingField == nil {
			return reflect.Value{}, "", fmt.Errorf("field corresponding to `%s' is not defined in %v", name, rv.Type())
		} else {
			return reflect.Value{}, "", cfg.MissingField(rv.Type(), name)
		}
	} else if info.ignored {
		return reflect.Value{}, "", fmt.Errorf("field corresponding to `%s' in %v cannot be set through TOML", name, rv.Type())
	}
	return rv.FieldByIndex(info.index), info.name, nil
}

func extractTag(tag string) (col, rest string) {
	tags := strings.SplitN(tag, ",", 2)
	if len(tags) == 2 {
		return strings.TrimSpace(tags[0]), strings.TrimSpace(tags[1])
	}
	return strings.TrimSpace(tags[0]), ""
}
