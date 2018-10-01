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

var (
	NoPhase           Phase = ""
	AcceptedPhase     Phase = "accepted"
	ProvisioningPhase Phase = "provisioning"
	CompletePhase     Phase = "complete"
	FailedPhase       Phase = "failed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ManagedServiceNamespaceList struct {
	metav1.TypeMeta                           `json:",inline"`
	metav1.ListMeta                           `json:"metadata"`
	Items           []ManagedServiceNamespace `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ManagedServiceNamespace struct {
	metav1.TypeMeta                    `json:",inline"`
	metav1.ObjectMeta                  `json:"metadata"`
	Spec ManagedServiceNamespaceSpec   `json:"spec"`
}

type ManagedServiceNamespaceSpec struct {
	metav1.TypeMeta                 `json:",inline"`
	metav1.ObjectMeta               `json:"metadata"`
	ManagedNamespace string         `json:"managedNamespace"`
	ConsumerNamespaces []string     `json:"consumerNamespaces"`
}

func NewManagedServiceNamespaceList() *ManagedServiceNamespaceList{
	return &ManagedServiceNamespaceList{
		TypeMeta: metav1.TypeMeta{
			Kind: "ManagedServiceNamespace",
			APIVersion: "integreatly.org/v1alpha1",
		},
	}
}