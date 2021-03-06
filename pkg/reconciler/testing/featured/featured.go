/*
Copyright 2021 The Knative Authors

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

package featured

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/cucumber/messages-go/v10"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
	. "knative.dev/discovery/pkg/reconciler/testing/v1alpha1"
	pkgtest "knative.dev/pkg/reconciler/testing"
	"knative.dev/reconciler-test/pkg/manifest"
)

var opt = godog.Options{
	Output: colors.Colored(os.Stdout),
}

// Step is a wrapper for a call to godog.ScenarioContext.Step()
type Step struct {
	Expr         string
	StepFuncCtor func(rt *ReconcilerTest) interface{}
}

func Run(m *testing.M) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if len(flag.Args()) > 0 {
		opt.Paths = flag.Args()
	} else {
		opt.Paths = []string{
			filepath.Join(callerPath(), "testdata/features/"),
		}
	}

	format := "progress"
	for _, arg := range os.Args[1:] {
		if arg == "-test.v=true" { // go test transforms -v option
			format = "pretty"
			break
		}
	}

	opt.Format = format

	os.Exit(m.Run())
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {})
}

func TestReconcileKind(t *testing.T, kind string, factory pkgtest.Factory, steps ...Step) {
	status := godog.TestSuite{
		Name:                 kind,
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			ReconcileKindFeatureContext(t, s, kind, factory, steps...)
		},
		Options: &opt,
	}.Run()

	if status != 0 {
		t.Fail()
	}
}

func ReconcileKindFeatureContext(t *testing.T, s *godog.ScenarioContext, kind string, factory pkgtest.Factory, steps ...Step) {
	ctx := context.Background()

	rt := &ReconcilerTest{
		T: t,
		Row: pkgtest.TableRow{
			Name: kind,
			Ctx:  ctx,
			//OtherTestData:           nil,
			//Objects:                 nil,
			//Key:                     "",
			//WantErr:                 false,
			//WantCreates:             nil,
			//WantUpdates:             nil,
			//WantStatusUpdates:       nil,
			//WantDeletes:             nil,
			//WantDeleteCollections:   nil,
			//WantPatches:             nil,
			//WantEvents:              nil,
			//WithReactors:            nil,
			//SkipNamespaceValidation: false,
			//PostConditions:          nil,
			//Reconciler:              nil,
		}}

	s.Step(`^the following objects:$`, rt.theFollowingObjects)
	s.Step(`^the following objects \(from file\):$`, rt.theFollowingObjectFiles)
	s.Step(fmt.Sprintf(`^a %s reconciler$`, kind), func() error {
		rt.Factory = factory
		return nil
	})
	s.Step(`^reconciling "([^"]*)"$`, rt.reconcilingKey)
	s.Step(`^expect nothing$`, rt.expectNothing)
	s.Step(`^expect status updates:$`, rt.expectStatusUpdates)
	s.Step(`^expect status updates \(from file\):$`, rt.expectStatusUpdateFiles)
	s.Step(`^expect Kubernetes Events:$`, rt.expectKubernetesEvents)

	// Inject custom steps.
	for _, step := range steps {
		s.Step(step.Expr, step.StepFuncCtor(rt))
	}

	s.AfterScenario(func(pickle *messages.Pickle, err error) {
		originObjects := make([]runtime.Object, 0, len(rt.Row.Objects))
		for _, obj := range rt.Row.Objects {
			originObjects = append(originObjects, obj.DeepCopyObject())
		}

		if rt.Factory == nil {
			rt.T.Fatalf("factory is not set for test %s", rt.Row.Name)
		}

		// Reconcile.
		rt.Row.Name = pickle.Name
		rt.T.Run(pickle.Name, func(t *testing.T) {
			t.Helper()
			rt.Row.Test(t, rt.Factory)
		})

		// Validate cached objects do not get soiled after controller loops
		if diff := cmp.Diff(originObjects, rt.Row.Objects, safeDeployDiff, cmpopts.EquateEmpty()); diff != "" {
			rt.T.Errorf("Unexpected objects in test %s (-want, +got): %v", rt.Row.Name, diff)
		}
	})

	_ = rt
}

type ReconcilerTest struct {
	T       *testing.T
	Row     pkgtest.TableRow
	Factory pkgtest.Factory
}

func (rt *ReconcilerTest) theFollowingObjectFiles(y *messages.PickleStepArgument_PickleTable) error {
	keys := make([]string, 0)
	for row, v := range y.Rows {
		var file string
		config := make(map[string]interface{})

		for i, c := range v.Cells {
			if row == 0 {
				if i == 0 {
					continue
				}
				keys = append(keys, c.Value)
				continue
			}

			switch i {
			case 0:
				file = "testdata/" + c.Value
			default:
				config[keys[i-1]] = c.Value
			}
		}

		// Leverage ParseTemplates to parse the template files.
		// TODO: move this code around to let it be reusable
		// in a more clear way.
		if file != "" {
			files, err := manifest.ParseTemplates(file, nil, config)
			if err != nil {
				return err
			}

			list, err := ioutil.ReadDir(files)
			if err != nil {
				return err
			}
			// len zero would be an invalid or missing file.
			if len(list) == 0 {
				return fmt.Errorf("expected to read a yaml file from %q but found none", file)
			}
			for _, f := range list {
				name := path.Join(files, f.Name())
				if !f.IsDir() {
					ff, err := os.Open(name)
					if err != nil {
						return err
					}
					objs, err := ParseYAML(bufio.NewReader(ff))
					if err != nil {
						return err
					}

					rt.addObjects(objs)
				}
			}
		}
	}
	return nil
}

func (rt *ReconcilerTest) theFollowingObjects(y *messages.PickleStepArgument_PickleDocString) error {
	objs, err := ParseYAML(strings.NewReader(y.Content))
	if err != nil {
		return err
	}
	rt.addObjects(objs)

	return nil
}

func (rt *ReconcilerTest) addObjects(objs []unstructured.Unstructured) {
	if rt.Row.Objects == nil {
		rt.Row.Objects = make([]runtime.Object, 0, len(objs))
	}

	for _, obj := range objs {
		rt.Row.Objects = append(rt.Row.Objects, obj.DeepCopyObject())
	}

	rt.Row.Objects = ToKnownObjects(rt.Row.Objects)
}

func (rt *ReconcilerTest) reconcilingKey(key string) error {
	// Set the reconciler key
	rt.Row.Key = key

	return nil
}

func (rt *ReconcilerTest) expectNothing() error {
	return nil
}

func (rt *ReconcilerTest) expectStatusUpdates(y *messages.PickleStepArgument_PickleDocString) error {
	objs, err := ParseYAML(strings.NewReader(y.Content))
	if err != nil {
		return err
	}

	rt.addWantStatusUpdates(objs)
	return nil
}

func (rt *ReconcilerTest) expectStatusUpdateFiles(y *messages.PickleStepArgument_PickleTable) error {
	keys := make([]string, 0)
	for row, v := range y.Rows {
		var file string
		config := make(map[string]interface{})

		for i, c := range v.Cells {
			if row == 0 {
				if i == 0 {
					continue
				}
				keys = append(keys, c.Value)
				continue
			}

			switch i {
			case 0:
				file = "testdata/" + c.Value
			default:
				config[keys[i-1]] = c.Value
			}
		}

		// Leverage ParseTemplates to parse the template files.
		// TODO: move this code around to let it be reusable
		// in a more clear way.
		files, err := manifest.ParseTemplates(file, nil, config)
		if err != nil {
			return err
		}
		if file != "" {
			list, err := ioutil.ReadDir(files)
			if err != nil {
				return err
			}

			for _, f := range list {
				name := path.Join(files, f.Name())
				if !f.IsDir() {
					ff, err := os.Open(name)
					if err != nil {
						return err
					}
					o, err := ParseYAML(bufio.NewReader(ff))
					if err != nil {
						return err
					}

					rt.addWantStatusUpdates(o)
				}
			}
		}
	}
	return nil
}

func (rt *ReconcilerTest) addWantStatusUpdates(objs []unstructured.Unstructured) {
	if rt.Row.WantStatusUpdates == nil {
		rt.Row.WantStatusUpdates = make([]clientgotesting.UpdateActionImpl, 0)
	}

	updates := make([]runtime.Object, 0, len(objs))
	for _, obj := range objs {
		updates = append(updates, obj.DeepCopyObject())
	}
	updates = ToKnownObjects(updates)
	for _, u := range updates {
		updateAction := clientgotesting.UpdateActionImpl{
			Object: u,
		}
		rt.Row.WantStatusUpdates = append(rt.Row.WantStatusUpdates, updateAction)
	}
}

func (rt *ReconcilerTest) expectKubernetesEvents(attributes *messages.PickleStepArgument_PickleTable) error {

	rt.Row.WantEvents = make([]string, 0)

	for _, row := range attributes.Rows {
		eventType := row.Cells[0].Value
		reason := row.Cells[1].Value
		message := row.Cells[2].Value

		if eventType == "Type" {
			// ignore the headers
			continue
		}

		rt.Row.WantEvents = append(rt.Row.WantEvents, fmt.Sprintf(eventType+" "+reason+" "+message))
	}
	return nil
}

var (
	safeDeployDiff = cmpopts.IgnoreUnexported(resource.Quantity{})
)
