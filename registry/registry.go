package registry

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/hashicorp/waypoint/builtin/docker"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	wpdockerclient "github.com/hashicorp/waypoint/builtin/docker/client"
)

type RegistryConfig struct {
	Project      string `hcl:"project"`
	Location     string `hcl:"version"`
	RepositoryId string `hcl:"repository_id"`
}

type Registry struct {
	config RegistryConfig
	log    hclog.Logger
	client *artifactregistry.Service
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
	if c.Project == "" {
		r.log.Debug("Project is empty")
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

	return validateConfig(*c)
}

// Implement Registry
func (r *Registry) PushFunc() interface{} {
	// return a function which will be called by Waypoint
	return r.push
}

func (r *Registry) pushWithDocker(
	ctx context.Context,
	log hclog.Logger,
	ui terminal.UI,
	location,
	project,
	repositoryId string,
	source *docker.Image,
) error {
	_, _, err := ui.OutputWriters()
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "unable to create output for logs:%s", err)
	}

	sg := ui.StepGroup()
	defer sg.Wait()
	step := sg.Add("Initializing Docker client...")
	defer func() { step.Abort() }()

	cli, err := wpdockerclient.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "unable to create Docker client:%s", err)
	}
	cli.NegotiateAPIVersion(ctx)

	targetTag := fmt.Sprintf("%s-docker.pkg.dev/%s/%s:%s", location, project, repositoryId, source.Tag)
	step.Update("Tagging Docker image: %s => %s", source.Name(), targetTag)

	err = cli.ImageTag(ctx, source.Name(), targetTag)
	if err != nil {
		return status.Errorf(codes.Internal, "unable to tag image:%s", err)
	}

	step.Done()

	step = sg.Add("Pushing Docker image...")

	cmd := exec.Command("docker", "push", targetTag)
	output, err := cmd.Output()

	if err != nil {
		return status.Errorf(codes.Internal, "unable to push image:%s", err)
	}

	log.Debug(string(output))
	step.Update("Pushed Docker image: %s", targetTag)

	step.Done()
	return nil
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
func (r *Registry) push(ctx context.Context,
	img *docker.Image,
	ui terminal.UI,
	log hclog.Logger,
) (*Artifact, error) {
	u := ui.Status()
	defer u.Close()
	u.Update("Pushing image to registry")

	// push the image to the registry
	err := r.pushWithDocker(ctx, log, ui, r.config.Location, r.config.Project, r.config.RepositoryId, img)

	if err != nil {
		return nil, err
	}

	return &Artifact{}, nil
}

// Implement Authenticator
