package cmd

import (
	"github.com/mbark/punkt/db"
	"github.com/mbark/punkt/exec"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init the required directory structure and install dependencies",
	Long: `create the required directory structure and basic files needed
to make punkt work. Will also run ansible-galaxy to install dependencies for
punkt's ansible setup.`,
	Run: func(cmd *cobra.Command, args []string) {
		Initialize()
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}

// Initialize the necessary directory structure for a punkt by placing the
// basic ansible configuration places at the config directory.
func Initialize() {
	db.CreateStructure()

	exec.Run("ansible-galaxy", "install", "-r", "requirements.yml")
	exec.Run("ansible-playbook", "main.yml", "-i", "inventory", "-K")
}
