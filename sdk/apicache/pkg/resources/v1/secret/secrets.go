package secret

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type secretSearcher struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &secretSearcher{client: client}
}

func (s *secretSearcher) Get(namespace, name string) (runtime.Object, error) {
	ctx := context.Background()
	sc := &v1.Secret{}

	err := s.client.Get(ctx, types.NamespacedName{Name: name}, sc)
	return sc, err
}

func (s *secretSearcher) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &v1.SecretList{}
	options := client.ListOptions{
		LabelSelector: query.Selector(),
	}
	// first retrieves all tenant within given namespace
	err := s.client.List(ctx, list, &options)
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, secret := range list.Items {
		tmp := secret
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, s.compare, s.filter), nil
}

func (s *secretSearcher) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftSecret, ok := left.(*v1.Secret)
	if !ok {
		return false
	}

	rightSecret, ok := right.(*v1.Secret)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftSecret.ObjectMeta, rightSecret.ObjectMeta, field)
}

func (s *secretSearcher) filter(object runtime.Object, filter query.Filter) bool {
	secret, ok := object.(*v1.Secret)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(secret.ObjectMeta, filter)
}
