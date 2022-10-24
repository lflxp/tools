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

package node

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	resourceheper "k8s.io/kubectl/pkg/util/resource"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	v1alpha3 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Those annotations were added to node only for display purposes
const (
	nodeCPURequests                                 = "node.example.com/cpu-requests"
	nodeMemoryRequests                              = "node.example.com/memory-requests"
	nodeCPULimits                                   = "node.example.com/cpu-limits"
	nodeMemoryLimits                                = "node.example.com/memory-limits"
	nodeCPURequestsFraction                         = "node.example.com/cpu-requests-fraction"
	nodeCPULimitsFraction                           = "node.example.com/cpu-limits-fraction"
	nodeMemoryRequestsFraction                      = "node.example.com/memory-requests-fraction"
	nodeMemoryLimitsFraction                        = "node.example.com/memory-limits-fraction"
	nodeConfigOK               v1.NodeConditionType = "ConfigOK"
	nodeKubeletReady           v1.NodeConditionType = "KubeletReady"
	statusRunning                                   = "running"
	statusWarning                                   = "warning"
	statusUnschedulable                             = "unschedulable"
)

type nodesGetter struct {
	client client.Client
}

func New(client client.Client) v1alpha3.Interface {
	return &nodesGetter{
		client: client,
	}
}

func (n *nodesGetter) Get(_, name string) (runtime.Object, error) {
	ctx := context.Background()
	node := &v1.Node{}
	err := n.client.Get(ctx, types.NamespacedName{Name: name}, node)
	if err != nil {
		return node, err
	}

	// ignore the error, skip annotating process if error happened
	podList := &v1.PodList{}
	err = n.client.List(ctx, podList, nil)
	if err != nil {
		return podList, err
	}

	// Never mutate original objects!
	// Caches are shared across controllers,
	// this means that if you mutate your "copy" (actually a reference or shallow copy) of an object,
	// you'll mess up other controllers (not just your own).
	// Also, if the mutated field is a map,
	// a "concurrent map (read &) write" panic might occur,
	// causing the ks-apiserver to crash.
	// Refer:
	// https://github.com/kubesphere/kubesphere/issues/4357
	// https://github.com/kubesphere/kubesphere/issues/3469
	// https://github.com/kubesphere/kubesphere/pull/4599
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/controllers.md
	node = node.DeepCopy()
	tmp := []*v1.Pod{}
	for _, x := range podList.Items {
		y := x
		tmp = append(tmp, &y)
	}
	n.annotateNode(node, tmp)

	return node, nil
}

func (n *nodesGetter) List(_ string, q *query.Query) (*api.ListResult, error) {
	ctx := context.Background()
	nodeList := &v1.NodeList{}

	options := client.ListOptions{
		LabelSelector: q.Selector(),
	}
	// first retrieves all deployments within given namespace
	err := n.client.List(ctx, nodeList, &options)
	if err != nil {
		return nil, err
	}

	var filtered []*v1.Node
	for _, object := range nodeList.Items {
		selected := true
		for field, value := range q.Filters {
			if !n.filter(&object, query.Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}

		if selected {
			tmp := object
			filtered = append(filtered, &tmp)
		}
	}

	// sort by sortBy field
	sort.Slice(filtered, func(i, j int) bool {
		if !q.Ascending {
			return n.compare(filtered[i], filtered[j], q.SortBy)
		}
		return !n.compare(filtered[i], filtered[j], q.SortBy)
	})

	total := len(filtered)
	if q.Pagination == nil {
		q.Pagination = query.NoPagination
	}
	start, end := q.Pagination.GetValidPagination(total)
	selectedNodes := filtered[start:end]

	// ignore the error, skip annotating process if error happened
	pods := &v1.PodList{}
	err = n.client.List(ctx, pods)
	if err != nil {
		return nil, err
	}

	var nonTerminatedPodsList []*v1.Pod
	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
			tmp := pod
			nonTerminatedPodsList = append(nonTerminatedPodsList, &tmp)
		}
	}

	var result = make([]interface{}, 0)
	for _, node := range selectedNodes {
		node = node.DeepCopy()
		n.annotateNode(node, nonTerminatedPodsList)
		result = append(result, node)
	}

	return &api.ListResult{
		Data: result,
		Pagination: api.Pagination{
			Limit:  q.Pagination.Limit,
			Total:  total,
			Offset: q.Pagination.Offset,
			Page:   q.Pagination.Page,
		},
	}, nil
}

