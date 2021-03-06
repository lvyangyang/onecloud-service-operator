// Copyright 2020 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/util/intstr"
)

// +kubebuilder:object:generate=false
type IValue interface {
	IsZero() bool
	Interface() interface{}
}

// +kubebuilder:object:generate=false
type IStore interface {
	Value(ctx context.Context) (IValue, error)
}

type StringStore struct {
	// +optional
	Direct string `json:"direct,omitempty"`
	// +optional
	Indirect *ObjectFieldReference `json:"indirect,omitempty"`
}

type String string

func (sv String) IsZero() bool {
	return sv == ""
}

func (sv String) String() string {
	return string(sv)
}

func (sv String) Interface() interface{} {
	return sv.String()
}

func (st *StringStore) Value(ctx context.Context) (IValue, error) {
	if len(st.Direct) > 0 {
		return String(st.Direct), nil
	}
	in, err := st.Indirect.Value(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, nil
	}
	s, ok := in.(string)
	if !ok {
		ts := reflect.TypeOf(in).String()
		return nil, fmt.Errorf("Type of ObjectFieldReference' Value in not 'string' but '%s'", ts)
	}
	return String(s), nil
}

type IntOrStringStore struct {
	// +optional
	Direct *IntOrString `json:"direct,omitempty"`
	// +optional
	Indirect *ObjectFieldReference `json:"indirect,omitempty"`
}

type IntOrString struct {
	intstr.IntOrString `json:",inline"`
}

func (isv *IntOrString) String() (string, bool) {
	if isv.Type == intstr.String {
		return isv.StrVal, true
	}
	return "", false
}

func (isv *IntOrString) Int() (int32, bool) {
	if isv.Type == intstr.Int {
		return isv.IntVal, true
	}
	return 0, false
}

func (isv *IntOrString) IsZero() bool {
	if isv == nil {
		return true
	}
	if s, ok := isv.String(); ok {
		return s == ""
	}
	if i, ok := isv.Int(); ok {
		return i == 0
	}
	return true
}

func (isv *IntOrString) Interface() interface{} {
	if s, ok := isv.String(); ok {
		return s
	}
	if i, ok := isv.Int(); ok {
		return i
	}
	return isv
}

func (ist *IntOrStringStore) Value(ctx context.Context) (IValue, error) {
	if ist.Direct != nil {
		return ist.Direct, nil
	}
	in, err := ist.Indirect.Value(ctx)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, err
	}

	var is intstr.IntOrString
	switch v := in.(type) {
	case string:
		is = intstr.FromString(v)
	case int:
		is = intstr.FromInt(v)
	case int32:
		is = intstr.FromInt(int(v))
	case int64:
		is = intstr.FromInt(int(v))
	default:
		ts := reflect.TypeOf(in).String()
		return nil, fmt.Errorf("Type of ObjectFieldReference' Value in not 'string' but '%s'", ts)
	}
	return &IntOrString{is}, nil
}
