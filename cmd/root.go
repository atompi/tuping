/*
Copyright © 2022 Atom Pi <coder.atompi@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"gitee.com/autom-studio/tuping/pkg/options"
	"gitee.com/autom-studio/tuping/pkg/tuping"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "tuping",
	Short:   "A network connectivity testing tool.",
	Long:    `A network connectivity testing tool. Supports testing via ICMP, TCP, and UDP.`,
	Version: options.Version,
	Args:    cobra.RangeArgs(1, 2),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		pingOptions := options.NewPingOptions(args)

		if pingOptions.Protocol == "icmp" {
			err := tuping.ICMPing(pingOptions)
			if err != nil {
				fmt.Fprintln(os.Stderr, "icmp ping failed: ", err)
				os.Exit(1)
			}
		} else if pingOptions.Protocol == "tcp" || pingOptions.Protocol == "udp" {
			if pingOptions.Port < 0 {
				fmt.Fprintln(os.Stderr, "the port is not correct, please provide a correct port")
				os.Exit(1)
			}
			l4pinger := tuping.NewL4Pinger(pingOptions)
			err := tuping.L4Ping(l4pinger)
			if err != nil {
				fmt.Fprintln(os.Stderr, "tcp/udp ping failed: ", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, "unknown protocol: ", pingOptions.Protocol)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringP("protocol", "p", "icmp", "specify protocol to use")
	rootCmd.PersistentFlags().IntP("count", "c", 0, "stop after <count> replies")
	rootCmd.PersistentFlags().IntP("size", "s", 64, "use <size> as number of data bytes to be sent")
	rootCmd.PersistentFlags().IntP("interval", "i", 1000, "millisecond between sending each packet")
	rootCmd.PersistentFlags().IntP("ttl", "t", 64, "define time to live")
	rootCmd.PersistentFlags().BoolP("wait", "w", false, "whether to wait for server response, should be set in udp mode only, default false")
	rootCmd.PersistentFlags().StringP("dns", "d", "", "specify the dns server instead of using the system default dns server")

	viper.BindPFlag("protocol", rootCmd.PersistentFlags().Lookup("protocol"))
	viper.BindPFlag("count", rootCmd.PersistentFlags().Lookup("count"))
	viper.BindPFlag("size", rootCmd.PersistentFlags().Lookup("size"))
	viper.BindPFlag("interval", rootCmd.PersistentFlags().Lookup("interval"))
	viper.BindPFlag("ttl", rootCmd.PersistentFlags().Lookup("ttl"))
	viper.BindPFlag("wait", rootCmd.PersistentFlags().Lookup("wait"))
	viper.BindPFlag("dns", rootCmd.PersistentFlags().Lookup("dns"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {}
