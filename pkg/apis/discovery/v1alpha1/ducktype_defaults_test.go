/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1alpha1

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestDuckTypeDefaulting(t *testing.T) {
	tests := map[string]struct {
		in   *DuckType
		want *DuckType
	}{
		"empty": {
			in:   &DuckType{},
			want: &DuckType{},
		},
		"name set - lowercase": {
			in: &DuckType{
				Spec: DuckTypeSpec{
					Names: DuckTypeNames{
						Name: "thisduck",
					},
				}},
			want: &DuckType{
				Spec: DuckTypeSpec{
					Names: DuckTypeNames{
						Name:     "thisduck",
						Singular: "thisduck",
					},
				}},
		},
		"name set - camelcase": {
			in: &DuckType{
				Spec: DuckTypeSpec{
					Names: DuckTypeNames{
						Name: "ThisDuck",
					},
				}},
			want: &DuckType{
				Spec: DuckTypeSpec{
					Names: DuckTypeNames{
						Name:     "ThisDuck",
						Singular: "thisduck",
					},
				}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.in
			got.SetDefaults(context.Background())
			if !cmp.Equal(got, tc.want) {
				t.Errorf("SetDefaults (-want, +got) = %v",
					cmp.Diff(tc.want, got))
			}
		})
	}
}
