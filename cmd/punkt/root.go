package punkt

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/reconquest/loreley"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/kyokomi/emoji.v1"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/mgr"
)

var (
	logLevel   string
	configFile string
	punktHome  string
	dotfiles   string
)

var config *conf.Config
var snapshot *fs.Snapshot
var rootMgr mgr.RootManager

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "punkt",
	Short: emoji.Sprint(":package: punkt; a dotfile manager to be dotty about"),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.AddTemplateFunc("colorizeFlags", colorizeFlags)
	cobra.AddTemplateFunc("colorizeUseline", colorizeUseline)
	cobra.AddTemplateFunc("colorizeCommand", colorizeCommand)

	RootCmd.Version = "0.0.1"
	RootCmd.SetUsageTemplate(compileUsage())

	var err error
	snapshot, err = fs.NewSnapshot()
	if err != nil {
		logrus.WithError(err).Fatal("failed to create filesystem snapshot")
		os.Exit(1)
	}

	configFile = snapshot.ExpandHome("~/.config/punkt/config.toml")
	punktHome = snapshot.ExpandHome("~/.config/punkt")
	dotfiles = snapshot.ExpandHome("~/.dotfiles")

	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", configFile, `The configuration file to read custom configuration from`)
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	RootCmd.PersistentFlags().StringVarP(&punktHome, "punkt-home", "p", punktHome, `Where all punkt configuration files should be stored`)
	RootCmd.PersistentFlags().StringVarP(&dotfiles, "dotfiles", "d", dotfiles, `The directory containing the user's dotfiles`)

	var result error
	err = viper.BindPFlag("logLevel", RootCmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = viper.BindPFlag("punktHome", RootCmd.PersistentFlags().Lookup("punkt-home"))
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = viper.BindPFlag("dotfiles", RootCmd.PersistentFlags().Lookup("dotfiles"))
	if err != nil {
		result = multierror.Append(result, err)
	}

	if err != nil {
		logrus.WithError(result).Fatal("failed to bind flags to configuration")
	}
}

func initConfig() {
	var err error
	config, err = conf.NewConfig(*snapshot, snapshot.ExpandHome(configFile))
	if err != nil {
		logrus.WithError(err).Fatal("failed to red configuration file")
		os.Exit(1)
	}

	rootMgr = *mgr.NewRootManager(*config, *snapshot)
}

func compileUsage() string {

	withEmojis := emoji.Sprint(usageTemplate)
	return compileTemplate(withEmojis, map[string]interface{}{})
}

func compileTemplate(s string, data map[string]interface{}) string {
	usage, err := loreley.CompileAndExecuteToString(s, nil, data)
	if err != nil {
		logrus.WithError(err).Fatal("unable to compile string")
	}

	return usage
}

func pad(s string) string {
	if len(s) > 0 {
		return " " + s
	}

	return ""
}

func colorizeFlags(in string) string {
	colorized := []string{}

	for _, line := range strings.Split(in, "\n") {
		var pattern = regexp.MustCompile(`(-[a-zA-Z], )?--[\w-]+`)
		res := pattern.FindStringSubmatch(line)

		splits := strings.Split(line, res[0])
		format := "<.pre><fg 6><.flags><reset><.post>"
		withColor := compileTemplate(format, map[string]interface{}{
			"pre":   splits[0],
			"flags": res[0],
			"post":  pad(splits[1]),
		})

		colorized = append(colorized, withColor)
	}

	return strings.Join(colorized, "\n")
}

func colorizeUseline(in string) string {
	splits := strings.Split(in, " ")
	return compileTemplate("<fg 2><.root> <fg 5><.mid><fg 6><.flags><reset>", map[string]interface{}{
		"root":  splits[0],
		"mid":   strings.Join(splits[1:len(splits)-1], " "),
		"flags": pad(splits[len(splits)-1]),
	})
}

func colorizeCommand(in string) string {
	splits := strings.Split(in, " ")
	var rest string

	if len(splits) > 1 {
		rest = compileTemplate("<fg 5><.subcommand><reset><.rest>", map[string]interface{}{
			"subcommand": splits[1],
			"rest":       pad(strings.Join(splits[2:], " ")),
		})
	} else {
		rest = strings.Join(splits[1:], " ")
	}

	return compileTemplate("<fg 2><.root><reset><.rest>", map[string]interface{}{
		"root": splits[0],
		"rest": pad(rest),
	})
}

const usageTemplate = `<bold>:keyboard: Usage:<nobold>{{if .Runnable}}
  {{.UseLine | colorizeUseline}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath | colorizeCommand}} <fg 5>[command]<reset>{{end}}{{if gt (len .Aliases) 0}}

<bold>Aliases:<nobold>
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

<bold>:open_book: Examples:<nobold>
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

<bold>:wrench: Available Commands:<nobold>{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  <fg 5>{{rpad .Name .NamePadding }}<reset> {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

<bold>:white_flag: Flags:<nobold>
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces | colorizeFlags}}{{end}}{{if .HasAvailableInheritedFlags}}

<bold>:white_flag: Global Flags:<nobold>
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces | colorizeFlags}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath | colorizeCommand}} <fg 5>[command]<reset> <fg 6>--help<reset>" for more information about a command.{{end}}
`
