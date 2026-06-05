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
	configDirName        = ".wolllama"
	configFileName       = "config"
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

// initViper configures the global viper instance with defaults, config file, and env vars.
// Called once from the root command's PersistentPreRunE.
func initViper() error {
	configPath, err := configFilePath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	viper.SetConfigName(configFileName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Dir(configPath))

	// Defaults
	viper.SetDefault("publisher_url", defaultPublisherURL)
	viper.SetDefault("aggregator_url", defaultAggregatorURL)
	viper.SetDefault("epochs", defaultEpochs)
	viper.SetDefault("ollama_path", "")

	// Env var overrides: WOLLLAMA_PUBLISHER_URL, WOLLLAMA_EPOCHS, etc.
	viper.SetEnvPrefix("WOLLLAMA")
	viper.AutomaticEnv()

	// Read config file (ok if missing — use defaults)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config: %w", err)
		}
	}

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	fmt.Printf("publisher_url:  %s\n", viper.GetString("publisher_url"))
	fmt.Printf("aggregator_url: %s\n", viper.GetString("aggregator_url"))
	fmt.Printf("epochs:         %d\n", viper.GetInt("epochs"))

	if viper.ConfigFileUsed() != "" {
		fmt.Printf("\nconfig file: %s\n", viper.ConfigFileUsed())
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	validKeys := map[string]bool{
		"publisher_url":  true,
		"aggregator_url": true,
		"epochs":         true,
	}
	if !validKeys[key] {
		return fmt.Errorf("unknown config key %q (valid: publisher_url, aggregator_url, epochs)", key)
	}

	viper.Set(key, value)

	configPath, err := configFilePath()
	if err != nil {
		return err
	}

	// viper.WriteConfig will fail if the file doesn't exist; SafeWriteConfig creates it
	if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
		if err := viper.SafeWriteConfigAs(configPath); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	} else {
		if err := viper.WriteConfigAs(configPath); err != nil {
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
	return filepath.Join(home, configDirName, configFileName+".yaml"), nil
}
