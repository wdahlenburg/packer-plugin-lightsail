package lightsail

import (
	"log"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`

	ctx          interpolate.Context
	SnapshotName string   `mapstructure:"snapshot_name" required:"true"`
	Regions      []string `mapstructure:"regions" required:"true"`
	BundleId     string   `mapstructure:"bundle_id" required:"true"`
	Blueprint    string   `mapstructure:"blueprint_id" required:"true"`

	AccessKey string `mapstructure:"access_key" required:"true"`
	SecretKey string `mapstructure:"secret_key" required:"true"`

	Timeout time.Duration `mapstructure:"timeout"`
}

func NewConfig(raws ...interface{}) (*Config, error) {
	c := new(Config)
	log.Printf("Loading configuration")
	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	c.Comm.Type = "ssh"
	c.Comm.SSHUsername = "ubuntu"

	c.Comm.Prepare(&c.ctx)

	if err != nil {
		return c, err
	}
	packer.LogSecretFilter.Set(c.SecretKey)
	packer.LogSecretFilter.Set(c.AccessKey)

	return c, nil
}

func getCentralRegion(region string) string {
	return region[:len(region)-1]
}
