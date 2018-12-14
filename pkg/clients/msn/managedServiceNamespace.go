package msn

import (
	apis "github.com/integr8ly/managed-service-broker/pkg/apis/integreatly/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

type ManagedServiceNamespaceClient struct {
	Namespace string
}

// GetManagedServiceNamespaces returns a list ManagedServiceNamespace for a namespace
func (msnsc *ManagedServiceNamespaceClient) GetManagedServiceNamespaces() (*apis.ManagedServiceNamespaceList, error){
	msnl := apis.NewManagedServiceNamespaceList()
	err := sdk.List(msnsc.Namespace, msnl)
	if err != nil {
		return nil, apiErrors.NewInternalError(err)
	}
	if len(msnl.Items) == 0 {
		return nil, errors.New("There are no ManagedServiceNamespace resources in " + msnsc.Namespace + " namespace")
	}

	return msnl, nil
}

// GetMangedServiceNamespaceForInstance returns the ManagedServiceNamespace for a consumer namespace
func (msnsc *ManagedServiceNamespaceClient) GetManagedServiceNamespaceForConsumer(ns string) (*apis.ManagedServiceNamespace, error){
	msnsList, err := msnsc.GetManagedServiceNamespaces();
	if err != nil {
		return nil, err
	}
	msns := msnsList.GetManagedNamespaceForConsumer(ns); if msns == nil {
		return nil, errors.New("Consumer namespace " + ns + " has not being stored")
	}

	return msns, nil
}

// GetMangedServiceNamespaceForInstance returns the ManagedServiceNamespace for a service instance
func (msnsc *ManagedServiceNamespaceClient) GetMangedServiceNamespaceForInstance(id string) (*apis.ManagedServiceNamespace, error){
	msnsList, err := msnsc.GetManagedServiceNamespaces();
	if err != nil {
		return nil, err
	}

	msns := msnsList.GetMangedServiceNamespaceForInstance(id); if msns == nil {
		return nil, errors.New("Instance " + id + " has not being stored")
	}

	return msns, nil
}