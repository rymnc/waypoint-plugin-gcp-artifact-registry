package registry

import (
	"context"
	"fmt"
	reflect "reflect"
	"strings"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (r *Registry) ValidateAuthFunc() interface{} {
	return r.validateAuth
}

//// AuthFunc satisfies the Authenticator interface
func (r *Registry) AuthFunc() interface{} {
	return r.authenticate
}

// A ValidateAuthFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// If an error is returned, Waypoint will attempt to call
// AuthFunc
func (r *Registry) validateAuth(
	ctx context.Context,
	ui terminal.UI,
) error {
	s := ui.Status()
	defer s.Close()
	s.Update("Authenticating with GCP")
	errString := "Failed to authenticate with GCP"

	artifactregistryService, err := artifactregistry.NewService(ctx)

	if err != nil {
		return fmt.Errorf(errString)
	}

	// The only permission required is to write
	expectedPermissions := []string{
		"roles/artifactregistry.writer",
	}

	testPerms := artifactregistry.TestIamPermissionsRequest{
		Permissions: expectedPermissions,
	}

	// Used https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/artifact_registry_repository#import
	// to pull the global resource identifier
	apiResource := fmt.Sprintf("projects/%s/locations/%s/repositories/%s",
		r.config.GcpProject,
		r.config.Location,
		r.config.RepositoryId,
	)

	s.Update("Testing Artifact Registry permissions...")

	// Testing IAM Permissions on GCP
	result, err := artifactregistryService.Projects.Locations.Repositories.TestIamPermissions(
		apiResource,
		&testPerms,
	).Do()

	if err != nil {
		s.Step(terminal.StatusError, "Error testing Artifact Registry permissions: "+err.Error())
		return err
	}

	// If resultant permissions do not equal expected permissions, invalidate
	// https://github.com/hashicorp/waypoint/blob/dde0897810c856a5d2cee977549c5cc911870441/builtin/google/cloudrun/platform.go#L141
	if !reflect.DeepEqual(result.Permissions, expectedPermissions) {
		s.Step(terminal.StatusError, "Incorrect IAM permissions, received "+strings.Join(result.Permissions, ", "))
		return status.Errorf(codes.PermissionDenied, "incorrect IAM permissions, received %s", strings.Join(result.Permissions, ", "))
	}

	r.client = artifactregistryService
	return nil
}

// A AuthFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// Output parameters must be *component.AuthResult, error
func (r *Registry) authenticate(
	ctx context.Context,
	ui terminal.UI,
) (*component.AuthResult, error) {

	ui.Output("Failed to authenticate with GCP")
	ui.Output("Please ensure that you have set the correct GCP Project and Location")
	ui.Output("You can find the correct values in the GCP Console")
	ui.Output("https://console.cloud.google.com/apis/credentials")
	ui.Output("")
	ui.Output("You can also set the GCP Project and Location in the Waypoint config file")
	ui.Output("")
	ui.Output("Please ensure that you have the correct permissions to write to the Artifact Registry")
	ui.Output("")

	return &component.AuthResult{Authenticated: false}, nil
}
