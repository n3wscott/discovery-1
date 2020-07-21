/*
Copyright 2020 The Knative Authors.

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
	"knative.dev/pkg/apis"
	"strings"
)

// SetDefaults implements apis.Defaultable
func (dt *DuckType) SetDefaults(ctx context.Context) {
	ctx = apis.WithinParent(ctx, dt.ObjectMeta)
	dt.Spec.SetDefaults(apis.WithinSpec(ctx))
}

// SetDefaults implements apis.Defaultable
func (dts *DuckTypeSpec) SetDefaults(ctx context.Context) {
	// names.singular defaults to lowercase names.name if not set.
	if dts.Names.Singular == "" {
		dts.Names.Singular = strings.ToLower(dts.Names.Name)
	}
}
