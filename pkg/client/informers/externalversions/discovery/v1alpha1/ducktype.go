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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	discoveryv1alpha1 "knative.dev/discovery/pkg/apis/discovery/v1alpha1"
	versioned "knative.dev/discovery/pkg/client/clientset/versioned"
	internalinterfaces "knative.dev/discovery/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "knative.dev/discovery/pkg/client/listers/discovery/v1alpha1"
)

// DuckTypeInformer provides access to a shared informer and lister for
// DuckTypes.
type DuckTypeInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.DuckTypeLister
}

type duckTypeInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewDuckTypeInformer constructs a new informer for DuckType type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewDuckTypeInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredDuckTypeInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredDuckTypeInformer constructs a new informer for DuckType type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredDuckTypeInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DiscoveryV1alpha1().DuckTypes(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DiscoveryV1alpha1().DuckTypes(namespace).Watch(options)
			},
		},
		&discoveryv1alpha1.DuckType{},
		resyncPeriod,
		indexers,
	)
}

func (f *duckTypeInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredDuckTypeInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *duckTypeInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&discoveryv1alpha1.DuckType{}, f.defaultInformer)
}

func (f *duckTypeInformer) Lister() v1alpha1.DuckTypeLister {
	return v1alpha1.NewDuckTypeLister(f.Informer().GetIndexer())
}
