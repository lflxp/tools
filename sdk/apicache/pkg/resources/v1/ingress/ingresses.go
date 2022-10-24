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

package ingress

import (
	"context"

	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"
)

type ingressGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &ingressGetter{client: client}
}

func (i *ingressGetter) Get(namespace, name string) (runtime.Object, error) {
	ctx := context.Background()
	cm := &v1.Ingress{}

	err := i.client.Get(ctx, types.NamespacedName{Name: name}, cm)
	return cm, err
}

func (i *ingressGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &v1.IngressList{}
	options := client.ListOptions{
		LabelSelector: query.Selector(),
	}
	// first retrieves all tenant within given namespace
	err := i.client.List(ctx, list, &options)
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, user := range list.Items {
		tmp := user
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, i.compare, i.filter), nil
}

func (i *ingressGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftIngress, ok := left.(*v1.Ingress)
	if !ok {
		return false
	}

	rightIngress, ok := right.(*v1.Ingress)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftIngress.ObjectMeta, rightIngress.ObjectMeta, field)
	}
}

func (i *ingressGetter) filter(object runtime.Object, filter query.Filter) bool {
	deployment, ok := object.(*v1.Ingress)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaFilter(deployment.ObjectMeta, filter)
}
