package lightsail

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"golang.org/x/sync/errgroup"
)

// This is a definition of a builder step and should implement multistep.Step
type StepCloneSnapshot struct{}

// Run should execute the purpose of this step
func (s *StepCloneSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	creds := state.Get("creds").(credentials.Credentials)
	snapshot := state.Get("snapshot_details").(lightsail.InstanceSnapshot)

	if len(config.Regions) < 1 {
		return multistep.ActionContinue
	}
	ui.Say(fmt.Sprintf("Deploying snapshot \"%s\" into regions: %v", *snapshot.Name, config.Regions[1:]))
	errgrp, _ := errgroup.WithContext(ctx)
	var snapshots []lightsail.InstanceSnapshot
	for _, region := range config.Regions[1:] {
		region := region
		errgrp.Go(func() error {
			awsRegion := getCentralRegion(region)
			awsCfg := &aws.Config{
				Credentials: &creds,
				Region:      aws.String(awsRegion),
			}
			newSession, err := session.NewSession(awsCfg)
			if err != nil {
				err := fmt.Errorf("Failed setting up aws session: %v", newSession)
				return err
			}
			lsClient := lightsail.New(newSession)

			ui.Say(fmt.Sprintf("Connected to AWS region -  \"%s\" ...", awsRegion))
			ui.Say(fmt.Sprintf("Creating snapshot \"%s\" in  \"%s\" ..", config.SnapshotName, region))
			_, err = lsClient.CopySnapshot(&lightsail.CopySnapshotInput{
				SourceRegion:       snapshot.Location.RegionName,
				SourceSnapshotName: snapshot.Name,
				TargetSnapshotName: snapshot.Name,
			})
			if err != nil {
				err = fmt.Errorf("Failed cloning snapshot: %w", err)
				return err
			}
			ui.Say(fmt.Sprintf("Waiting for snapshot \"%s\" in  \"%s\" ..", config.SnapshotName, region))
			var snapshot *lightsail.GetInstanceSnapshotOutput
			ticker := time.NewTicker(5 * time.Second)
			for {
				select {
				case <-ticker.C:
					snapshot, err = lsClient.GetInstanceSnapshot(&lightsail.GetInstanceSnapshotInput{InstanceSnapshotName: aws.
						String(config.SnapshotName)})
					if err != nil {
						return err
					}
					if *snapshot.InstanceSnapshot.State != lightsail.InstanceSnapshotStateAvailable {
						continue
					}
					break
				case <-ctx.Done():
					ticker.Stop()
					return ctx.Err()
				}
				break
			}
			snapshots = append(snapshots, *snapshot.InstanceSnapshot)
			ui.Say(fmt.Sprintf("Deployed snapshot \"%s\" is now in \"%s\" state", *snapshot.InstanceSnapshot.Name,
				*snapshot.InstanceSnapshot.State))
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return handleError(err, state)
	}

	state.Put("snapshots", snapshots)

	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepCloneSnapshot) Cleanup(_ multistep.StateBag) {}
