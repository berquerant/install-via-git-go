package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	setConfigFlag(parseCmd)
	parseCmd.Flags().StringP("out", "o", "yaml", "Format [yaml, json]")
	rootCmd.AddCommand(parseCmd)
}

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse config file",
	Long:  `Parse config file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := parseConfigFromFlag(cmd)
		if err != nil {
			return err
		}

		outputFormat, _ := cmd.Flags().GetString("out")
		switch outputFormat {
		case "json":
			v, _ := json.Marshal(config)
			cmd.Println(string(v))
		case "yaml", "yml":
			v, _ := yaml.Marshal(config)
			cmd.Println(string(v))
		default:
			return fmt.Errorf("unknown format %s", outputFormat)
		}
		return nil
	},
}
