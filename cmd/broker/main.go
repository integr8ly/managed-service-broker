package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"

	"github.com/aerogear/managed-services-broker/pkg/broker"
	"github.com/aerogear/managed-services-broker/pkg/broker/controller"
	"github.com/aerogear/managed-services-broker/pkg/broker/server"
	"github.com/aerogear/managed-services-broker/pkg/clients/openshift"
	"github.com/aerogear/managed-services-broker/pkg/deploys/fuse"
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/pkg/errors"
	glog "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

var options struct {
	Port    int
	TLSCert string
	TLSKey  string
}

func init() {
	flag.IntVar(&options.Port, "port", 8005, "use '--port' option to specify the port for broker to listen on")
	flag.StringVar(&options.TLSCert, "tlsCert", os.Getenv("TLS_CERT"), "base-64 encoded PEM block to use as the certificate for TLS. If '--tlsCert' is used, then '--tlsKey' must also be used. If '--tlsCert' is not used, then TLS will not be used.")
	flag.StringVar(&options.TLSKey, "tlsKey", os.Getenv("TLS_KEY"), "base-64 encoded PEM block to use as the private key matching the TLS certificate. If '--tlsKey' is used, then '--tlsCert' must also be used")
	flag.Parse()
}

func main() {
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		glog.Fatalln(err)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cancelOnInterrupt(ctx, cancelFunc)

	return runWithContext(ctx)
}

func runWithContext(ctx context.Context) error {
	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), broker.VERSION)
		return nil
	}
	if (options.TLSCert != "" || options.TLSKey != "") &&
		(options.TLSCert == "" || options.TLSKey == "") {
		fmt.Println("To use TLS, both --tlsCert and --tlsKey must be used")
		return nil
	}

	namespace := os.Getenv("POD_NAMESPACE")

	addr := ":" + strconv.Itoa(options.Port)
	var err error

	// Instantiate loader for kubeconfig file.
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	cfg, err := kubeconfig.ClientConfig()
	if err != nil {
		return errors.Wrap(err, "error creating kube client config")
	}
	k8sClient := k8sclient.GetKubeClient()
	osClient := openshift.NewClientFactory(cfg)

	ctrlr := controller.CreateController(namespace, k8sClient, osClient)

	ctrlr.RegisterDeployer(fuse.NewDeployer("fuse-deployer"))
	ctrlr.Catalog()

	if options.TLSCert == "" && options.TLSKey == "" {
		err = server.Run(ctx, addr, ctrlr)
	} else {
		err = server.RunTLS(ctx, addr, options.TLSCert, options.TLSKey, ctrlr)
	}
	return err
}

// cancelOnInterrupt calls f when os.Interrupt or SIGTERM is received.
// It ignores subsequent interrupts on purpose - program should exit correctly after the first signal.
func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			f()
		}
	}()
}
