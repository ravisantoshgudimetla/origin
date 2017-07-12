// This file was automatically generated by informer-gen

package internalversion

import (
	authorization "github.com/openshift/origin/pkg/authorization/apis/authorization"
	internalinterfaces "github.com/openshift/origin/pkg/authorization/generated/informers/internalversion/internalinterfaces"
	internalclientset "github.com/openshift/origin/pkg/authorization/generated/internalclientset"
	internalversion "github.com/openshift/origin/pkg/authorization/generated/listers/authorization/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// PolicyBindingInformer provides access to a shared informer and lister for
// PolicyBindings.
type PolicyBindingInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() internalversion.PolicyBindingLister
}

type policyBindingInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

func newPolicyBindingInformer(client internalclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	sharedIndexInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.Authorization().PolicyBindings(v1.NamespaceAll).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.Authorization().PolicyBindings(v1.NamespaceAll).Watch(options)
			},
		},
		&authorization.PolicyBinding{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	return sharedIndexInformer
}

func (f *policyBindingInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&authorization.PolicyBinding{}, newPolicyBindingInformer)
}

func (f *policyBindingInformer) Lister() internalversion.PolicyBindingLister {
	return internalversion.NewPolicyBindingLister(f.Informer().GetIndexer())
}
