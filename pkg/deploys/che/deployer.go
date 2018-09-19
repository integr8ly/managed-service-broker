package che

import (
	"k8s.io/api/authentication/v1"
	"net/http"
	"os"

	brokerapi "github.com/aerogear/managed-services-broker/pkg/broker"
	"github.com/aerogear/managed-services-broker/pkg/clients/openshift"
	glog "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type CheDeployer struct {
	id string
}

func NewDeployer(id string) *CheDeployer {
	return &CheDeployer{id: id}
}

func (fd *CheDeployer) DoesDeploy(serviceID string) bool {
	return serviceID == "che-service-id"
}

func (fd *CheDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting che catalog entries")
	return getCatalogServicesObj()
}

func (fd *CheDeployer) GetID() string {
	return fd.id
}

func (fd *CheDeployer) Deploy(instanceID, brokerNamespace string, contextProfile brokerapi.ContextProfile, userInfo v1.UserInfo, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Infof("Deploying che from deployer, id: %s", instanceID)

	dashboardUrl := os.Getenv("CHE_DASHBOARD_URL")

	return &brokerapi.CreateServiceInstanceResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (fd *CheDeployer) LastOperation(instanceID string, k8sclient kubernetes.Interface, osclient *openshift.ClientFactory) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", instanceID)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "che deployed successfully",
	}, nil
}
