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
	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	glog "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
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

func getSharedResourceClient(namespace, kind string) (dynamic.ResourceInterface, error) {
	apiVersion := "aerogear.org/v1alpha1"
	sharedResourceClient, _, err := k8sclient.GetResourceClient(apiVersion, kind, namespace)
	return sharedResourceClient, err
}

func getSharedServiceSliceResourceClient(namespace string) (dynamic.ResourceInterface, error) {
	apiVersion := "aerogear.org/v1alpha1"
	kind := "SharedServiceSlice"
	sharedResourceClient, _, err := k8sclient.GetResourceClient(apiVersion, kind, namespace)
	return sharedResourceClient, err
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

	sharedServiceClient, err := getSharedResourceClient(namespace, "SharedService")
	if err != nil {
		return err
	}
	sharedServiceSliceClient, err := getSharedResourceClient(namespace, "SharedServiceSlice")
	if err != nil {
		return err
	}
	sharedServicePlanClient, err := getSharedResourceClient(namespace, "SharedServicePlan")
	if err != nil {
		return err
	}
	ctrlr := controller.CreateController(namespace, sharedServiceClient, sharedServiceSliceClient, sharedServicePlanClient)
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
