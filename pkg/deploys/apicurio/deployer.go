package apicurio

import (
	"net/http"
	"os"

	"k8s.io/api/authentication/v1"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	glog "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type ApiCurioDeployer struct {
	id string
}

func NewDeployer(id string) *ApiCurioDeployer {
	return &ApiCurioDeployer{id: id}
}

func (ac *ApiCurioDeployer) IsForService(serviceID string) bool {
	return serviceID == "apicurio-service-id"
}

func (ac *ApiCurioDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting apicurio catalog entries")
	return getCatalogServicesObj()
}

func (ac *ApiCurioDeployer) GetID() string {
	return ac.id
}

func (ac *ApiCurioDeployer) Deploy(instanceID, brokerNamespace string, contextProfile brokerapi.ContextProfile, parameters map[string]interface{}, userInfo v1.UserInfo, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Infof("Deploying apicurio from deployer, id: %s", instanceID)

	dashboardUrl := os.Getenv("APICURIO_DASHBOARD_URL")

	return &brokerapi.CreateServiceInstanceResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (ac *ApiCurioDeployer) RemoveDeploy(serviceInstanceId string, namespace string, k8sclient kubernetes.Interface) error {
	return nil
}

func (ac *ApiCurioDeployer) LastOperation(instanceID string, k8sclient kubernetes.Interface, osclient *openshift.ClientFactory, operation string) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", instanceID)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "apicurio deployed successfully",
	}, nil
}
