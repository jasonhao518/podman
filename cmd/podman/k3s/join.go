//go:build amd64 || arm64

package k3s

import (
	"fmt"
	"os"

	"github.com/containers/common/pkg/completion"
	"github.com/containers/common/pkg/strongunits"
	"github.com/containers/podman/v5/cmd/podman/registry"
	ldefine "github.com/containers/podman/v5/libpod/define"
	"github.com/containers/podman/v5/libpod/events"
	define "github.com/containers/podman/v5/pkg/k3s/define"
	"github.com/spf13/cobra"
)

var (
	joinCmd = &cobra.Command{
		Use:               "join [options] [NAME]",
		Short:             "join a k3s cluster",
		Long:              "join a k3s cluster",
		PersistentPreRunE: machinePreRunE,
		RunE:              joinMachine,
		Args:              cobra.MaximumNArgs(1),
		Example:           `podman k3s join podman-machine-default`,
		ValidArgsFunction: completion.AutocompleteNone,
	}

	joinOpts          = define.JoinOptions{}
	joinOptionalFlags = JoinOptionalFlags{}
)

// Flags which have a meaning when unspecified that differs from the flag default
type JoinOptionalFlags struct {
	UserModeNetworking bool
}

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: joinCmd,
		Parent:  k3sCmd,
	})

}

func joinMachine(cmd *cobra.Command, args []string) error {

	if !ldefine.NameRegex.MatchString(joinOpts.Username) {
		return fmt.Errorf("invalid username %q: %w", joinOpts.Username, ldefine.RegexError)
	}

	// check if a system connection already exists
	cons, err := registry.PodmanConfig().ContainersConfDefaultsRO.GetAllConnections()
	if err != nil {
		return err
	}
	for _, con := range cons {
		if con.ReadWrite {
			for _, connection := range []string{joinOpts.Name, fmt.Sprintf("%s-root", joinOpts.Name)} {
				if con.Name == connection {
					return fmt.Errorf("system connection %q already exists. consider a different machine name or remove the connection with `podman system connection rm`", connection)
				}
			}
		}
	}

	for idx, vol := range joinOpts.Volumes {
		joinOpts.Volumes[idx] = os.ExpandEnv(vol)
	}

	// Process optional flags (flags where unspecified / nil has meaning )
	if cmd.Flags().Changed("user-mode-networking") {
		joinOpts.UserModeNetworking = &joinOptionalFlags.UserModeNetworking
	}

	if cmd.Flags().Changed("memory") {
		if err := checkMaxMemory(strongunits.MiB(joinOpts.Memory)); err != nil {
			return err
		}
	}

	// TODO need to work this back in
	// if finished, err := vm.Init(joinOpts); err != nil || !finished {
	// 	// Finished = true,  err  = nil  -  Success! Log a message with further instructions
	// 	// Finished = false, err  = nil  -  The installation is partially complete and podman should
	// 	//                                  exit gracefully with no error and no success message.
	// 	//                                  Examples:
	// 	//                                  - a user has chosen to perform their own reboot
	// 	//                                  - reexec for limited admin operations, returning to parent
	// 	// Finished = *,     err != nil  -  Exit with an error message
	// 	return err
	// }

	newMachineEvent(events.Init, events.Event{Name: joinOpts.Name})
	fmt.Println("Machine init complete")

	extra := ""

	fmt.Printf("To start your machine run:\n\n\tpodman machine start%s\n\n", extra)
	return err
}
