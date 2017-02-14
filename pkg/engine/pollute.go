/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package engine

import (
	"bytes"
	"os"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/apimachinery/announced"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	_ "k8s.io/kubernetes/pkg/apis/apps/install"
	_ "k8s.io/kubernetes/pkg/apis/authentication/install"
	_ "k8s.io/kubernetes/pkg/apis/authorization/install"
	_ "k8s.io/kubernetes/pkg/apis/autoscaling/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/certificates/install"
	_ "k8s.io/kubernetes/pkg/apis/componentconfig/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
	_ "k8s.io/kubernetes/pkg/apis/imagepolicy/install"
	_ "k8s.io/kubernetes/pkg/apis/policy/install"
	_ "k8s.io/kubernetes/pkg/apis/rbac/install"
	_ "k8s.io/kubernetes/pkg/apis/storage/install"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/serializer/json"
)

const (
	// defaultPollutantKey is the key of pollutant
	defaultPollutantKey = "helm.sh/pollutant"
)

var (
	// defaultSerializer is a codec and used for encoding and decoding kubernetes resources
	defaultSerializer *json.Serializer = nil
)

func init() {
	// create default serializer
	schema := runtime.NewScheme()
	errs := []error{api.AddToScheme(schema), v1.AddToScheme(schema)}
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
	announced.DefaultGroupFactoryRegistry.RegisterAndEnableAll(registered.NewOrDie(os.Getenv("KUBE_API_VERSIONS")), schema)
	defaultSerializer = json.NewYAMLSerializer(json.DefaultMetaFactory, schema, schema)
}

// pollute adds pollutant to resource annotations. If resource is not a valid kubernetes
// resource, it does not pollute the resource and returns original resource.
func pollute(resource string, pollutant string) string {
	obj, _, err := defaultSerializer.Decode([]byte(resource), nil, nil)
	if err != nil {
		return resource
	}
	err = meta.NewAccessor().SetAnnotations(obj, map[string]string{
		defaultPollutantKey: pollutant,
	})
	if err != nil {
		return resource
	}
	buf := bytes.NewBuffer(nil)
	err = defaultSerializer.Encode(obj, buf)
	if err != nil {
		return resource
	}
	return buf.String()
}
