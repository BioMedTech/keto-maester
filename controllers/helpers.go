package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	ketov1alpha1 "github.com/ory/keto-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FinalizerName = "finalizer.ory.keto.sh"
)

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

type ReconcilerInterface interface {
	GetLog() logr.Logger
	GetResource() string

	Status() client.StatusWriter
}

type WithStatus interface {
	SetReconciliationError(err ketov1alpha1.ReconciliationError)
	SetObservedGeneration(generation int64)

	GetGeneration() int64
	GetName() string
	GetNamespace() string
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() runtime.Object
}

func updateReconciliationStatusError(ctx context.Context, r ReconcilerInterface, obj WithStatus, err error) error {
	r.GetLog().Error(err, fmt.Sprintf("error processing %s %s/%s ", r.GetResource(), obj.GetName(), obj.GetNamespace()), r.GetResource(), "register")
	obj.SetReconciliationError(ketov1alpha1.ReconciliationError{
		Description: err.Error(),
	})

	return updateStatus(ctx, r, obj)
}

func updateStatus(ctx context.Context, r ReconcilerInterface, obj WithStatus) error {
	obj.SetObservedGeneration(obj.GetGeneration())

	if err := r.Status().Update(ctx, obj); err != nil {
		r.GetLog().Error(err, fmt.Sprintf("status update failed for %s %s/%s", r.GetResource(), obj.GetName(), obj.GetNamespace()), r.GetResource(), "update status")
		return err
	}
	return nil
}

func ensureEmptyStatusError(ctx context.Context, r ReconcilerInterface, obj WithStatus) error {
	obj.SetReconciliationError(ketov1alpha1.ReconciliationError{})
	return updateStatus(ctx, r, obj)
}
