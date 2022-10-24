package apicache

import (
	"github.com/lflxp/tools/sdk/apicache/server/cluster"
	"github.com/lflxp/tools/sdk/apicache/server/namespace"
	"github.com/lflxp/tools/sdk/apicache/server/swagger"

	"github.com/gin-gonic/gin"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = appv1.AddToScheme(scheme)

	// +kubebuilder:scaffold:scheme
}

func RunServer(mgr manager.Manager, router *gin.Engine) {
	namespace.RegisterNamespace(router, mgr)
	cluster.RegisterCluster(router, mgr)
	swagger.RegisterSwaggerMiddleware(router)
}
