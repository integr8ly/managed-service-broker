package threescale

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	glog "github.com/sirupsen/logrus"
	"k8s.io/api/authentication/v1"
	"k8s.io/client-go/kubernetes"
)

type ThreeScaleDeployer struct {
	id string
}

func NewDeployer(id string) *ThreeScaleDeployer {
	return &ThreeScaleDeployer{id: id}
}

func (fd *ThreeScaleDeployer) IsForService(serviceID string) bool {
	return serviceID == "3scale-service-id"
}

func (fd *ThreeScaleDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting 3scale catalog entries")
	return getCatalogServicesObj()
}

func (fd *ThreeScaleDeployer) GetID() string {
	return fd.id
}

func (fd *ThreeScaleDeployer) Deploy(instanceID, brokerNamespace string, contextProfile brokerapi.ContextProfile, parameters map[string]interface{}, userInfo v1.UserInfo, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Infof("Deploying 3scale from deployer, id: %s", instanceID)

	dashboardUrl := os.Getenv("THREESCALE_DASHBOARD_URL")

	return &brokerapi.CreateServiceInstanceResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (fd *ThreeScaleDeployer) RemoveDeploy(serviceInstanceId string, namespace string, k8sclient kubernetes.Interface) error {
	return nil
}

func (fd *ThreeScaleDeployer) LastOperation(instanceID string, k8sclient kubernetes.Interface, osclient *openshift.ClientFactory, operation string) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", instanceID)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "3scale deployed successfully",
	}, nil
}
