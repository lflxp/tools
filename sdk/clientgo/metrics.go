package clientgo

import (
	"flag"
	"path/filepath"

	log "github.com/go-eden/slf4go"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func InitClientMetrics() (client *versioned.Clientset, err error) {
	client, err = doMetricsInit()
	if err != nil {
		client, err = doMetricsInnerInit()
		log.Debugf("init out of cluster: %v", err)
		if err != nil {
			log.Fatal(err)
		}
	}
	return client, err
}

func doMetricsInnerInit() (client *versioned.Clientset, err error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return client, err
	}

	client = versioned.NewForConfigOrDie(restConfig)

	return
}

func doMetricsInit() (client *versioned.Clientset, err error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	client = versioned.NewForConfigOrDie(restConfig)
	return
}
