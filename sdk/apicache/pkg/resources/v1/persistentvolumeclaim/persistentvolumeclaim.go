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

package persistentvolumeclaim

import (
	"context"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"
)

const (
	storageClassName = "storageClassName"

	annotationInUse              = "example.com/in-use"
	annotationAllowSnapshot      = "example.com/allow-snapshot"
	annotationStorageProvisioner = "volume.beta.example.com/storage-provisioner"
)

type persistentVolumeClaimGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &persistentVolumeClaimGetter{client: client}
}

func (p *persistentVolumeClaimGetter) Get(namespace, name string) (runtime.Object, error) {
	ctx := context.Background()
	pvc := &v1.PersistentVolumeClaim{}

	err := p.client.Get(ctx, types.NamespacedName{Name: name}, pvc)
	if err != nil {
		return pvc, err
	}
	// we should never mutate the shared objects from informers
	pvc = pvc.DeepCopy()
	p.annotatePVC(pvc)
	return pvc, nil
}

func (p *persistentVolumeClaimGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &v1.PersistentVolumeClaimList{}
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
	for _, pvc := range list.Items {
		tmp := pvc
		p.annotatePVC(&tmp)
		result = append(result, &tmp)
	}
	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *persistentVolumeClaimGetter) compare(left, right runtime.Object, field query.Field) bool {
	leftSnapshot, ok := left.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	rightSnapshot, ok := right.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaCompare(leftSnapshot.ObjectMeta, rightSnapshot.ObjectMeta, field)
}

func (p *persistentVolumeClaimGetter) filter(object runtime.Object, filter query.Filter) bool {
	pvc, ok := object.(*v1.PersistentVolumeClaim)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.EqualFold(string(pvc.Status.Phase), string(filter.Value))
	case storageClassName:
		return pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(pvc.ObjectMeta, filter)
	}
}

func (p *persistentVolumeClaimGetter) annotatePVC(pvc *v1.PersistentVolumeClaim) {
	inUse := p.countPods(pvc.Name, pvc.Namespace)
	if pvc.Annotations == nil {
		pvc.Annotations = make(map[string]string)
	}
	pvc.Annotations[annotationInUse] = strconv.FormatBool(inUse)
}

func (p *persistentVolumeClaimGetter) countPods(name, namespace string) bool {
	ctx := context.Background()
	list := &v1.PodList{}
	options := client.ListOptions{
		Namespace: namespace,
	}
	// first retrieves all tenant within given namespace
	err := p.client.List(ctx, list, &options)
	if err != nil {
		return false
	}

	for _, pod := range list.Items {
		for _, pvc := range pod.Spec.Volumes {
			if pvc.PersistentVolumeClaim != nil && pvc.PersistentVolumeClaim.ClaimName == name {
				return true
			}
		}
	}
	return false
}
