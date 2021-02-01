package v1alpha4

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *HcloudCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

func (r *HcloudClusterList) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cluster-api-provider-hcloud-capihc-com-v1alpha4-hcloudcluster,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=cluster-api-provider-hcloud.capihc.com,resources=hcloudclusters,versions=v1alpha4,name=validation.hcloudcluster.cluster-api-provider-hcloud.capihc.com

var _ webhook.Validator = &HcloudCluster{}

func (r *HcloudCluster) ValidateCreate() error {
	return nil
}

func (r *HcloudCluster) ValidateDelete() error {
	return nil
}

func (r *HcloudCluster) ValidateUpdate(old runtime.Object) error {
	var allErrs field.ErrorList

	oldC, ok := old.(*HcloudCluster)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected an HcloudCluster but got a %T", old))
	}

	oldLocations := sets.NewString()
	for _, l := range oldC.Spec.Locations {
		oldLocations.Insert(string(l))
	}
	newLocations := sets.NewString()
	for _, l := range r.Spec.Locations {
		newLocations.Insert(string(l))
	}

	if !oldLocations.Equal(newLocations) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "locations"), r.Spec.Locations, "field is immutable"),
		)
	}

	return aggregateObjErrors(r.GroupVersionKind().GroupKind(), r.Name, allErrs)
}
