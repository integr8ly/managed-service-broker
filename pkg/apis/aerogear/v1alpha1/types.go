package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type HasClusterServiceClass interface {
	GetClusterServiceClassExternalName() string
	GetClusterServiceClassName() string
	SetClusterServiceClassName(string)
}

type SharedServicePlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedServicePlan `json:"items"`
}

type SharedServicePlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServicePlanSpec   `json:"spec"`
	Status            SharedServicePlanStatus `json:"status"`
}

type SharedServicePlanStatus struct {
	Ready bool `json:"Ready"`
}

type SharedServicePlanSpec struct {
	ServiceType     string                      `json:"serviceType"`
	Name            string                      `json:"name"`
	ID              string                      `json:"id"`
	Description     string                      `json:"description"`
	Free            bool                        `json:"free"`
	BindParams      SharedServicePlanSpecParams `json:"bindParams"`
	ProvisionParams SharedServicePlanSpecParams `json:"provisionParams"`
}

type SharedServicePlanSpecParams struct {
	Schema     string                                         `json:"$schema"`
	Type       string                                         `json:"type"`
	Properties map[string]SharedServicePlanSpecParamsProperty `json:"properties"`
}

type SharedServicePlanSpecParamsProperty struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Title       string `json:"title"`
}

type SharedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedService `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServiceSpec   `json:"spec"`
	Status            SharedServiceStatus `json:"status,omitempty"`
}

type SharedServiceSpec struct {
	ServiceType                     string                 `json:"serviceType"`
	RequiredInstances               int                    `json:"requiredInstances"`
	MinimumInstances                int                    `json:"minInstances"`
	MaximumInstances                int                    `json:"maxInstances"`
	CurrentInstances                []string               `json:"currentInstances"`
	SlicesPerInstance               int                    `json:"slicesPerInstance"`
	Params                          map[string]interface{} `json:"params"`
	Image                           string                 `json:"image"`
	ClusterServiceClassExternalName string                 `json:"clusterServiceClassExternalName"`
	ClusterServiceClassName         string                 `json:"clusterServiceClassName"`
}
type SharedServiceStatus struct {
	Phase Phase `json:"phase,omitempty"`
	Ready bool  `json:"ready"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedServiceInstance `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServiceInstanceSpec   `json:"spec"`
	Status            SharedServiceInstanceStatus `json:"status,omitempty"`
}

func (s *SharedServiceInstance) GetClusterServiceClassExternalName() string {
	return s.Spec.ClusterServiceClassExternalName
}

func (s *SharedServiceInstance) GetClusterServiceClassName() string {
	return s.Spec.ClusterServiceClassName
}

func (s *SharedServiceInstance) SetClusterServiceClassName(csName string) {
	s.Spec.ClusterServiceClassName = csName
}

type SharedServiceInstanceSpec struct {
	//Image the docker image to run to provision the service
	Image                           string                 `json:"image"`
	ClusterServiceClassName         string                 `json:"clusterServiceClassName"`
	ClusterServiceClassExternalName string                 `json:"clusterServiceClassExternalName"`
	Params                          map[string]interface{} `json:"params"`
	SlicesPerInstance               int                    `json:"slicesPerInstance"`
}
type SharedServiceInstanceStatus struct {
	// Fill me
	Ready           bool     `json:"ready"`
	Status          string   `json:"status"` // provisioning, failed, provisioned
	Phase           Phase    `json:"phase"`
	ServiceInstance string   `json:"serviceInstance"`
	CurrentSlices   []string `json:"currentSlices"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedServiceSlice `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServiceSliceSpec   `json:"spec"`
	Status            SharedServiceSliceStatus `json:"status,omitempty"`
}

type SharedServiceSliceSpec struct {
	ServiceType    string                 `json:"serviceType"`
	Params         map[string]interface{} `json:"params"`
	SliceNamespace string                 `json:"sliceNamespace"`
	// Fill me
}
type SharedServiceSliceStatus struct {
	// Fill me
	Phase  Phase  `json:"phase"`
	Action string `json:"action"`
	Ready  bool   `json:"ready"`
	// the ServiceInstanceID that represents the slice
	SliceServiceInstance string `json:"sliceServiceInstance"`
	// the ServiceInstanceID that represents the parent shared service
	SharedServiceInstance string `json:"sharedServiceInstance"`
	// Human readable message about what is happening
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []SharedServiceClient `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SharedServiceClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              SharedServiceClientSpec   `json:"spec"`
	Status            SharedServiceClientStatus `json:"status,omitempty"`
}

type SharedServiceClientSpec struct {
	// Fill me
}
type SharedServiceClientStatus struct {
	// Fill me
}

type Phase string

var (
	NoPhase           Phase = ""
	AcceptedPhase     Phase = "accepted"
	ProvisioningPhase Phase = "provisioning"
	CompletePhase     Phase = "complete"
	FailedPhase       Phase = "failed"
)
