package lightsail

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// Helper for easily bailing out of the execution flow
func handleError(err error, state multistep.StateBag) multistep.StepAction {
	state.Put("error", err)
	ui := state.Get("ui").(packer.Ui)
	ui.Error(err.Error())

	return multistep.ActionHalt
}
