package launcher

import (
	"net/http"
	"os"

	brokerapi "github.com/aerogear/managed-services-broker/pkg/broker"
	"github.com/aerogear/managed-services-broker/pkg/clients/openshift"
	glog "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type LauncherDeployer struct {
	id string
}

func NewDeployer(id string) *LauncherDeployer {
	return &LauncherDeployer{id: id}
}

func (fd *LauncherDeployer) DoesDeploy(serviceID string) bool {
	return serviceID == "launcher-service-id"
}

func (fd *LauncherDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting launcher catalog entries")
	return getCatalogServicesObj()
}

func (fd *LauncherDeployer) GetID() string {
	return fd.id
}

func (fd *LauncherDeployer) Deploy(instanceID, brokerNamespace string, contextProfile brokerapi.ContextProfile, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Infof("Deploying launcher from deployer, id: %s", instanceID)

	dashboardUrl := os.Getenv("LAUNCHER_DASHBOARD_URL")

	return &brokerapi.CreateServiceInstanceResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (fd *LauncherDeployer) LastOperation(instanceID string, k8sclient kubernetes.Interface, osclient *openshift.ClientFactory) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", instanceID)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "launcher deployed successfully",
	}, nil
}
