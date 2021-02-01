package v1alpha4

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *HcloudMachine) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
func (r *HcloudMachineList) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cluster-api-provider-hcloud-capihc-com-v1alpha4-hcloudmachine,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=cluster-api-provider-hcloud.capihc.com,resources=hcloudmachines,versions=v1alpha4,name=validation.hcloudmachine.cluster-api-provider-hcloud.capihc.com

var _ webhook.Validator = &HcloudMachine{}

type hcloudMachineSpecer interface {
	GroupVersionKind() schema.GroupVersionKind
	GetName() string
	HcloudMachineSpec() *HcloudMachineSpec
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *HcloudMachine) ValidateCreate() error {
	return validateHcloudMachineSpec(r)
}

func validateHcloudMachineSpec(r hcloudMachineSpecer) error {
	var allErrs field.ErrorList

	if len(r.HcloudMachineSpec().Type) == 0 {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "type"), r.HcloudMachineSpec().Type, "field cannot be empty"),
		)
	}

	return aggregateObjErrors(r.GroupVersionKind().GroupKind(), r.GetName(), allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *HcloudMachine) ValidateUpdate(old runtime.Object) error {
	var allErrs field.ErrorList

	oldM, ok := old.(*HcloudMachine)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected an HcloudMachine but got a %T", old))
	}

	if r.Spec.Type != oldM.Spec.Type {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "type"), r.Spec.Type, "field is immutable"),
		)
	}

	return aggregateObjErrors(r.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *HcloudMachine) ValidateDelete() error {
	return nil
}
