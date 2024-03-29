package clientgo

import (
	"log/slog"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"flag"
	"path/filepath"
	"sync"

	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	once            sync.Once
	twice           sync.Once
	three           sync.Once
	clients         dynamic.Interface
	clientset       *kubernetes.Clientset
	discoveryclient *discovery.DiscoveryClient
	restConfig      *rest.Config
)

func InitClientDiscovery() *discovery.DiscoveryClient {
	var err error
	// 实现同时集群内外的支持
	// 便于本地调试
	three.Do(func() {
		slog.Info("start doInitDiscovery()")
		discoveryclient, err = doInitDiscovery()
		if err != nil {
			slog.Debug("init out of cluster error: %v", err)
			discoveryclient, err = doInitDiscoveryInner()
			if err != nil {
				// log.Fatal(err)
				panic(err)
			}
		}
	})
	return discoveryclient
}

func doInitDiscovery() (*discovery.DiscoveryClient, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfigdiscovery", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfigdiscovery", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	var err error
	// use the current context in kubeconfig
	restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	discoveryclient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return discoveryclient, nil
}

func doInitDiscoveryInner() (*discovery.DiscoveryClient, error) {
	var err error
	restConfig, err = rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func InitClient() *kubernetes.Clientset {
	var err error
	// 实现同时集群内外的支持
	// 便于本地调试
	twice.Do(func() {
		slog.Info("start doInit()")
		clientset, err = doInit()
		if err != nil {
			slog.Debug("init out of cluster error", "Error", err)
			clientset, err = doInitInner()
			if err != nil {
				// log.Fatal(err)
				panic(err)
			}
		}
	})
	return clientset
}

func InitClientDynamic() (dynamic.Interface, error) {
	var err error
	// 实现同时集群内外的支持
	// 便于本地调试
	once.Do(func() {
		slog.Info("start doInit()")
		clients, err = DoInitDynamic()
		if err != nil {
			slog.Debug("init out of cluster error", "Error", err.Error())
			clients, err = doInitInnerDynamic()
			if err != nil {
				slog.Error(err.Error())
			}
		}
	})
	return clients, nil
}

// 集群内client-go
func doInitInner() (*kubernetes.Clientset, error) {
	var err error
	restConfig, err = rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func doInitInnerDynamic() (dynamic.Interface, error) {
	var err error
	restConfig, err = rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func doInit() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	var err error
	// use the current context in kubeconfig
	restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func DoInitDynamic() (dynamic.Interface, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfigdynamic", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfigdynamic", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	var err error
	// use the current context in kubeconfig
	restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func InitRestClient() (*rest.Config, error, *corev1client.CoreV1Client) {
	// Instantiate loader for kubeconfig file.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	var err error
	// Get a rest.Config from the kubeconfig file.  This will be passed into all
	// the client objects we create.
	restConfig, err = kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	// Create a Kubernetes core/v1 client.
	coreclient, err := corev1client.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}
	return restConfig, err, coreclient
}

func RestConfig() *rest.Config {
	if restConfig == nil {
		InitClientDynamic()
	}
	return restConfig
}

var gvr = schema.GroupVersionResource{
	Group:    "traefik.containo.us",
	Version:  "v1alpha1",
	Resource: "middlewares",
}

func GetGVR(group, version, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
}

// func listCrontabs(client dynamic.Interface, namespace string) (*v1alpha1.MiddlewareList, error) {
// 	ctx := context.Background()
// 	list, err := client.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	data, err := list.MarshalJSON()
// 	if err != nil {
// 		return nil, err
// 	}
// 	var ctList v1alpha1.MiddlewareList
// 	if err := json.Unmarshal(data, &ctList); err != nil {
// 		return nil, err
// 	}
// 	return &ctList, nil
// }
