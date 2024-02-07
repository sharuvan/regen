package cmd

import (
	"fmt"
	"os"

	"github.com/sharuvan/regen/regen"
	"github.com/spf13/cobra"
)

var (
	ver                 bool
	file                string
	percentage          int
	checksumBlockLength int
	bruteforceLimit     int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "regen",
	Short: "Redundancy Generator",
	Long: `Regen is a data redundancy generator used to generate
an arbitrary amount of data for a given file and use
that to regenerate the file in case of integrity loss`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if ver {
			fmt.Printf("Regen %v\n", regen.PROGRAM_VERSION)
		}
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate hash and redundancy data",
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		err := regen.Generate(file, percentage, checksumBlockLength, true)
		if err != nil {
			fmt.Println(err)
		}
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify data integrity",
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		err := regen.Verify(file, true)
		if err != nil {
			fmt.Println(err)
		}
	},
}

var regenerateCmd = &cobra.Command{
	Use:   "regenerate",
	Short: "Regenerate archive using redundant data",
	// Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		// regenerate(file)
		err := regen.Regenerate(file, bruteforceLimit, true)
		if err != nil {
			fmt.Println(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(regenerateCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
	}
}

func init() {
	completion := completionCommand()
	completion.Hidden = true
	rootCmd.AddCommand(completion)
	// cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.regen.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVarP(&ver, "version", "v", false, "Show version information")
	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", "", "Archive file to work on")
	generateCmd.Flags().IntVarP(&percentage, "percentage", "p", 10, "Redundancy percentage")
	generateCmd.Flags().IntVarP(&checksumBlockLength, "checksum", "c", 64, "Checksum block length")
	regenerateCmd.Flags().IntVarP(&bruteforceLimit, "bruteforce-limit", "b", 1023, "Bruteforce limit")
}
