package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/mbark/punkt/brew"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/git"
	"github.com/mbark/punkt/symlink"
	"github.com/mbark/punkt/yarn"
)

var (
	magenta = color.New(color.FgMagenta).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	blue    = color.New(color.FgBlue).SprintFunc()
	cyan    = color.New(color.FgCyan).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	bold    = color.New(color.Bold).SprintFunc()
)

var (
	directories []string
	depth       int
	ignore      []string
)

var message = strings.TrimSpace(`
dump the current working environment to your dotfile configuration files. Doing
this will search for symlinks and add your currently installed packages to their
corresponding package-manager's config files.`)

// ensureCmd represents the ensure command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump your current environment to config files",
	Long:  message,
	Run: func(cmd *cobra.Command, args []string) {
		dump(cmd, args)
	},
}

func init() {
	dumpCmd.Flags().StringArrayVar(&directories, "symlink-directories", []string{"~"}, `Search the given directories for symlinks to add`)
	dumpCmd.Flags().IntVar(&depth, "depth", 2, `The depth to stop recursively searching for symlinks`)
	dumpCmd.Flags().StringArrayVar(&ignore, "ignore", []string{}, `Directories to ignore when searching for symlinks`)

	viper.BindPFlag("ignoredDirectories", dumpCmd.PersistentFlags().Lookup("ignore"))

	RootCmd.AddCommand(dumpCmd)
}

func dump(cmd *cobra.Command, args []string) {
	// dumpSymlinks(cmd, args)
	dumpHomebrew(cmd, args)
	dumpYarn(cmd, args)
	dumpGit(cmd, args)
}

func dumpSymlinks(cmd *cobra.Command, args []string) {
	symlinks := symlink.Dump(directories, depth, viper.GetStringSlice("ignore"))
	mapping := make(map[string]symlink.Symlink)
	options := []string{}

	for _, symlink := range symlinks {
		msg := fmt.Sprintf("%s to %s", magenta(symlink.From), cyan(symlink.To))
		options = append(options, msg)
		mapping[msg] = symlink
	}

	sort.Strings(options)

	selected := []string{}
	prompt := &survey.MultiSelect{
		Message:  "What symlinks should be stored in your dotfiles?",
		Help:     "Select the symlinks you want to add to your symlinks.yaml file",
		Options:  options,
		PageSize: 15,
	}
	survey.AskOne(prompt, &selected, nil)

	selectedSymlinks := []symlink.Symlink{}
	for _, msg := range selected {
		selectedSymlinks = append(selectedSymlinks, mapping[msg])
	}

	file.SaveYaml(selectedSymlinks, dotfiles, "symlinks")
}

func dumpHomebrew(cmd *cobra.Command, args []string) {
	brewfile := brew.Dump()
	addSymlink(brewfile, "")
}

func dumpYarn(cmd *cobra.Command, args []string) {
	files := yarn.Dump()
	for _, f := range files {
		addSymlink(f, "")
	}
}

func dumpGit(cmd *cobra.Command, args []string) {
	files, repos := git.Dump(punktHome)
	for _, f := range files {
		addSymlink(f, "")
	}

	file.SaveYaml(repos, dotfiles, "repos")
}
