package cmd

import (
	"errors"
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "x-knife",
	Short: "X-Knife is a tool for the Twitter/X-API v2",
	Long:  `A command line tool to detect bots on Twitter using the Twitter API v2.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to X-Knife!")
	},
}

func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (Built on %s from Git SHA %s)", version, date, commit)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing X-Knife '%s'\n", err)
		os.Exit(1)
	}
}

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().String("username", "mkoertg", "X user account")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.xknife.yaml)")
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".xknife" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".xknife")
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found; ignore error if desired
		}
	}
}
