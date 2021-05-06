//go:generate mapstructure-to-hcl2 -type Config

package lightsail

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

const BuilderId = "lightsail.builder"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	MockOption          string `mapstructure:"mock"`
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	err = config.Decode(&b.config, &config.DecodeOpts{
		PluginType:  "packer.builder.lightsail",
		Interpolate: true,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}
	// Return the placeholder for the generated data that will become available to provisioners and post-processors.
	// If the builder doesn't generate any data, just return an empty slice of string: []string{}
	buildGeneratedData := []string{""}
	return buildGeneratedData, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", *b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	staticCredentials := credentials.NewStaticCredentials(
		b.config.AccessKey,
		b.config.SecretKey,
		"",
	)
	state.Put("creds", *staticCredentials)

	ui.Say("Starting lightsail builder")

	steps := []multistep.Step{
		&StepKeyPair{DebugMode: b.config.PackerDebug, DebugKeyPath: fmt.Sprintf("ls_%s.pem",
			b.config.PackerBuildName), Comm: &b.config.Comm},
		new(StepCreateServer),
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      communicator.CommHost(b.config.Comm.Host(), "server_ip"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		new(common.StepProvision),
		new(StepCreateSnapshot),
		new(StepCloneSnapshot),
		&common.StepCleanupTempKeys{Comm: &b.config.Comm},
	}

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	startTime := time.Now()
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	if _, ok := state.GetOk("snapshots"); !ok {
		log.Println("Failed to find snapshots in state.")
		return nil, nil
	}

	ui.Say(fmt.Sprintf("Finished build in %f.2 min", time.Since(startTime).Minutes()))

	artifact := &Artifact{
		Name:        b.config.SnapshotName,
		RegionNames: b.config.Regions,
		creds:       staticCredentials,
		StateData:   map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	return artifact, nil
}
