package registry

import (
	"strings"
	"testing"
)

// lxai-mentor-matching is a project name found on google
// since the validation happens after authentication

func TestLocationInValidationByProject(t *testing.T) {
	err := validateLocation("test", "test")

	if err == nil {
		t.Errorf("Expected error, but got none")
	} else {
		expected := "Failed to get locations from GCP: failed to get json from url: https://artifactregistry.googleapis.com/v1beta2/projects/test/locations?alt=json, status code: 403"
		if err.Error() != expected {
			t.Errorf("Expected %s, but got %s", expected, err.Error())
		}
	}
}

func TestLocationInValidationByLocation(t *testing.T) {
	err := validateLocation("lxai-mentor-matching", "test")

	if err == nil {
		t.Errorf("Expected error, but got none")
	} else {
		expected := "Invalid location 'test'. The valid locations are"
		if !strings.HasPrefix(err.Error(), expected) && len(err.Error()) > len(expected) {
			t.Errorf("Expected '%s' to have '%s' as prefix", err.Error(), expected)
		}
	}
}

func TestLocationValidation(t *testing.T) {
	err := validateLocation("lxai-mentor-matching", "us")

	if err != nil {
		t.Errorf("Expected no error, but got: %s", err.Error())
	}
}
