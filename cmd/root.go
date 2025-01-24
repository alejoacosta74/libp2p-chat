/*
Copyright © 2024 Alejo Acosta

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"os"

	"github.com/alejoacosta74/libp2p-chat-app/app"

	"github.com/alejoacosta74/go-logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chat",
	Short: "libp2p chat room",
	Long: `example project builds a chat room application using go-libp2p-pubsub. 
	The app runs in the terminal, and uses a text UI to show messages from other peers`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run:               run,
	PersistentPreRunE: preRun,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("nickname", "n", "", "nickname")
	rootCmd.Flags().StringP("room", "r", "", "chat room name")
	rootCmd.Flags().StringP("log", "l", "info", "log level")
	viper.BindPFlag("nickname", rootCmd.Flags().Lookup("nickname"))
	viper.BindPFlag("room", rootCmd.Flags().Lookup("room"))
	viper.BindPFlag("log", rootCmd.Flags().Lookup("log"))
}

func run(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := app.Run(ctx)
		if err != nil {
			logger.WithFields("error", err.Error()).Error("failed to run app")
		}
		cancel()
	}()
	<-ctx.Done()
	logger.Warn("shutdown complete")
}

func preRun(cmd *cobra.Command, args []string) error {
	logLevel := viper.GetString("log")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.SetLevel(logLevel)
	return nil
}
