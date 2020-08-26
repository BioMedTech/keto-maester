package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	ketov1alpha1 "github.com/ory/keto-maester/api/v1alpha1"
	"github.com/ory/keto-maester/keto"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

type KetoRoleReconciler struct {
	*Reconciler
}

func (r KetoRoleReconciler) GetLog() logr.Logger {
	return r.Log
}
func (r KetoRoleReconciler) GetResource() string {
	return "role"
}

// +kubebuilder:rbac:groups=keto.ory.sh,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keto.ory.sh,resources=roles/status,verbs=get;update;patch

func (r *KetoRoleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues(r.GetResource(), req.NamespacedName)

	var role ketov1alpha1.Role
	if err := r.Get(ctx, req.NamespacedName, &role); err != nil {
		if apierrs.IsNotFound(err) {
			//if registerErr := r.removeRole(ctx, &role); registerErr != nil {
			//	return ctrl.Result{}, registerErr
			//}
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if role.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(role.ObjectMeta.Finalizers, FinalizerName) {
			typeMeta := role.TypeMeta
			role.ObjectMeta.Finalizers = append(role.ObjectMeta.Finalizers, FinalizerName)
			if err := r.Update(ctx, &role); err != nil {
				return ctrl.Result{}, err
			}
			// restore the TypeMeta object as it is removed during Update, but need to be accessed later
			role.TypeMeta = typeMeta
		}
	} else {
		// The object is being deleted
		if containsString(role.ObjectMeta.Finalizers, FinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.removeRole(ctx, &role); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			role.ObjectMeta.Finalizers = removeString(role.ObjectMeta.Finalizers, FinalizerName)
			if err := r.Update(ctx, &role); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if registerErr := r.upsertRole(ctx, &role); registerErr != nil {
		return ctrl.Result{}, registerErr
	}

	return ctrl.Result{}, nil
}

func (r *KetoRoleReconciler) removeRole(ctx context.Context, role *ketov1alpha1.Role) error {
	id := ketov1alpha1.GenerateId(role)
	_, exists, err := r.KetoClient.GetRole(keto.Exact, id)
	if err != nil {
		return err
	}
	if exists {
		return r.KetoClient.DeleteRole(keto.Exact, id)
	}

	return nil
}

func (r *KetoRoleReconciler) upsertRole(ctx context.Context, role *ketov1alpha1.Role) error {
	_, exists, _ := r.KetoClient.GetRole(keto.Exact, ketov1alpha1.GenerateId(role))
	if exists && role.Generation == role.Status.ObservedGeneration {
		return nil
	}

	_, err := r.KetoClient.UpsertRole(keto.Exact, role.ToRoleJSON())

	if err != nil {
		r.Log.Error(err, fmt.Sprintf("update failed for %s %s/%s ", r.GetResource(), role.GetName(), role.GetNamespace()), r.GetResource(), "update role")
		return updateReconciliationStatusError(ctx, r, role, err)
	}

	return ensureEmptyStatusError(ctx, r, role)
}

func (r *KetoRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ketov1alpha1.Role{}).
		Complete(r)
}
