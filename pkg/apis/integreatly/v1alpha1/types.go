package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HasClusterServiceClass interface {
	GetClusterServiceClassExternalName() string
	GetClusterServiceClassName() string
	SetClusterServiceClassName(string)
}

type Phase string

const (
	CONSUMER_NAMESPACE_ANNOTATION = "namespace.integreatly.org"
	SLICE_ANNOTATION = "integreatly.org/managed-slice-for."
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ManagedServiceNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ManagedServiceNamespace `json:"items"`
}

// GetManagedNamespaceForConsumer returns the ManagedServiceNamespace for a consumer namespace
func (msnsl *ManagedServiceNamespaceList) GetManagedNamespaceForConsumer(ns string) *ManagedServiceNamespace {
	for	_, msns := range msnsl.Items {
		if contains(msns.Spec.ConsumerNamespaces, ns) {
			return &msns
		}
	}

	return nil
}

// GetMangedServiceNamespaceForInstance returns the ManagedServiceNamespace for a service instance
func (msnsl *ManagedServiceNamespaceList) GetMangedServiceNamespaceForInstance(id string) *ManagedServiceNamespace {
	for	_, msn := range msnsl.Items {
		if msn.Annotations[CONSUMER_NAMESPACE_ANNOTATION + "/" + id] != "" {
			return &msn
		}
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func NewManagedServiceNamespaceList() *ManagedServiceNamespaceList{
	return &ManagedServiceNamespaceList{
		TypeMeta: metav1.TypeMeta{
			Kind: "ManagedServiceNamespace",
			APIVersion: "integreatly.org/v1alpha1",
		},
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ManagedServiceNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ManagedServiceNamespaceSpec   `json:"spec"`
}

type ManagedServiceNamespaceSpec struct {
	metav1.TypeMeta                 `json:",inline"`
	metav1.ObjectMeta               `json:"metadata"`
	ConsumerNamespaces []string     `json:"consumerNamespaces"`
	UserID             string       `json:"userId"`
}

// SetInstanceNamespace Creates a mapping of service instance ids to consumer namespaces
func (msns *ManagedServiceNamespace) SetInstanceNamespace(id, ns string) {
	if msns.Annotations == nil {
		msns.Annotations = make(map[string]string)
	}

	msns.Annotations[CONSUMER_NAMESPACE_ANNOTATION + "/" + id] = ns
}

// GetConsumerInstanceNamespace Returns a consumer namespaces for a service instance id
func (msns *ManagedServiceNamespace) GetConsumerInstanceNamespace(id string) string{
	if msns.Annotations == nil {
		msns.Annotations = make(map[string]string)
	}

	return msns.Annotations[CONSUMER_NAMESPACE_ANNOTATION + "/" + id]
}