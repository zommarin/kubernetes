/*
Copyright 2016 The Kubernetes Authors.

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

package core

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/quota"
	"k8s.io/kubernetes/pkg/quota/generic"
)

// listConfigMapsByNamespaceFuncUsingClient returns a configMap listing function based on the provided client.
func listConfigMapsByNamespaceFuncUsingClient(kubeClient clientset.Interface) generic.ListFuncByNamespace {
	// TODO: ideally, we could pass dynamic client pool down into this code, and have one way of doing this.
	// unfortunately, dynamic client works with Unstructured objects, and when we calculate Usage, we require
	// structured objects.
	return func(namespace string, options metav1.ListOptions) ([]runtime.Object, error) {
		itemList, err := kubeClient.CoreV1().ConfigMaps(namespace).List(options)
		if err != nil {
			return nil, err
		}
		results := make([]runtime.Object, 0, len(itemList.Items))
		for i := range itemList.Items {
			results = append(results, &itemList.Items[i])
		}
		return results, nil
	}
}

// NewConfigMapEvaluator returns an evaluator that can evaluate configMaps
// if the specified shared informer factory is not nil, evaluator may use it to support listing functions.
func NewConfigMapEvaluator(kubeClient clientset.Interface, f informers.SharedInformerFactory) quota.Evaluator {
	listFuncByNamespace := listConfigMapsByNamespaceFuncUsingClient(kubeClient)
	if f != nil {
		listFuncByNamespace = generic.ListResourceUsingInformerFunc(f, v1.SchemeGroupVersion.WithResource("configmaps"))
	}
	return &generic.ObjectCountEvaluator{
		AllowCreateOnUpdate: false,
		InternalGroupKind:   api.Kind("ConfigMap"),
		ResourceName:        api.ResourceConfigMaps,
		ListFuncByNamespace: listFuncByNamespace,
	}
}
