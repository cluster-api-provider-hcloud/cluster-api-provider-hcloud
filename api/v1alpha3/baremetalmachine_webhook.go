package v1alpha3

// TODO: Fix error that controller is not started if these webhooks are not commented out
/*
import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *BareMetalMachine) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}
func (r *BareMetalMachineList) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cluster-api-provider-hcloud-capihc-com-v1alpha3-baremetalmachine,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=cluster-api-provider-hcloud.capihc.com,resources=baremetalmachines,versions=v1alpha3,name=validation.baremetalmachine.cluster-api-provider-hcloud.capihc.com

var _ webhook.Validator = &BareMetalMachine{}

type bareMetalMachineSpecer interface {
	GroupVersionKind() schema.GroupVersionKind
	GetName() string
	BareMetalMachineSpec() *BareMetalMachineSpec
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *BareMetalMachine) ValidateCreate() error {
	return validateBareMetalMachineSpec(r)
}

func validateBareMetalMachineSpec(r bareMetalMachineSpecer) error {
	var allErrs field.ErrorList

	if len(*r.BareMetalMachineSpec().ServerType) == 0 {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "serverType"), *r.BareMetalMachineSpec().ServerType, "field cannot be empty"),
		)
	}

	if len(*r.BareMetalMachineSpec().ImagePath) == 0 {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "imagePath"), *r.BareMetalMachineSpec().ImagePath, "field cannot be empty"),
		)
	}

	if len(r.BareMetalMachineSpec().RobotTokenRef.TokenName) == 0 {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "robotTokenRef"), r.BareMetalMachineSpec().RobotTokenRef.TokenName, "field cannot be empty"),
		)
	}

	if len(r.BareMetalMachineSpec().SSHTokenRef.TokenName) == 0 {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "sshTokenRef"), r.BareMetalMachineSpec().SSHTokenRef.TokenName, "field cannot be empty"),
		)
	}

	return aggregateObjErrors(r.GroupVersionKind().GroupKind(), r.GetName(), allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *BareMetalMachine) ValidateUpdate(old runtime.Object) error {
	var allErrs field.ErrorList

	_, ok := old.(*BareMetalMachine)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected an BareMetalMachine but got a %T", old))
	}

	return aggregateObjErrors(r.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *BareMetalMachine) ValidateDelete() error {
	return nil
}
*/
