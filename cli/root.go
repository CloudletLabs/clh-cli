package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var home string
var cfgFile string
var context string

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
		ctx := context
		if len(args) > 0 {
			ctx = args[0]
		}

		fileName := viper.GetString("config")

		file, err := os.OpenFile(fileName, os.O_RDONLY, 0600)
		if err != nil {
			log.Panic("Can't open config:", err)
			os.Exit(1)
		}

		yamlFile, err := ioutil.ReadAll(file)
		if err != nil {
			log.Panic("Can't read config:", err)
			os.Exit(1)
		}

		file.Close()

		cfg := map[string]interface{}{}
		if err := yaml.Unmarshal(yamlFile, cfg); err != nil {
			log.Panic("Can't parse config: ", err)
			os.Exit(1)
		}
		cfg["context"] = ctx

		marshal, err := yaml.Marshal(cfg)
		if err != nil {
			log.Panic("Can't marshal config: ", err)
			os.Exit(1)
		}

		file, err = os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			log.Panic("Can't open config:", err)
			os.Exit(1)
		}

		if _, err := file.WriteString(string(marshal[:])); err != nil {
			log.Panic("Can't parse config: ", err)
			os.Exit(1)
		}

		file.Close()
	},
}

var configCli = &cobra.Command{
	Use:   "config",
	Short: "Configure clh",
	Long:  `Helps configuring clh tool such as Hub address and credentials`,
	Run: func(cmd *cobra.Command, args []string) {
		fileName := viper.GetString("config")
		if err := viper.WriteConfigAs(fileName + ".test.yaml"); err != nil {
			log.Panic("Can't save config: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Root

	rootCli.PersistentFlags().StringP("log_level", "l", "", "Level for logs")
	viper.BindPFlag("log_level", rootCli.PersistentFlags().Lookup("log_level"))
	viper.SetDefault("log_level", "info")

	rootCli.PersistentFlags().StringVarP(&cfgFile, "config", "", "", "Path to a config file")
	viper.BindPFlag("config", rootCli.PersistentFlags().Lookup("config"))

	rootCli.PersistentFlags().StringVarP(&context, "context", "c", "", "CLH context name")
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

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.MergeInConfig(); err != nil {
		log.Debug("Can't read config: ", err)
	}

	// Forth: + custom config file
	setLogLevel()

	if context == "" {
		context = viper.GetString("context")
	}

	// Bind and set defaults AFTER cobra is ready
	viperSecondPhase()
}

func viperSecondPhase() {
	// Root

	if viper.ConfigFileUsed() != "" {
		viper.SetDefault("config", viper.ConfigFileUsed())
	} else {
		viper.SetDefault("config", home+"/.clh/config.yaml")
	}

	// Config

	viper.BindPFlag(context+".endpoint", configCli.PersistentFlags().Lookup("endpoint"))
	viper.SetDefault(context+".endpoint", "https://api.cloudlethub.com/")

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

func Execute() {
	if err := rootCli.Execute(); err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}
