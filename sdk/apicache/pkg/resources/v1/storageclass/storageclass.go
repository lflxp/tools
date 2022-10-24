package storageclass

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	v1storage "k8s.io/api/storage/v1"
)

type storageclassGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &storageclassGetter{client: client}
}

func (p *storageclassGetter) Get(namespace, name string) (runtime.Object, error) {
	ctx := context.Background()
	sc := &v1storage.StorageClass{}

	err := p.client.Get(ctx, types.NamespacedName{Name: name}, sc)
	return sc, err
}

func (p *storageclassGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &v1storage.StorageClassList{}
	options := client.ListOptions{
		Namespace:     namespace,
		LabelSelector: query.Selector(),
	}
	// first retrieves all tenant within given namespace
	err := p.client.List(ctx, list, &options)
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, sc := range list.Items {
		tmp := sc
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *storageclassGetter) compare(obj1, obj2 runtime.Object, field query.Field) bool {
	pv1, ok := obj1.(*v1storage.StorageClass)
	if !ok {
		return false
	}
	pv2, ok := obj2.(*v1storage.StorageClass)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(pv1.ObjectMeta, pv2.ObjectMeta, field)
}

func (p *storageclassGetter) filter(object runtime.Object, filter query.Filter) bool {
	pv, ok := object.(*v1storage.StorageClass)
	if !ok {
		return false
	}
	switch filter.Field {
	default:
		return v1alpha3.DefaultObjectMetaFilter(pv.ObjectMeta, filter)
	}
}
