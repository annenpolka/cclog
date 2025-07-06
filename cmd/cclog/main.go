package main

import (
	"fmt"
	"os"

	"cclog/internal/cli"
)

func main() {
	config, err := cli.ParseArgs(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nUse 'cclog -h' for help.\n")
		os.Exit(1)
	}

	if config.ShowHelp {
		fmt.Println(cli.GetHelpText())
		return
	}

	if config.TUIMode {
		selectedFile, err := cli.RunTUI(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "TUI Error: %v\n", err)
			os.Exit(1)
		}
		
		// Run cclog on the selected file
		newConfig := config
		newConfig.InputPath = selectedFile
		newConfig.TUIMode = false
		
		output, err := cli.RunCommand(newConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		// Print output
		if config.OutputPath == "" {
			fmt.Print(output)
		} else {
			fmt.Printf("Output written to: %s\n", config.OutputPath)
		}
		return
	}

	output, err := cli.RunCommand(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Only print to stdout if no output file was specified
	if config.OutputPath == "" {
		fmt.Print(output)
	} else {
		fmt.Printf("Output written to: %s\n", config.OutputPath)
	}
}