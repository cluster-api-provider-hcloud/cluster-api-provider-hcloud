package v1alpha4

import (
	"testing"
)

func TestHcloudMachine_ValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		machine *HcloudMachine
		wantErr bool
	}{
		{
			name: "type needs to be set",
			machine: &HcloudMachine{
				Spec: HcloudMachineSpec{},
			},
			wantErr: true,
		},
		{
			name: "type needs to be set",
			machine: &HcloudMachine{
				Spec: HcloudMachineSpec{
					Type: "x",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.machine.ValidateCreate(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHcloudMachine_ValidateUpdate(t *testing.T) {
	tests := []struct {
		name       string
		oldMachine *HcloudMachine
		newMachine *HcloudMachine
		wantErr    bool
	}{
		{
			name: "type is immutable",
			oldMachine: &HcloudMachine{
				Spec: HcloudMachineSpec{
					Type: "y",
				},
			},
			newMachine: &HcloudMachine{
				Spec: HcloudMachineSpec{
					Type: "x",
				},
			},
			wantErr: true,
		},
		{
			name: "type is immutable",
			oldMachine: &HcloudMachine{
				Spec: HcloudMachineSpec{
					Type: "x",
				},
			},
			newMachine: &HcloudMachine{
				Spec: HcloudMachineSpec{
					Type: "x",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.newMachine.ValidateUpdate(tt.oldMachine); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
