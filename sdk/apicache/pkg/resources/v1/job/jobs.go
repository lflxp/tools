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

package job

import (
	"context"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	jobFailed    = "failed"
	jobCompleted = "completed"
	jobRunning   = "running"
)

type jobsGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &jobsGetter{client: client}
}

func (d *jobsGetter) Get(namespace, name string) (runtime.Object, error) {
	ctx := context.Background()
	m := &batchv1.Job{}

	err := d.client.Get(ctx, types.NamespacedName{Name: name}, m)
	return m, err
}

func (d *jobsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	list := &batchv1.JobList{}
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

func (d *jobsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftJob, ok := left.(*batchv1.Job)
	if !ok {
		return false
	}

	rightJob, ok := right.(*batchv1.Job)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftJob).After(lastUpdateTime(rightJob))
	case query.FieldStatus:
		return strings.Compare(jobStatus(leftJob.Status), jobStatus(rightJob.Status)) > 0
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftJob.ObjectMeta, rightJob.ObjectMeta, field)
	}
}

func (d *jobsGetter) filter(object runtime.Object, filter query.Filter) bool {
	job, ok := object.(*batchv1.Job)
	if !ok {
		return false
	}

	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(jobStatus(job.Status), string(filter.Value)) == 0
	default:
		return v1alpha3.DefaultObjectMetaFilter(job.ObjectMeta, filter)
	}
}

func jobStatus(status batchv1.JobStatus) string {
	for _, condition := range status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue {
			return jobCompleted
		} else if condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue {
			return jobFailed
		}
	}

	return jobRunning
}

func lastUpdateTime(job *batchv1.Job) time.Time {
	lut := job.CreationTimestamp.Time
	for _, condition := range job.Status.Conditions {
		if condition.LastTransitionTime.After(lut) {
			lut = condition.LastTransitionTime.Time
		}
	}
	return lut
}
