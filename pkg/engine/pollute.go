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
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
	app "k8s.io/client-go/pkg/api/v1"
	apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batch "k8s.io/client-go/pkg/apis/batch/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const (
	// defaultPathKey is the key of chart path
	defaultPathKey      = "helm.sh/pollutant"
	defaultNamespaceKey = "helm.sh/namespace"
	defaultReleaseKey   = "helm.sh/release"
	defaultRevisionKey  = "helm.sh/revision"
)

var (
	// defaultSerializer is a codec and used for encoding and decoding kubernetes resources
	defaultSerializer *json.Serializer = nil
)

func init() {
	// create default serializer
	defaultSerializer = json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
}

// pollute adds pollutant to resource annotations. If resource is not a valid kubernetes
// resource, it does nothing and returns original resource.
func pollute(resource string, r *renderable) string {
	// decode object
	obj, _, err := defaultSerializer.Decode([]byte(resource), nil, nil)
	if err != nil {
		return resource
	}
	accessor := meta.NewAccessor()
	annotations, err := accessor.Annotations(obj)
	if err != nil {
		return resource
	}
	err = accessor.SetAnnotations(obj, merge(annotations, r))
	if err != nil {
		return resource
	}

	// check and pollute specific types
	switch ins := obj.(type) {
	case *extensions.Deployment:
		{
			ins.Spec.Template.Annotations = merge(ins.Spec.Template.Annotations, r)
		}
	case *extensions.DaemonSet:
		{
			ins.Spec.Template.Annotations = merge(ins.Spec.Template.Annotations, r)
		}
	case *extensions.ReplicaSet:
		{
			ins.Spec.Template.Annotations = merge(ins.Spec.Template.Annotations, r)
		}
	case *apps.StatefulSet:
		{
			ins.Spec.Template.Annotations = merge(ins.Spec.Template.Annotations, r)
		}
	case *batch.Job:
		{
			ins.Spec.Template.Annotations = merge(ins.Spec.Template.Annotations, r)
		}
	case *app.ReplicationController:
		{
			ins.Spec.Template.Annotations = merge(ins.Spec.Template.Annotations, r)
		}
	}

	// encode object
	buf := bytes.NewBuffer(nil)
	err = defaultSerializer.Encode(obj, buf)
	if err != nil {
		return resource
	}
	return buf.String()
}

// merge merges renderable info into origin.
func merge(origin map[string]string, r *renderable) map[string]string {
	if origin == nil {
		origin = make(map[string]string)
	}
	origin[defaultPathKey] = r.path
	values := r.vals.AsMap()
	rmap, ok := values["Release"]
	if !ok {
		return origin
	}
	release, ok := rmap.(map[string]interface{})
	if !ok {
		return origin
	}
	if value, ok := release["Namespace"]; ok {
		origin[defaultNamespaceKey] = fmt.Sprint(value)
	}
	if value, ok := release["Name"]; ok {
		origin[defaultReleaseKey] = fmt.Sprint(value)
	}
	if value, ok := release["Revision"]; ok {
		origin[defaultRevisionKey] = fmt.Sprint(value)
	}
	return origin
}
