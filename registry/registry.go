package registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

type RegistryConfig struct {
	GcpProject   string `hcl:"project"`
	Location     string `hcl:"version"`
	RepositoryId string `hcl:"repository_id"`
}

type Registry struct {
	config RegistryConfig
	log    hclog.Logger
}

// Implement Configurable
func (r *Registry) Config() (interface{}, error) {
	return &r.config, nil
}

// Implement ConfigurableNotify
func (r *Registry) ConfigSet(config interface{}) error {
	c, ok := config.(*RegistryConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		return fmt.Errorf("Expected *RegisterConfig as parameter")
	}

	r.log.Debug("Starting validations")
	// validate the config
	if c.GcpProject == "" {
		r.log.Debug("GcpProject is empty")
		return fmt.Errorf("project must be set")
	}

	if c.Location == "" {
		r.log.Debug("Location is empty")
		return fmt.Errorf("location must be set")
	}

	if c.RepositoryId == "" {
		r.log.Debug("RepositoryId is empty")
		return fmt.Errorf("repository_id must be set")
	}

	return nil
}

// Implement Registry
func (r *Registry) PushFunc() interface{} {
	// return a function which will be called by Waypoint
	return r.push
}

// A PushFunc does not have a strict signature, you can define the parameters
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
// In addition to default input parameters the builder.Binary from the Build step
// can also be injected.
//
// The output parameters for PushFunc must be a Struct which can
// be serialzied to Protocol Buffers binary format and an error.
// This Output Value will be made available for other functions
// as an input parameter.
// If an error is returned, Waypoint stops the execution flow and
// returns an error to the user.
func (r *Registry) push(ctx context.Context, ui terminal.UI) (*Artifact, error) {
	u := ui.Status()
	defer u.Close()
	u.Update("Pushing binary to registry")

	return &Artifact{}, nil
}

// Implement Authenticator
