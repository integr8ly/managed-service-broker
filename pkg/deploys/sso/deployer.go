package sso

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

const (
	DefaultManagedURLEnv = "SSO_URL"
	DefaultUserURLEnv    = "USER_SSO_URL"
)

type RHSSODeployer struct {
	URLEnv                  string
	CatalogServiceObjGetter func() []*brokerapi.Service
}

func NewClusterDeployer() *RHSSODeployer {
	return &RHSSODeployer{
		URLEnv:                  DefaultManagedURLEnv,
		CatalogServiceObjGetter: getManagedCatalogServicesObj,
	}
}

func NewUserDeployer() *RHSSODeployer {
	return &RHSSODeployer{
		URLEnv:                  DefaultUserURLEnv,
		CatalogServiceObjGetter: getUserCatalogServicesObj,
	}
}

func (d *RHSSODeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting RH-SSO managed catalog entries")
	return d.CatalogServiceObjGetter()
}

func (d *RHSSODeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying RH-SSO from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv(d.URLEnv)

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
		Operation:    "deploy",
	}, nil
}

func (d *RHSSODeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (d *RHSSODeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "RH-SSO deployed successfully",
	}, nil
}
