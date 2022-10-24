package daemonset

import (
	"context"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type daemonSetGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &daemonSetGetter{client: client}
}

func (d *daemonSetGetter) Get(namespace, name string) (runtime.Object, error) {
	ctx := context.Background()
	ds := &appsv1.DaemonSet{}

	err := d.client.Get(ctx, types.NamespacedName{Name: name}, ds)
	return ds, err
}

func (d *daemonSetGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &appsv1.DaemonSetList{}
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
	for _, daemonSet := range list.Items {
		tmp := daemonSet
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *daemonSetGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftDaemonSet, ok := left.(*appsv1.DaemonSet)
	if !ok {
		return false
	}

	rightDaemonSet, ok := right.(*appsv1.DaemonSet)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftDaemonSet.ObjectMeta, rightDaemonSet.ObjectMeta, field)
}

func (d *daemonSetGetter) filter(object runtime.Object, filter query.Filter) bool {
	daemonSet, ok := object.(*appsv1.DaemonSet)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(daemonsetStatus(&daemonSet.Status), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(daemonSet.ObjectMeta, filter)
	}
}

func daemonsetStatus(status *appsv1.DaemonSetStatus) string {
	if status.DesiredNumberScheduled == 0 && status.NumberReady == 0 {
		return statusStopped
	} else if status.DesiredNumberScheduled == status.NumberReady {
		return statusRunning
	} else {
		return statusUpdating
	}
}
