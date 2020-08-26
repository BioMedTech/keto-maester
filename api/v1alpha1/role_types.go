package v1alpha1

import (
	"fmt"
	"github.com/ory/keto-maester/keto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateId(named metav1.Object) string {
	return fmt.Sprintf("%s:%s", named.GetNamespace(), named.GetName())
}

// PolicyStatus defines the observed state of Policy
type RoleStatus struct {
	// ObservedGeneration represents the most recent generation observed by the daemon set controller.
	ObservedGeneration  int64               `json:"observedGeneration,omitempty"`
	ReconciliationError ReconciliationError `json:"reconciliationError,omitempty"`
}

type RoleSpec struct {
	// Members of role
	Members []string `json:"members,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleSpec   `json:"spec,omitempty"`
	Status RoleStatus `json:"status,omitempty"`
}

func (r *Role) SetObservedGeneration(generation int64) {
	r.Status.ObservedGeneration = generation
}

func (r *Role) SetReconciliationError(err ReconciliationError) {
	r.Status.ReconciliationError = err
}

// +kubebuilder:object:root=true

//RoleList contains a list of Role
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Role{}, &RoleList{})
}

func (r *Role) ToRoleJSON() *keto.Role {
	return &keto.Role{
		Id:      GenerateId(r),
		Members: r.Spec.Members,
	}
}
