package launcher

import (
	"net/http"
	"os"

	"k8s.io/api/authentication/v1"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	glog "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type LauncherDeployer struct {
	id string
}

func NewDeployer(id string) *LauncherDeployer {
	return &LauncherDeployer{id: id}
}

func (ld *LauncherDeployer) IsForService(serviceID string) bool {
	return serviceID == "launcher-service-id"
}

func (ld *LauncherDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting launcher catalog entries")
	return getCatalogServicesObj()
}

func (ld *LauncherDeployer) GetID() string {
	return ld.id
}

func (ld *LauncherDeployer) Deploy(instanceID, brokerNamespace string, contextProfile brokerapi.ContextProfile, parameters map[string]interface{}, userInfo v1.UserInfo, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Infof("Deploying launcher from deployer, id: %s", instanceID)

	dashboardUrl := os.Getenv("LAUNCHER_DASHBOARD_URL")

	return &brokerapi.CreateServiceInstanceResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (ld *LauncherDeployer) RemoveDeploy(serviceInstanceId string, namespace string, k8sclient kubernetes.Interface) error {
	return nil
}

func (ld *LauncherDeployer) LastOperation(instanceID string, k8sclient kubernetes.Interface, osclient *openshift.ClientFactory, operation string) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", instanceID)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "launcher deployed successfully",
	}, nil
}
