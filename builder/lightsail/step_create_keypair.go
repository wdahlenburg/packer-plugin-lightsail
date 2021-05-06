package lightsail

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

// This is a definition of a builder step and should implement multistep.Step
type StepKeyPair struct {
	DebugMode    bool
	DebugKeyPath string
	Comm         *communicator.Config
}

// Run should execute the purpose of this step
func (s *StepCreateKeypair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	creds := state.Get("creds").(credentials.Credentials)

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

	tempSSHKeyName := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	state.Put("keyPairName", tempSSHKeyName) // default name for ssh step
	keyPairResp, err := lsClient.CreateKeyPair(&lightsail.CreateKeyPairInput{
		KeyPairName: aws.String(tempSSHKeyName),
		Tags:        nil,
	})
	if err != nil {
		err = fmt.Errorf("Failed creating key pair: %w", err)
		return handleError(err, state)
	}
	ui.Say(fmt.Sprintf("Created temporery key pair - %s", tempSSHKeyName))

	decodedPrivateKey := []byte(*keyPairResp.PrivateKeyBase64)
	s.Comm.SSHPrivateKey = decodedPrivateKey
	s.Comm.SSHPublicKey = []byte(*keyPairResp.PublicKeyBase64)

	state.Put("keypair", *keyPairResp) // default name for ssh step

	if s.DebugMode {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
		defer f.Close()
		if _, err := f.Write(decodedPrivateKey); err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				state.Put("error", fmt.Errorf("Error setting permissions of debug key: %s", err))
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepCreateKeypair) Cleanup(_ multistep.StateBag) {}
