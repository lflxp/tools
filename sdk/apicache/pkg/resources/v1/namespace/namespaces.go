/*
Copyright 2019 The KubeSphere Authors.

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

package namespace

import (
	"context"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type namespacesGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &namespacesGetter{client: client}
}

func (n namespacesGetter) Get(_, name string) (runtime.Object, error) {
	ctx := context.Background()
	ns := &v1.Namespace{}

	err := n.client.Get(ctx, types.NamespacedName{Name: name}, ns)
	return ns, err
}

func (n namespacesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &v1.NamespaceList{}
	options := client.ListOptions{
		LabelSelector: query.Selector(),
	}
	// first retrieves all tenant within given namespace
	err := n.client.List(ctx, list, &options)
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, item := range list.Items {
		tmp := item
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n namespacesGetter) filter(item runtime.Object, filter query.Filter) bool {
	namespace, ok := item.(*v1.Namespace)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(string(namespace.Status.Phase), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(namespace.ObjectMeta, filter)
	}
}

func (n namespacesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftNs, ok := left.(*v1.Namespace)
	if !ok {
		return false
	}

	rightNs, ok := right.(*v1.Namespace)
	if !ok {
		return true
	}
	return v1alpha3.DefaultObjectMetaCompare(leftNs.ObjectMeta, rightNs.ObjectMeta, field)
}
