package config

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsUnderMaintenance(t *testing.T) {
	err := registerFlag(maintenanceFeatureFlagKey, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("isUnderMaintenance:", isUnderMaintenance())
}
func TestName(t *testing.T) {
	err := errors.New("hhh")
	t.Fatal(err)
}
