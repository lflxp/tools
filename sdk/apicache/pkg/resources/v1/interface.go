package v1

import (
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/lflxp/tools/sdk/apicache/pkg/api"

	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
)

type Interface interface {
	// Get retrieves a single object by its namespace and name
	Get(namespace, name string) (runtime.Object, error)

	// List retrieves a collection of objects matches given query
	List(namespace string, query *query.Query) (*api.ListResult, error)
}

// CompareFunc return true is left great than right
type CompareFunc func(runtime.Object, runtime.Object, query.Field) bool

type FilterFunc func(runtime.Object, query.Filter) bool

type TransformFunc func(runtime.Object) runtime.Object

func DefaultList(objects []runtime.Object, q *query.Query, compareFunc CompareFunc, filterFunc FilterFunc, transformFuncs ...TransformFunc) *api.ListResult {
	// selected matched ones
	var filtered []runtime.Object
	for _, object := range objects {
		selected := true
		for field, value := range q.Filters {
			if !filterFunc(object, query.Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}

		if selected {
			for _, transform := range transformFuncs {
				object = transform(object)
			}
			filtered = append(filtered, object)
		}
	}

	// sort by sortBy field
	sort.Slice(filtered, func(i, j int) bool {
		if !q.Ascending {
			return compareFunc(filtered[i], filtered[j], q.SortBy)
		}
		return !compareFunc(filtered[i], filtered[j], q.SortBy)
	})

	total := len(filtered)

	if q.Pagination == nil {
		q.Pagination = query.NoPagination
	}

	start, end := q.Pagination.GetValidPagination(total)

	return &api.ListResult{
		Data: objectsToInterfaces(filtered[start:end]),
		Pagination: api.Pagination{
			Limit:  q.Pagination.Limit,
			Total:  len(filtered),
			Offset: q.Pagination.Offset,
			Page:   q.Pagination.Page,
		},
	}
}

// DefaultObjectMetaCompare return true is left great than right
func DefaultObjectMetaCompare(left, right metav1.ObjectMeta, sortBy query.Field) bool {
	switch sortBy {
	// ?sortBy=name
	case query.FieldName:
		return strings.Compare(left.Name, right.Name) > 0
	//	?sortBy=creationTimestamp
	default:
		fallthrough
	case query.FieldCreateTime:
		fallthrough
	case query.FieldCreationTimeStamp:
		// compare by name if creation timestamp is equal
		if left.CreationTimestamp.Equal(&right.CreationTimestamp) {
			return strings.Compare(left.Name, right.Name) > 0
		}
		return left.CreationTimestamp.After(right.CreationTimestamp.Time)
	}
}

// Default metadata filter
func DefaultObjectMetaFilter(item metav1.ObjectMeta, filter query.Filter) bool {
	switch filter.Field {
	case query.FieldNames:
		for _, name := range strings.Split(string(filter.Value), ",") {
			if item.Name == name {
				return true
			}
		}
		return false
	// /namespaces?page=1&limit=10&name=default
	case query.FieldName:
		return strings.Contains(item.Name, string(filter.Value))
		// /namespaces?page=1&limit=10&uid=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldUID:
		return strings.Compare(string(item.UID), string(filter.Value)) == 0
		// /deployments?page=1&limit=10&namespace=kubesphere-system
	case query.FieldNamespace:
		return strings.Compare(item.Namespace, string(filter.Value)) == 0
		// /namespaces?page=1&limit=10&ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706
	case query.FieldOwnerReference:
		for _, ownerReference := range item.OwnerReferences {
			if strings.Compare(string(ownerReference.UID), string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&ownerKind=Workspace
	case query.FieldOwnerKind:
		for _, ownerReference := range item.OwnerReferences {
			if strings.Compare(ownerReference.Kind, string(filter.Value)) == 0 {
				return true
			}
		}
		return false
		// /namespaces?page=1&limit=10&annotation=openpitrix=bad
	case query.FieldAnnotation:
		return labelMatch(item.Annotations, string(filter.Value))
		// /namespaces?page=1&limit=10&label=kubesphere.io/workspace=system-workspace
	case query.FieldLabel:
		return labelMatch(item.Labels, string(filter.Value))
	case query.FieldDogo:
		return customMatch(item.Annotations, string(filter.Value))
	default:
		return false
	}
}

// dogo=a=b,c=d
// dogo=a=b||c=d
// dogo=a=b||c=d,e=f
// dogo=a=b||c=d,e=f||d=h
func customMatch(labels map[string]string, filter string) bool {
	status := false

	ands := strings.Split(filter, ",")
	if len(ands) > 0 {
		andstatus := make([]bool, len(ands))
		for i, and := range ands {
			ors := strings.Split(and, "||")
			if len(ors) <= 1 {
				fields := strings.SplitN(and, "=", 2)
				var key, value string
				var opposite bool
				if len(fields) == 2 {
					key = fields[0]
					if strings.HasSuffix(key, "!") {
						key = strings.TrimSuffix(key, "!")
						opposite = true
					}
					value = fields[1]
				} else {
					key = fields[0]
					value = "*"
				}
				for k, v := range labels {
					if opposite {
						if (k == key) && !strings.Contains(v, value) {
							andstatus[i] = true
							break
						}
					} else {
						if (k == key) && (value == "*" || strings.Contains(v, value)) {
							andstatus[i] = true
							break
						}
					}
				}
			} else if len(ors) > 1 {
				orstatus := false
				for _, or := range ors {
					fields := strings.SplitN(or, "=", 2)
					var key, value string
					var opposite bool
					if len(fields) == 2 {
						key = fields[0]
						if strings.HasSuffix(key, "!") {
							key = strings.TrimSuffix(key, "!")
							opposite = true
						}
						value = fields[1]
					} else {
						key = fields[0]
						value = "*"
					}

					for k, v := range labels {
						if opposite {
							if (k == key) && !strings.Contains(v, value) {
								orstatus = true
								break
							}
						} else {
							if (k == key) && (value == "*" || strings.Contains(v, value)) {
								orstatus = true
								break
							}
						}
					}

					if orstatus {
						andstatus[i] = true
						break
					}
				}
			}
		}

		status = true
		for _, isok := range andstatus {
			if !isok {
				status = false
				break
			}
		}
	}

	return status
}

func labelMatch(labels map[string]string, filter string) bool {
	fields := strings.SplitN(filter, "=", 2)
	var key, value string
	var opposite bool
	if len(fields) == 2 {
		key = fields[0]
		if strings.HasSuffix(key, "!") {
			key = strings.TrimSuffix(key, "!")
			opposite = true
		}
		value = fields[1]
	} else {
		key = fields[0]
		value = "*"
	}
	for k, v := range labels {
		if opposite {
			if (k == key) && v != value {
				return true
			}
		} else {
			if (k == key) && (value == "*" || v == value) {
				return true
			}
		}
	}
	return false
}

func objectsToInterfaces(objs []runtime.Object) []interface{} {
	res := make([]interface{}, 0)
	for _, obj := range objs {
		res = append(res, obj)
	}
	return res
}
