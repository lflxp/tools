package cluster

import (
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	"github.com/lflxp/tools/sdk/apicache/server/utils"

	v1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"
	nsv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/namespace"
	nodev1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/node"
	scv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/storageclass"

	"github.com/lflxp/tools/sdk/apicache/server/consted"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var mgr manager.Manager

func RegisterCluster(router *gin.Engine, m manager.Manager) {
	mgr = m
	group := router.Group("/cache")
	group.GET("/cluster/:cluster/resource/:resources", ListResource)
	group.GET("/cluster/:cluster/resource/:resources/name/:name", GetResource)
}

// @Summary  获取集群级别资源列表
// @Description 例如node tenant project
// @Tags Cluster
// @Param cluster path string true "集群ID"
// @Param resources path string true "资源名称"
// @Param page query string false "当前页"
// @Param limit query string false "每页限制结果数"
// @Param labelSelector query string false "labelSelector过滤"
// @Param names query string false "fieldBy Metadata.name数组过滤 eg: /namespaces?names=default,kube,xiaoli"
// @Param name query string false "fieldBy Metadata.name基础过滤 eg: /namespaces?name=default"
// @Param namespace query string false "fieldBy Metadata.namespace 命名空间过滤 eg: /namespaces?namespace=dogo-system"
// @Param uid query string false "fieldBy Metadata.uid 过滤 eg: /namespaces?uid=abc-f6a5-4fea-9c1b-e57610115706"
// @Param ownerReference query string false "fieldBy Metadata.OwnerReference.uid 过滤 eg: /namespaces?ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706"
// @Param ownerKind query string false "fieldBy Metadata.OwnerReference.ownerKind 过滤 eg: /namespaces?ownerKind=Workspace"
// @Param annotation query string false "fieldBy Metadata.Annotation 过滤 eg: /namespaces?annotation=enabled=true"
// @Param label query string false "fieldBy Metadata.Label 过滤 eg: /namespaces?label=example.com/workspace=system-workspace"
// @Param sortBy query string false "字段排序，选项: name|creationTimestamp|createTime| 默认：creationTimestamp eg: ?sortBy=creationTimestamp"
// @Param ascending query string false "是否正序排序"
// @Success 200 {string} string "success"
// @Security ApiKeyAuth
// @Router /dogo-cache/cluster/{cluster}/resource/{resources} [get]
func ListResource(c *gin.Context) {
	query := query.ParseQueryParameter(c)
	resource := c.Param("resources")

	var pv1 v1.Interface
	switch resource {
	case consted.Namespace:
		pv1 = nsv1.New(mgr.GetClient())
	case consted.Node:
		pv1 = nodev1.New(mgr.GetClient())
	case consted.StorageClass:
		pv1 = scv1.New(mgr.GetClient())
	default:
		c.String(404, "unsuported type variable")
		return
	}

	data, err := pv1.List("", query)
	if err != nil {
		utils.SendErrorMessage(c, 400, "error", err.Error())
		return
	}
	utils.SendSuccessMessage(c, 200, data)
}

// @Summary  获取指定集群级别资源
// @Description 例如node tenant project
// @Tags Cluster
// @Param cluster path string true "集群ID"
// @Param resources path string true "资源名称"
// @Param name path string true "资源名称"
// @Param page query string false "当前页"
// @Param limit query string false "每页限制结果数"
// @Param labelSelector query string false "labelSelector过滤"
// @Param names query string false "fieldBy Metadata.name数组过滤 eg: /namespaces?names=default,kube,xiaoli"
// @Param name query string false "fieldBy Metadata.name基础过滤 eg: /namespaces?name=default"
// @Param namespace query string false "fieldBy Metadata.namespace 命名空间过滤 eg: /namespaces?namespace=dogo-system"
// @Param uid query string false "fieldBy Metadata.uid 过滤 eg: /namespaces?uid=abc-f6a5-4fea-9c1b-e57610115706"
// @Param ownerReference query string false "fieldBy Metadata.OwnerReference.uid 过滤 eg: /namespaces?ownerReference=a8a8d6cf-f6a5-4fea-9c1b-e57610115706"
// @Param ownerKind query string false "fieldBy Metadata.OwnerReference.ownerKind 过滤 eg: /namespaces?ownerKind=Workspace"
// @Param annotation query string false "fieldBy Metadata.Annotation 过滤 eg: /namespaces?annotation=enabled=true"
// @Param label query string false "fieldBy Metadata.Label 过滤 eg: /namespaces?label=example.com/workspace=system-workspace"
// @Param sortBy query string false "字段排序，选项: name|creationTimestamp|createTime| 默认：creationTimestamp eg: ?sortBy=creationTimestamp"
// @Param ascending query string false "是否正序排序"
// @Success 200 {string} string "success"
// @Security ApiKeyAuth
// @Router /dogo-cache/cluster/{cluster}/resource/{resources}/name/{name} [get]
func GetResource(c *gin.Context) {
	resource := c.Param("resources")
	name := c.Param("name")

	var pv1 v1.Interface
	switch resource {
	case consted.Namespace:
		pv1 = nsv1.New(mgr.GetClient())
	case consted.Node:
		pv1 = nodev1.New(mgr.GetClient())
	case consted.StorageClass:
		pv1 = scv1.New(mgr.GetClient())
	default:
		c.String(404, "unsuported type variable")
		return
	}

	if name != "" {
		data, err := pv1.Get("", name)
		if err != nil {
			utils.SendErrorMessage(c, 400, "error", err.Error())
			return
		}
		utils.SendSuccessMessage(c, 200, data)
	} else {
		utils.SendErrorMessage(c, 400, "resource name is empty", "resource name is empty")
	}
}
