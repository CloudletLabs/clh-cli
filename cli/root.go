package cli

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var home string

var rootCli = &cobra.Command{
	Use:   "clh",
	Short: "clh is a CloudletHub CLI tool",
	Long: `CloudletHub is a Continous Delivery as a Service,
		the only CD you ever need.
		Complete documentation is available at https://cloudlethub.com/docs`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var versionCli = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of clh",
	Long:  "All software has versions. We have it too",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("clh v0.1 -- HEAD")
	},
}

var useContextCli = &cobra.Command{
	Use:   "use-context",
	Short: "Switch to another context and save it as default",
	Long:  "Use provided context as default",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			viper.Set("context", args[0])
		}
		saveConfig()
	},
}

var configCli = &cobra.Command{
	Use:   "config",
	Short: "Configure clh",
	Long:  `Helps configuring clh tool such as Hub address and credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		saveConfig()
	},
}

func init() {
	// Root

	rootCli.PersistentFlags().StringP("log_level", "l", "", "Level for logs")
	viper.BindPFlag("log_level", rootCli.PersistentFlags().Lookup("log_level"))
	viper.SetDefault("log_level", "info")

	rootCli.PersistentFlags().StringP("config", "", "", "Path to a config file")
	viper.BindPFlag("config", rootCli.PersistentFlags().Lookup("config"))

	rootCli.PersistentFlags().StringP("context", "c", "", "CLH context name")
	viper.BindPFlag("context", rootCli.PersistentFlags().Lookup("context"))
	viper.SetDefault("context", "default")

	// When root core arguments is defined - read environment and configs
	viperFirstPhase()

	// Version

	rootCli.AddCommand(versionCli)

	// Use Context

	rootCli.AddCommand(useContextCli)

	// Config

	configCli.PersistentFlags().StringP("endpoint", "e", "", "CLH address")

	configCli.PersistentFlags().StringP("username", "u", "", "CLH username")

	configCli.PersistentFlags().StringP("secret_key", "k", "", "CLH Secret Key ID")

	rootCli.AddCommand(configCli)

	// Finish with cobra - set context and read custom config
	cobra.OnInitialize(cobraSecondPhase)
}

func viperFirstPhase() {
	viper.SetEnvPrefix("CLH")
	viper.AutomaticEnv()

	// First: at least consider environment variables
	setLogLevel()

	h, err := homedir.Dir()
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
	home = h

	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/clh/")
	viper.AddConfigPath(home + "/.clh")
	viper.AddConfigPath("./.clh")
	viper.SetConfigName("config")

	if err := viper.MergeInConfig(); err != nil {
		log.Debug("Can't read config: ", err)
	}

	// Second: + standard config files
	setLogLevel()
}

func cobraSecondPhase() {
	// Third: + cli
	setLogLevel()

	cfgFile := viper.GetString("config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.MergeInConfig(); err != nil {
		log.Debug("Can't read config: ", err)
	}

	// Forth: + custom config file
	setLogLevel()

	// Bind and set defaults AFTER cobra is ready
	viperSecondPhase()
}

func viperSecondPhase() {
	context := viper.GetString("context")

	// Root

	if viper.ConfigFileUsed() != "" {
		viper.SetDefault("config", viper.ConfigFileUsed())
	} else {
		viper.SetDefault("config", home+"/.clh/config.yaml")
	}

	// Config

	viper.BindPFlag(context+".endpoint", configCli.PersistentFlags().Lookup("endpoint"))
	viper.SetDefault(context+".endpoint", "https://api.cloudlethub.com/")

	viper.BindPFlag(context+".username", configCli.PersistentFlags().Lookup("username"))

	viper.BindPFlag(context+".secret_key", configCli.PersistentFlags().Lookup("secret_key"))
}

func setLogLevel() {
	ll, err := log.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		ll = log.DebugLevel
		log.Error("Error in log level parsing, fall back to DEBUG: ", err)
	}
	log.SetLevel(ll)
}

func saveConfig() {
	fileName := viper.GetString("config")
	dirName := filepath.Dir(fileName)

	if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
		log.Panic("Can't create config directory: ", err)
		os.Exit(1)
	}

	// TODO: Some stuff needs to be filtered out before saving
	// Needs: https://github.com/spf13/viper/issues/632
	if err := viper.WriteConfigAs(fileName); err != nil {
		log.Panic("Can't save config: ", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCli.Execute(); err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}
