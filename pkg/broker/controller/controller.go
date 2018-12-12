/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"sync"

	"net/http"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	glog "github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
)

//Deployer deploys a service from this broker
type Deployer interface {
	GetCatalogEntries() []*brokerapi.Service
	Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error)
	ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error)
	RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error)
}

// Controller defines the APIs that all controllers are expected to support. Implementations
// should be concurrency-safe
type Controller interface {
	Catalog() (*brokerapi.Catalog, error)

	ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error)
	Provision(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error)
	Deprovision(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error)

	Bind(req *brokerapi.BindRequest, async bool) (*brokerapi.BindResponse, error)
	UnBind(req *brokerapi.UnBindRequest, async bool) (*brokerapi.UnBindResponse, error)
}

type errNoSuchInstance struct {
	instanceID string
}

func (e errNoSuchInstance) Error() string {
	return fmt.Sprintf("no such instance with ID %s", e.instanceID)
}

type userProvidedServiceInstance struct {
	Name       string
	Credential *brokerapi.Credential
}

type userProvidedController struct {
	rwMutex             sync.RWMutex
	instanceMap         map[string]*userProvidedServiceInstance
	k8sclient           kubernetes.Interface
	osClientFactory     *openshift.ClientFactory
	brokerNS            string
	deployers           []Deployer
}

// CreateController creates an instance of a User Provided service broker controller.
func CreateController(ds []Deployer) Controller {
	var instanceMap = make(map[string]*userProvidedServiceInstance)
	return &userProvidedController{
		instanceMap:         instanceMap,
		deployers:           ds,
	}
}

var services []*brokerapi.Service

func (c *userProvidedController) Catalog() (*brokerapi.Catalog, error) {
	glog.Info("Catalog()")

	services = []*brokerapi.Service{}
	for _, deployer := range c.deployers {
		services = append(services, deployer.GetCatalogEntries()...)
	}

	return &brokerapi.Catalog{
		Services: services,
	}, nil
}

func (c *userProvidedController) Provision(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Create service instance: %s, user: %s", req.ServiceID, req.OriginatingUserInfo.Username)

	deployer := c.getDeployer(req.ServiceID)
	if deployer != nil {
		return deployer.Deploy(req, async)
	}

	return &brokerapi.ProvisionResponse{Code: http.StatusInternalServerError}, nil
}

func (c *userProvidedController) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Info("ServiceInstanceLastOperation()", "operation "+req.Operation, req.ServiceID)

	deployer := c.getDeployer(req.ServiceID)
	if deployer != nil {
		return deployer.ServiceInstanceLastOperation(req)
	}

	return &brokerapi.LastOperationResponse{State: brokerapi.StateFailed, Description: "Could not find deployer for " + req.ServiceID}, nil
}

func (c *userProvidedController) Deprovision(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	glog.Info("Deprovision()", req.InstanceId)

	deployer := c.getDeployer(req.ServiceID)
	if deployer != nil {
		glog.Info("RemoveDeploy()", req.InstanceId)
		return deployer.RemoveDeploy(req, async)
	}

	return &brokerapi.DeprovisionResponse{}, nil
}

func (c *userProvidedController) Bind(req *brokerapi.BindRequest, async bool) (*brokerapi.BindResponse, error) {
	glog.Info("Bind()")

	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	instance, ok := c.instanceMap[req.InstanceId]
	if !ok {
		return nil, errNoSuchInstance{instanceID: req.InstanceId}
	}

	cred := instance.Credential
	return &brokerapi.BindResponse{Credentials: *cred}, nil
}

func (c *userProvidedController) UnBind(req *brokerapi.UnBindRequest, async bool) (*brokerapi.UnBindResponse, error) {
	glog.Info("UnBind()")
	// Since we don't persist the binding, there's nothing to do here.
	return nil, nil
}


func (c *userProvidedController) getDeployer(serviceId string) Deployer{
	for _, d := range c.deployers {
		if isForService(serviceId, d) {
			return d
		}
	}

	return nil
}

func isForService(serviceId string, d Deployer) bool {
	for _, s := range d.GetCatalogEntries() {
		if s.ID == serviceId {
			return true
	    }
	}

	return false
}
