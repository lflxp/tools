package cronjob

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type cronjobsGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &cronjobsGetter{client: client}
}

func (d *cronjobsGetter) Get(namespace, name string) (runtime.Object, error) {
	ctx := context.Background()
	m := &batchv1.CronJob{}

	err := d.client.Get(ctx, types.NamespacedName{Name: name}, m)
	return m, err
}

func (d *cronjobsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &batchv1.CronJobList{}
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
	for _, job := range list.Items {
		tmp := job
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *cronjobsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftJob, ok := left.(*batchv1.CronJob)
	if !ok {
		return false
	}

	rightJob, ok := right.(*batchv1.CronJob)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftJob.ObjectMeta, rightJob.ObjectMeta, field)
	}
}

func (d *cronjobsGetter) filter(object runtime.Object, filter query.Filter) bool {
	job, ok := object.(*batchv1.CronJob)
	if !ok {
		return false
	}

	switch filter.Field {
	default:
		return v1alpha3.DefaultObjectMetaFilter(job.ObjectMeta, filter)
	}
}
