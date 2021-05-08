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
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

// This is a definition of a builder step and should implement multistep.Step
type StepCreateServer struct {
	MockConfig string
}

// Run should execute the purpose of this step
func (s *StepCreateServer) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	creds := state.Get("creds").(credentials.Credentials)
	keyPairName := state.Get("keyPairName").(string)

	awsRegion := getCentralRegion(config.Regions[0])
	awsCfg := &aws.Config{
		Credentials: &creds,
		Region:      aws.String(awsRegion),
	}
	newSession, err := session.NewSession(awsCfg)
	if err != nil {
		err = fmt.Errorf("Failed setting up aws session: %v", newSession)
		return handleError(err, state)
	}
	lsClient := lightsail.New(newSession)

	ui.Say(fmt.Sprintf("Connected to AWS region -  \"%s\" ...", awsRegion))
	tempInstanceName := fmt.Sprintf("%s-%s", config.SnapshotName, uuid.TimeOrderedUUID())
	output, err := lsClient.CreateInstances(&lightsail.CreateInstancesInput{
		AvailabilityZone: aws.String(config.Regions[0]),
		BlueprintId:      aws.String(config.Blueprint),
		BundleId:         aws.String(config.BundleId),
		InstanceNames:    []*string{aws.String(tempInstanceName)},
		KeyPairName:      aws.String(keyPairName),
	})
	ui.Say(fmt.Sprintf("Data from AWS: %s\n", output.GoString()))
	ui.Say(fmt.Sprintf("Created lightsail instance -  \"%s\" ...", tempInstanceName))

	var lsInstance *lightsail.GetInstanceOutput
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ticker.C:
			lsInstance, err = lsClient.GetInstance(&lightsail.GetInstanceInput{InstanceName: aws.String(tempInstanceName)})
			if err != nil {
				err = fmt.Errorf("Failed creating instance: %w", err)
				return handleError(err, state)
			}
			state.Put("server_details", *lsInstance)
			if *lsInstance.Instance.State.Code != 16 {
				continue
			}
			break
		case <-ctx.Done():
			ticker.Stop()
			err = fmt.Errorf("Failed creating instance: %w", err)
			return handleError(ctx.Err(), state)
		}
		break
	}

	state.Put("server_details", *lsInstance)
	state.Put("server_ip", *lsInstance.Instance.PublicIpAddress)

	ui.Say(fmt.Sprintf("Deployed snapshot instance \"%s\" is now \"%s\" state", *lsInstance.Instance.Name,
		*lsInstance.Instance.State.Name))

	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepCreateServer) Cleanup(state multistep.StateBag) {
	rawDetails, isExist := state.GetOk("server_details")
	if !isExist {
		return
	}
	serverDetails := rawDetails.(lightsail.GetInstanceOutput)

	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	creds := state.Get("creds").(credentials.Credentials)

	awsCfg := &aws.Config{
		Credentials: &creds,
		Region:      aws.String(getCentralRegion(config.Regions[0])),
	}
	newSession, err := session.NewSession(awsCfg)
	if err != nil {
		ui.Say(fmt.Sprintf("Failed setting up aws session: %v", newSession))
		return
	}
	lsClient := lightsail.New(newSession)
	ui.Say(fmt.Sprintf("Deleting server \"%s\" ...", serverDetails))

	_, err = lsClient.DeleteInstance(&lightsail.DeleteInstanceInput{
		InstanceName: serverDetails.Instance.Name,
	})
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete server \"%s\": %s", *serverDetails.Instance.Name, err))
	}

}
