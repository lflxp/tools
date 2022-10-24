package deployment

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type deploymentsGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &deploymentsGetter{client: client}
}

func (d *deploymentsGetter) Get(namespace, name string) (runtime.Object, error) {
	// return d.sharedInformers.Apps().V1().Deployments().Lister().Deployments(namespace).Get(name)
	ctx := context.Background()
	deploy := &v1.Deployment{}

	err := d.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, deploy)
	return deploy, err
}

func (d *deploymentsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	deployList := &v1.DeploymentList{}
	options := client.ListOptions{
		Namespace:     namespace,
		LabelSelector: query.Selector(),
	}
	// first retrieves all deployments within given namespace
	err := d.client.List(ctx, deployList, &options)
	// deployments, err := d.sharedInformers.Apps().V1().Deployments().Lister().Deployments(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, deployment := range deployList.Items {
		tmp := deployment
		result = append(result, &tmp)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *deploymentsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftDeployment, ok := left.(*v1.Deployment)
	if !ok {
		return false
	}

	rightDeployment, ok := right.(*v1.Deployment)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftDeployment).After(lastUpdateTime(rightDeployment))
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftDeployment.ObjectMeta, rightDeployment.ObjectMeta, field)
	}
}

func (d *deploymentsGetter) filter(object runtime.Object, filter query.Filter) bool {
	deployment, ok := object.(*v1.Deployment)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(deploymentStatus(deployment.Status), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(deployment.ObjectMeta, filter)
	}
}

func deploymentStatus(status v1.DeploymentStatus) string {
	if status.ReadyReplicas == 0 && status.Replicas == 0 {
		return statusStopped
	} else if status.ReadyReplicas == status.Replicas {
		return statusRunning
	} else {
		return statusUpdating
	}
}

func lastUpdateTime(deployment *v1.Deployment) time.Time {
	lut := deployment.CreationTimestamp.Time
	for _, condition := range deployment.Status.Conditions {
		if condition.LastUpdateTime.After(lut) {
			lut = condition.LastUpdateTime.Time
		}
	}
	return lut
}
