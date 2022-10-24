package namespace

import (
	"github.com/lflxp/tools/sdk/apicache/pkg/apiserver/query"
	"github.com/lflxp/tools/sdk/apicache/server/utils"

	v1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1"
	cmv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/configmap"
	cronjobv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/cronjob"
	dsv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/daemonset"
	deployv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/deployment"
	ingressv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/ingress"
	jobv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/job"
	pvv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/persistentvolume"
	pvcv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/persistentvolumeclaim"
	podv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/pod"
	rolev1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/role"
	rolebv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/rolebinding"
	secretv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/secret"
	servicev1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/service"
	sav1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/serviceaccount"
	stsv1 "github.com/lflxp/tools/sdk/apicache/pkg/resources/v1/statefulset"

	. "github.com/lflxp/tools/sdk/apicache/server/consted"

	"github.com/gin-gonic/gin"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var mgr manager.Manager

func RegisterNamespace(router *gin.Engine, m manager.Manager) {
	mgr = m
	group := router.Group("/cache")
	group.GET("/cluster/:cluster/resource/:resources/namespace/:namespace", ListResource)
	group.GET("/cluster/:cluster/resource/:resources/namespace/:namespace/:name", GetResource)
}

// @Summary  获取指定命名空间级别资源
// @Description 例如ingress svc cm
// @Tags Namespace
// @Param cluster path string true "集群ID"
// @Param resources path string true "资源名称"
// @Param namespace path string true "命名空间名称"
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
// @Router /dogo-cache/cluster/{cluster}/resource/{resources}/namespace/{namespace} [get]
func ListResource(c *gin.Context) {
	query := query.ParseQueryParameter(c)
	resource := c.Param("resources")
	namespace := c.Param("namespace")

	var pv1 v1.Interface
	switch resource {
	case Deployment:
		pv1 = deployv1.New(mgr.GetClient())
	case Pod:
		pv1 = podv1.New(mgr.GetClient())
	case ConfigMap:
		pv1 = cmv1.New(mgr.GetClient())
	case CronJob:
		pv1 = cronjobv1.New(mgr.GetClient())
	case Daemonset:
		pv1 = dsv1.New(mgr.GetClient())
	case Ingress:
		pv1 = ingressv1.New(mgr.GetClient())
	case Job:
		pv1 = jobv1.New(mgr.GetClient())
	case Pvc:
		pv1 = pvcv1.New(mgr.GetClient())
	case Pv:
		pv1 = pvv1.New(mgr.GetClient())
	case Role:
		pv1 = rolev1.New(mgr.GetClient())
	case RoleBinding:
		pv1 = rolebv1.New(mgr.GetClient())
	case Secret:
		pv1 = secretv1.New(mgr.GetClient())
	case Service:
		pv1 = servicev1.New(mgr.GetClient())
	case ServiceAccount:
		pv1 = sav1.New(mgr.GetClient())
	case Statefulset:
		pv1 = stsv1.New(mgr.GetClient())
	default:
		c.String(404, "unsuported type variable")
		return
	}

	data, err := pv1.List(namespace, query)
	if err != nil {
		utils.SendErrorMessage(c, 400, "error", err.Error())
		return
	}
	utils.SendSuccessMessage(c, 200, data)
}

// @Summary  获取指定命名空间级别资源
// @Description 例如ingress svc cm
// @Tags Namespace
// @Param cluster path string true "集群ID"
// @Param resources path string true "资源名称"
// @Param namespace path string true "命名空间名称"
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
// @Router /dogo-cache/cluster/{cluster}/resource/{resources}/namespace/{namespace}/{name} [get]
func GetResource(c *gin.Context) {
	resource := c.Param("resources")
	namespace := c.Param("namespace")
	name := c.Param("name")

	var pv1 v1.Interface
	switch resource {
	case Deployment:
		pv1 = deployv1.New(mgr.GetClient())
	case Pod:
		pv1 = podv1.New(mgr.GetClient())
	case ConfigMap:
		pv1 = cmv1.New(mgr.GetClient())
	case CronJob:
		pv1 = cronjobv1.New(mgr.GetClient())
	case Daemonset:
		pv1 = dsv1.New(mgr.GetClient())
	case Ingress:
		pv1 = ingressv1.New(mgr.GetClient())
	case Job:
		pv1 = jobv1.New(mgr.GetClient())
	case Pvc:
		pv1 = pvcv1.New(mgr.GetClient())
	case Pv:
		pv1 = pvv1.New(mgr.GetClient())
	case Role:
		pv1 = rolev1.New(mgr.GetClient())
	case RoleBinding:
		pv1 = rolebv1.New(mgr.GetClient())
	case Secret:
		pv1 = secretv1.New(mgr.GetClient())
	case Service:
		pv1 = servicev1.New(mgr.GetClient())
	case ServiceAccount:
		pv1 = sav1.New(mgr.GetClient())
	case Statefulset:
		pv1 = stsv1.New(mgr.GetClient())
	default:
		c.String(404, "unsuported type variable")
		return
	}

	data, err := pv1.Get(namespace, name)
	if err != nil {
		utils.SendErrorMessage(c, 400, "error", err.Error())
		return
	}
	utils.SendSuccessMessage(c, 200, data)
}