func (n *nodesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftNode, ok := left.(*v1.Node)
	if !ok {
		return false
	}

	rightNode, ok := right.(*v1.Node)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftNode.ObjectMeta, rightNode.ObjectMeta, field)
}

func (n *nodesGetter) filter(object runtime.Object, filter query.Filter) bool {
	node, ok := object.(*v1.Node)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return getNodeStatus(node) == string(filter.Value)
	}

	return v1alpha3.DefaultObjectMetaFilter(node.ObjectMeta, filter)
}

// annotateNode adds cpu/memory requests usage data to node's annotations
// this operation mutates the *v1.Node passed in
// so DO A DEEPCOPY before calling
func (n *nodesGetter) annotateNode(node *v1.Node, pods []*v1.Pod) {
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	if len(pods) == 0 {
		return
	}

	var nodePods []*v1.Pod
	for _, pod := range pods {
		if pod.Spec.NodeName == node.Name {
			nodePods = append(nodePods, pod)
		}
	}

	reqs, limits := n.getPodsTotalRequestAndLimits(nodePods)

	cpuReqs, cpuLimits, memoryReqs, memoryLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU], reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
	node.Annotations[nodeCPURequests] = cpuReqs.String()
	node.Annotations[nodeCPULimits] = cpuLimits.String()
	node.Annotations[nodeMemoryRequests] = memoryReqs.String()
	node.Annotations[nodeMemoryLimits] = memoryLimits.String()

	fractionCpuReqs, fractionCpuLimits := float64(0), float64(0)
	allocatable := node.Status.Allocatable
	if allocatable.Cpu().MilliValue() != 0 {
		fractionCpuReqs = float64(cpuReqs.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
		fractionCpuLimits = float64(cpuLimits.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	}
	fractionMemoryReqs, fractionMemoryLimits := float64(0), float64(0)
	if allocatable.Memory().Value() != 0 {
		fractionMemoryReqs = float64(memoryReqs.Value()) / float64(allocatable.Memory().Value()) * 100
		fractionMemoryLimits = float64(memoryLimits.Value()) / float64(allocatable.Memory().Value()) * 100
	}

	node.Annotations[nodeCPURequestsFraction] = fmt.Sprintf("%d%%", int(fractionCpuReqs))
	node.Annotations[nodeCPULimitsFraction] = fmt.Sprintf("%d%%", int(fractionCpuLimits))
	node.Annotations[nodeMemoryRequestsFraction] = fmt.Sprintf("%d%%", int(fractionMemoryReqs))
	node.Annotations[nodeMemoryLimitsFraction] = fmt.Sprintf("%d%%", int(fractionMemoryLimits))
}

func (n *nodesGetter) getPodsTotalRequestAndLimits(pods []*v1.Pod) (reqs map[v1.ResourceName]resource.Quantity, limits map[v1.ResourceName]resource.Quantity) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}
	for _, pod := range pods {
		podReqs, podLimits := resourceheper.PodRequestsAndLimits(pod)
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = podReqValue.DeepCopy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = podLimitValue.DeepCopy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return
}

func getNodeStatus(node *v1.Node) string {
	if node.Spec.Unschedulable {
		return statusUnschedulable
	}
	for _, condition := range node.Status.Conditions {
		if isUnhealthyStatus(condition) {
			return statusWarning
		}
	}

	return statusRunning
}

var expectedConditions = map[v1.NodeConditionType]v1.ConditionStatus{
	v1.NodeMemoryPressure:     v1.ConditionFalse,
	v1.NodeDiskPressure:       v1.ConditionFalse,
	v1.NodePIDPressure:        v1.ConditionFalse,
	v1.NodeNetworkUnavailable: v1.ConditionFalse,
	nodeConfigOK:              v1.ConditionTrue,
	nodeKubeletReady:          v1.ConditionTrue,
	v1.NodeReady:              v1.ConditionTrue,
}

func isUnhealthyStatus(condition v1.NodeCondition) bool {
	expectedStatus := expectedConditions[condition.Type]
	if expectedStatus != "" && condition.Status != expectedStatus {
		return true
	}
	return false
}
