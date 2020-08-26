package v1alpha1

import (
	"encoding/json"
	"github.com/ory/keto-maester/keto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PolicySpec defines the desired state of Ory Keto Policy
type PolicySpec struct {
	// Define a way of rule matching(more info https://www.ory.sh/keto/docs/engines/acp-ory#pattern-matching-strategies)
	PatternMatching PatternMatching `json:"pattern_matching"`

	// Description is the human-readable string that describes permission
	Description string `json:"description,omitempty"`

	// Subjects for whom policies will applied to(for users: users:${username}, for groups: ${scope}:${group_name})
	Subjects []string `json:"subjects,omitempty"`

	// Defines actions (ex, read, write, etc)
	Actions []string `json:"actions"`

	// Allow or deny access
	Effect Action `json:"effect"`

	// Resources defines object which you want to restrict access to
	Resources []string `json:"resources"`

	// Condition when to apply policy(see https://www.ory.sh/keto/docs/engines/acp-ory#conditions for details)
	// +kubebuilder:validation:Type=object
	Conditions *runtime.RawExtension `json:"condition,omitempty"`
}

// +kubebuilder:validation:Enum=exact;regex;glob
// more info https://www.ory.sh/keto/docs/engines/acp-ory#pattern-matching-strategies
type PatternMatching string

// +kubebuilder:validation:Enum=allow;deny
type Action string

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	// ObservedGeneration represents the most recent generation observed by the daemon set controller.
	ObservedGeneration  int64               `json:"observedGeneration,omitempty"`
	ReconciliationError ReconciliationError `json:"reconciliationError,omitempty"`
}

// ReconciliationError represents an error that occurred during the reconciliation process
type ReconciliationError struct {
	// Description is the description of the reconciliation error
	Description string `json:"description,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Policy is the Schema for the keto policy API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

func (p *Policy) SetObservedGeneration(generation int64) {
	p.Status.ObservedGeneration = generation
}

func (p *Policy) SetReconciliationError(err ReconciliationError) {
	p.Status.ReconciliationError = err
}

// +kubebuilder:object:root=true

//PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}

// ToPolicyJSON converts an Policy into a PolicyJSON object that represents an Policy digestible by ORY Keto
func (p *Policy) ToPolicyJSON() *keto.PolicyJSON {
	conditions, _ := json.Marshal(p.Spec.Conditions)

	return &keto.PolicyJSON{
		Id:          GenerateId(p),
		Actions:     p.Spec.Actions,
		Conditions:  conditions,
		Description: p.Spec.Description,
		Effect:      string(p.Spec.Effect),
		Resources:   p.Spec.Resources,
		Subjects:    p.Spec.Subjects,
	}
}
