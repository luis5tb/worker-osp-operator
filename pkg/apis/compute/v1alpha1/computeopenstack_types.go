package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ComputeOpenStackSpec defines the desired state of ComputeOpenStack
type ComputeOpenStackSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// Name of the worker role created for OSP computes
	RoleName string `json:"roleName"`
	// Cluster name
	ClusterName string `json:"clusterName"`
	// Base Worker MachineSet Name
	BaseWorkerMachineSetName string `json:"baseWorkerMachineSetName"`
	// Kubernetes service cluster IP
	K8sServiceIp string `json:"k8sServiceIp"`
	// Internal Cluster API IP (app-int)
	ApiIntIp string `json:"apiIntIp"`
	// Number of workers
	Workers int32 `json:"workers,omitempty"`
	// Cores Pinning
	CorePinning string `json:"corePinning,omitempty"`
	// Infra DaemonSets needed
	InfraDaemonSets []InfraDaemonSet `json:"infraDaemonSets,omitempty"`
}

// InfraDaemonSet defines the daemon set required
type InfraDaemonSet struct {
	// Namespace
	Namespace string `json:"namespace"`
	// Name
	Name string `json:"name"`
}

// ComputeOpenStackStatus defines the observed state of ComputeOpenStack
type ComputeOpenStackStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// Number of workers
	Workers int32 `json:"workers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComputeOpenStack is the Schema for the computeopenstacks API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=computeopenstacks,scope=Namespaced
type ComputeOpenStack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComputeOpenStackSpec   `json:"spec,omitempty"`
	Status ComputeOpenStackStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComputeOpenStackList contains a list of ComputeOpenStack
type ComputeOpenStackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeOpenStack `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeOpenStack{}, &ComputeOpenStackList{})
}
