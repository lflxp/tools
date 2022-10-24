package serviceaccount

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"
)

type serviceaccountsGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &serviceaccountsGetter{client: client}
}

func (d *serviceaccountsGetter) Get(namespace, name string) (runtime.Object, error) {
	ctx := context.Background()
	sa := &corev1.ServiceAccount{}

	err := d.client.Get(ctx, types.NamespacedName{Name: name}, sa)
	return sa, err
}

func (d *serviceaccountsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &corev1.ServiceAccountList{}
	options := client.ListOptions{
		Namespace:     namespace,
		LabelSelector: query.Selector(),
	}
	// first retrieves all tenant within given namespace
	err := d.client.List(ctx, list, &options)
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, serviceaccount := range list.Items {
		tmp := serviceaccount
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *serviceaccountsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftCM, ok := left.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	rightCM, ok := right.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftCM.ObjectMeta, rightCM.ObjectMeta, field)
}

func (d *serviceaccountsGetter) filter(object runtime.Object, filter query.Filter) bool {
	serviceAccount, ok := object.(*corev1.ServiceAccount)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(serviceAccount.ObjectMeta, filter)
}
