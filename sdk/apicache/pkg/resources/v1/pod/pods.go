package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fieldNodeName    = "nodeName"
	fieldPVCName     = "pvcName"
	fieldServiceName = "serviceName"
	fieldStatus      = "status"
)

type podsGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &podsGetter{client: client}
}

func (p *podsGetter) Get(namespace, name string) (runtime.Object, error) {
	// return p.informer.Core().V1().Pods().Lister().Pods(namespace).Get(name)
	ctx := context.Background()
	deploy := &corev1.Pod{}

	err := p.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, deploy)
	return deploy, err
}

func (p *podsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	pods := &corev1.PodList{}
	options := client.ListOptions{
		Namespace:     namespace,
		LabelSelector: query.Selector(),
	}
	// first retrieves all deployments within given namespace
	err := p.client.List(ctx, pods, &options)
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, pod := range pods.Items {
		tmp := pod
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *podsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftPod, ok := left.(*corev1.Pod)
	if !ok {
		return false
	}

	rightPod, ok := right.(*corev1.Pod)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftPod.ObjectMeta, rightPod.ObjectMeta, field)
}

func (p *podsGetter) filter(object runtime.Object, filter query.Filter) bool {
	pod, ok := object.(*corev1.Pod)

	if !ok {
		return false
	}
	switch filter.Field {
	case fieldNodeName:
		return pod.Spec.NodeName == string(filter.Value)
	case fieldPVCName:
		return p.podBindPVC(pod, string(filter.Value))
	case fieldServiceName:
		return p.podBelongToService(pod, string(filter.Value))
	case fieldStatus:
		return string(pod.Status.Phase) == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(pod.ObjectMeta, filter)
	}
}

func (p *podsGetter) podBindPVC(item *corev1.Pod, pvcName string) bool {
	for _, v := range item.Spec.Volumes {
		if v.VolumeSource.PersistentVolumeClaim != nil &&
			v.VolumeSource.PersistentVolumeClaim.ClaimName == pvcName {
			return true
		}
	}
	return false
}

func (p *podsGetter) podBelongToService(item *corev1.Pod, serviceName string) bool {
	ctx := context.Background()
	service := &corev1.Service{}
	err := p.client.Get(ctx, types.NamespacedName{Namespace: item.Namespace, Name: serviceName}, service)
	// service, err := p.informer.Core().V1().Services().Lister().Services(item.Namespace).Get(serviceName)
	if err != nil {
		return false
	}
	selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
	if selector.Empty() || !selector.Matches(labels.Set(item.Labels)) {
		return false
	}
	return true
}
