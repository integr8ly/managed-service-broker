package main

import (
	"context"
	"runtime"

	"net"
	"os"

	"github.com/aerogear/managed-services/pkg/operator/managed"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)

}

func main() {
	printVersion()
	k8sclient.GetKubeClient()
	cfg := mustNewKubeConfig()
	svcClient, err := sc.NewForConfig(cfg)
	if err != nil {
		logrus.Fatal("failed to get service catalog client ", err)
	}
	resource := "aerogear.org/v1alpha1"
	SharedServicekind := "SharedService"
	SharedServiceInstancekind := "SharedServiceInstance"
	SharedServiceSlicekind := "SharedServiceSlice"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}
	resyncPeriod := 5
	logrus.Infof("Watching %s, %s, %s, %d", resource, SharedServicekind, namespace, resyncPeriod)
	sdk.Watch(resource, SharedServiceInstancekind, namespace, resyncPeriod)
	sdk.Watch(resource, SharedServiceSlicekind, namespace, resyncPeriod)
	sdk.Watch(resource, SharedServicekind, namespace, resyncPeriod)
	k8client := k8sclient.GetKubeClient()

	resourceClient, _, err := k8sclient.GetResourceClient(resource, SharedServicekind, namespace)
	sdk.Handle(managed.NewHandler(k8client, resourceClient, namespace, svcClient))
	sdk.Run(context.TODO())
}

// mustNewKubeClientAndConfig returns the in-cluster config and kubernetes client
// or if KUBERNETES_CONFIG is given an out of cluster config and client
func mustNewKubeConfig() *rest.Config {
	var cfg *rest.Config
	var err error
	if os.Getenv(k8sutil.KubeConfigEnvVar) != "" {
		cfg, err = outOfClusterConfig()
	} else {
		cfg, err = inClusterConfig()
	}
	if err != nil {
		panic(err)
	}
	return cfg
}

// inClusterConfig returns the in-cluster config accessible inside a pod
func inClusterConfig() (*rest.Config, error) {
	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// See https://github.com/coreos/etcd-operator/issues/731#issuecomment-283804819
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			return nil, err
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	return rest.InClusterConfig()
}

func outOfClusterConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv(k8sutil.KubeConfigEnvVar)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	return config, err
}
