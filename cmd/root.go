package cmd

import (
	"context"
	"fmt"
	"goft/pkg/ftapi"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// API is used to interact with the 42 API
	API ftapi.APIInterface
	// Version the current used version
	Version = "development-build"
	token   *oauth2.Token
)

// NewRootCmd Create new root command
func NewRootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "goft",
		Short: "CLI tool to interact with 42's API",
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/goft/secret.yml)")
	cmd.Version = Version
	return &cmd
}

var rootCmd = NewRootCmd()

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		dir, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		dir = filepath.Join(dir, ".config", "goft")
		if err := os.MkdirAll(dir, 700); err != nil {
			os.Exit(1)
		}

		// Search config in home directory with name ".goft" (without extension).
		viper.AddConfigPath(dir)
		viper.SetConfigName("config")
	}

	// Load config from env variables if available
	viper.SetConfigType("yml")
	viper.AutomaticEnv()

	viper.SetDefault("auth_endpoint", "https://api.intra.42.fr/oauth/authorize")
	viper.SetDefault("token_endpoint", "https://api.intra.42.fr/oauth/token")
	viper.SetDefault("api_endpoint", "https://api.intra.42.fr/v2")
	viper.SetDefault("scopes", []string{"profile"})

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if cfgFile != "" {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}

	requiredConfigs := []string{
		"client_id",
		"client_secret",
		"redirect_uri",
	}

	for _, requiredConfig := range requiredConfigs {
		if viper.GetString(requiredConfig) == "" {
			_, _ = fmt.Fprintf(rootCmd.OutOrStderr(), "%s is required but not set in the config file\n", requiredConfig)
			os.Exit(1)
		}
	}

	config := &oauth2.Config{
		ClientID:     viper.GetString("client_id"),
		ClientSecret: viper.GetString("client_secret"),
		Scopes:       viper.GetStringSlice("scopes"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  viper.GetString("auth_endpoint"),
			TokenURL: viper.GetString("token_endpoint"),
		},
		RedirectURL: viper.GetString("redirect_uri"),
	}

	token := genToken(config)
	ctx := context.Background()
	client := config.Client(ctx, token)
	API = ftapi.New(viper.GetString("api_endpoint"), client)
}
