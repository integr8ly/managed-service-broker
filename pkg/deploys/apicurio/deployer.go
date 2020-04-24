package apicurio

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

type ApiCurioDeployer struct {
	id string
}

func NewDeployer() *ApiCurioDeployer {
	return &ApiCurioDeployer{}
}

func (ac *ApiCurioDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting apicurito catalog entries")
	return getCatalogServicesObj()
}

func (ac *ApiCurioDeployer) GetID() string {
	return ac.id
}

func (ac *ApiCurioDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying apicurito from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv("APICURIO_DASHBOARD_URL")

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (ac *ApiCurioDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (ac *ApiCurioDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "apicurito deployed successfully",
	}, nil
}
