package lightsail

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
)

// packersdk.Artifact implementation
type Artifact struct {
	Name        string
	RegionNames []string
	creds       *credentials.Credentials
	StateData   map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s:%s", strings.Join(a.RegionNames[:], ","), a.Name)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: '%v' in regions '%v'", a.Name, strings.Join(a.RegionNames[:], ","))
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	log.Printf("Deleting snapshot \"%s\"", a.Name)

	awsCfg := &aws.Config{
		Credentials: a.creds,
		Region:      &a.RegionNames[0],
	}
	newSession, err := session.NewSession(awsCfg)
	if err != nil {
		return fmt.Errorf("Failed setting up aws session: %v", newSession)
	}
	lsClient := lightsail.New(newSession)

	_, err = lsClient.DeleteInstanceSnapshot(&lightsail.DeleteInstanceSnapshotInput{
		InstanceSnapshotName: aws.String(a.Name),
	})
	if err != nil {
		return fmt.Errorf("Failed to delete snapshot \"%s\": %s", a.Name, err)
	}
	return nil
}
