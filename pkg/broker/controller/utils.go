package controller

import (
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/api/errors"
	apis "github.com/integr8ly/managed-service-broker/pkg/apis/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

type UserNamespaceMissingError struct {}

func (unsm UserNamespaceMissingError) Error() string {
	return "User namespace has not being set as consumerNamespace for this User"
}

type ServiceInstanceAnnotationMissingError struct {}

func (sima ServiceInstanceAnnotationMissingError) Error() string {
	return "User namespace has not being set as an annotation for this User"
}

func listManagedServiceNamespace(namespace string) (*apis.ManagedServiceNamespaceList, error){
	msnl := apis.NewManagedServiceNamespaceList()
	err := sdk.List(namespace, msnl, sdk.WithListOptions((&metav1.ListOptions{})))
	if err != nil {
		return nil, errors.NewInternalError(err)
	}

	return msnl, nil
}

func getMangedServiceNamespaceForUserNamespace(namespace, userNamespace string) (*apis.ManagedServiceNamespace, error){
	items, err := getMangedServiceNamespaceItems(namespace);if err != nil {
		return nil, err
	}

	for	_, msn := range items {
		if contains(msn.Spec.ConsumerNamespaces, userNamespace) {
			if msn.Annotations == nil {
				msn.Annotations = make(map[string]string)
			}
			return &msn, nil
		}
	}

	return nil, UserNamespaceMissingError{}
}

func getMangedServiceNamespaceItems(namespace string) ([]apis.ManagedServiceNamespace, error){
	msnl, err := listManagedServiceNamespace(namespace);if err != nil {
		return nil, err
	}

	return msnl.Items, nil
}

func getMangedServiceNamespaceForInstance(namespace, instanceID string) (*apis.ManagedServiceNamespace, error){
	items, err := getMangedServiceNamespaceItems(namespace);if err != nil {
		return nil, err
	}

	for	_, msn := range items {
		if msn.Annotations[getInstanceAnnotation(instanceID)] != "" {
			return &msn, nil
		}
	}

	return nil, ServiceInstanceAnnotationMissingError{}
}

func getInstanceAnnotation(instanceID string) string{
	return "namespace.integreatly.org/" + instanceID
}