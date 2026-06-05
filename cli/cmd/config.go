package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultPublisherURL  = "https://publisher.walrus-testnet.walrus.space"
	defaultAggregatorURL = "https://aggregator.walrus-testnet.walrus.space"
	defaultEpochs        = 10
	configDir            = ".wolllama"
	configFile           = "config"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Wolllama configuration",
	Long:  `View and modify Wolllama settings (Walrus endpoints, epochs).`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration key. Supported keys:
  publisher-url   Walrus Publisher endpoint
  aggregator-url  Walrus Aggregator endpoint
  epochs          Number of storage epochs (default: 10)

Examples:
  wolllama config set publisher-url https://publisher.walrus-testnet.walrus.space
  wolllama config set epochs 20`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func initConfig() (*viper.Viper, error) {
	v := viper.New()

	configPath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	v.SetConfigName(configFile)
	v.SetConfigType("yaml")
	v.AddConfigPath(filepath.Dir(configPath))

	// Defaults
	v.SetDefault("publisher_url", defaultPublisherURL)
	v.SetDefault("aggregator_url", defaultAggregatorURL)
	v.SetDefault("epochs", defaultEpochs)

	// Env var overrides
	v.SetEnvPrefix("WOLLLAMA")
	v.AutomaticEnv()

	// Read config file (ok if missing — use defaults)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	return v, nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	v, err := initConfig()
	if err != nil {
		return err
	}

	fmt.Printf("publisher_url:  %s\n", v.GetString("publisher_url"))
	fmt.Printf("aggregator_url: %s\n", v.GetString("aggregator_url"))
	fmt.Printf("epochs:         %d\n", v.GetInt("epochs"))

	if v.ConfigFileUsed() != "" {
		fmt.Printf("\nconfig file: %s\n", v.ConfigFileUsed())
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	v, err := initConfig()
	if err != nil {
		return err
	}

	validKeys := map[string]bool{
		"publisher_url":  true,
		"aggregator_url": true,
		"epochs":         true,
	}
	if !validKeys[key] {
		return fmt.Errorf("unknown config key %q (valid: publisher_url, aggregator_url, epochs)", key)
	}

	v.Set(key, value)

	configPath, err := configFilePath()
	if err != nil {
		return err
	}

	if err := v.WriteConfigAs(configPath); err != nil {
		// If the config file doesn't exist yet, create it
		if os.IsNotExist(err) {
			if err := v.SafeWriteConfigAs(configPath); err != nil {
				return fmt.Errorf("write config: %w", err)
			}
		} else {
			return fmt.Errorf("write config: %w", err)
		}
	}

	fmt.Printf("%s = %s\n", key, value)
	return nil
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, configDir, configFile+".yaml"), nil
}
