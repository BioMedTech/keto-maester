package controllers

import (
	"context"
	"github.com/ory/keto-maester/keto"

	"github.com/go-logr/logr"
	ketov1alpha1 "github.com/ory/keto-maester/api/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KetoClient interface {
	GetPolicy(flavour keto.Flavour, id string) (*keto.PolicyJSON, bool, error)
	ListPolicy(flavour keto.Flavour) ([]*keto.PolicyJSON, error)
	UpsertPolicy(flavour keto.Flavour, o *keto.PolicyJSON) (*keto.PolicyJSON, error)
	DeletePolicy(flavour keto.Flavour, id string) error

	GetRole(flavour keto.Flavour, id string) (*keto.Role, bool, error)
	ListRole(flavour keto.Flavour) ([]*keto.Role, error)
	UpsertRole(flavour keto.Flavour, o *keto.Role) (*keto.Role, error)
	DeleteRole(flavour keto.Flavour, id string) error
}

type Reconciler struct {
	KetoClient KetoClient
	Log        logr.Logger
	client.Client
}

// KetoPolicyReconciler reconciles a Policy object
type KetoPolicyReconciler struct {
	*Reconciler
}

func (r KetoPolicyReconciler) GetLog() logr.Logger {
	return r.Log
}
func (r KetoPolicyReconciler) GetResource() string {
	return "policy"
}

// +kubebuilder:rbac:groups=keto.ory.sh,resources=policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keto.ory.sh,resources=policies/status,verbs=get;update;patch

func (r *KetoPolicyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues(r.GetResource(), req.NamespacedName)

	var policy ketov1alpha1.Policy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		if apierrs.IsNotFound(err) {
			//if registerErr := r.removePolicies(ctx, &policy); registerErr != nil {
			//	return ctrl.Result{}, registerErr
			//}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if policy.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(policy.ObjectMeta.Finalizers, FinalizerName) {
			typeMeta := policy.TypeMeta
			policy.ObjectMeta.Finalizers = append(policy.ObjectMeta.Finalizers, FinalizerName)
			if err := r.Update(ctx, &policy); err != nil {
				return ctrl.Result{}, err
			}
			// restore the TypeMeta object as it is removed during Update, but need to be accessed later
			policy.TypeMeta = typeMeta
		}
	} else {
		// The object is being deleted
		if containsString(policy.ObjectMeta.Finalizers, FinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.removePolicies(ctx, &policy); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			policy.ObjectMeta.Finalizers = removeString(policy.ObjectMeta.Finalizers, FinalizerName)
			if err := r.Update(ctx, &policy); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if registerErr := r.upsertPolicy(ctx, &policy); registerErr != nil {
		return ctrl.Result{}, registerErr
	}

	return ctrl.Result{}, nil
}

func (r *KetoPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ketov1alpha1.Policy{}).
		Complete(r)
}

func (r *KetoPolicyReconciler) upsertPolicy(ctx context.Context, p *ketov1alpha1.Policy) error {
	_, exists, _ := r.KetoClient.GetPolicy(keto.Exact, ketov1alpha1.GenerateId(p))
	if exists && p.Generation == p.Status.ObservedGeneration {
		return nil
	}

	_, err := r.KetoClient.UpsertPolicy(keto.Flavour(p.Spec.PatternMatching), p.ToPolicyJSON())

	if err != nil {
		return updateReconciliationStatusError(ctx, r, p, err)
	}

	return ensureEmptyStatusError(ctx, r, p)
}

func (r *KetoPolicyReconciler) removePolicies(ctx context.Context, p *ketov1alpha1.Policy) error {

	// if a reqired field is empty, that means this is a delete after
	// the finalizers have done their job, so just return
	if p.Spec.Effect == "" && p.Spec.PatternMatching == "" {
		return nil
	}
	id := ketov1alpha1.GenerateId(p)
	_, exists, err := r.KetoClient.GetPolicy(keto.Flavour(p.Spec.PatternMatching), id)
	if err != nil {
		return err
	}

	if exists {
		if err := r.KetoClient.DeletePolicy(keto.Flavour(p.Spec.PatternMatching), id); err != nil {
			return err
		}
	}

	return nil
}
