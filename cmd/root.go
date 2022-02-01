/*
Copyright © 2022 Scott Ware <scottdware@gmail.com>

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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/scottdware/go-easycsv"
	"github.com/spf13/viper"
	"gopkg.in/resty.v1"
)

type AWSIPRanges struct {
	SyncToken    string      `json:"syncToken"`
	CreateDate   string      `json:"createDate"`
	Prefixes     []AWSPrefix `json:"prefixes"`
	Ipv6Prefixes []AWSPrefix `json:"ipv6_prefixes"`
}

type AWSPrefix struct {
	Ipv6Prefix         string `json:"ipv6_prefix,omitempty"`
	Region             string `json:"region"`
	Service            string `json:"service"`
	NetworkBorderGroup string `json:"network_border_group"`
	IPPrefix           string `json:"ip_prefix,omitempty"`
}

type GoogleIPRanges struct {
	SyncToken    string         `json:"syncToken"`
	CreationTime string         `json:"creationTime"`
	Prefixes     []GooglePrefix `json:"prefixes"`
}

type GooglePrefix struct {
	Ipv4Prefix string `json:"ipv4Prefix,omitempty"`
	Ipv6Prefix string `json:"ipv6Prefix,omitempty"`
}

type AzureIPRanges struct {
	ChangeNumber int64   `json:"changeNumber"`
	Cloud        string  `json:"cloud"`
	Values       []Value `json:"values"`
}

type Value struct {
	Name       string     `json:"name"`
	ID         string     `json:"id"`
	Properties Properties `json:"properties"`
}

type Properties struct {
	ChangeNumber    int64    `json:"changeNumber"`
	Region          string   `json:"region"`
	RegionID        int64    `json:"regionId"`
	Platform        string   `json:"platform"`
	SystemService   string   `json:"systemService"`
	AddressPrefixes []string `json:"addressPrefixes"`
	NetworkFeatures []string `json:"networkFeatures"`
}

var (
	cfgFile   string
	vendor    string
	file      string
	iptype    int
	ipsources = map[string]string{
		"aws":    "https://ip-ranges.amazonaws.com/ip-ranges.json",
		"google": "https://www.gstatic.com/ipranges/goog.json",
		"azure":  "https://download.microsoft.com/download/7/1/D/71D86715-5596-4529-9B13-DA13A5DE5B63/ServiceTags_Public_20220124.json",
		// "azure-gov": "https://download.microsoft.com/download/6/4/D/64DB03BF-895B-4173-A8B1-BA4AD5D4DF22/ServiceTags_AzureGovernment_20210124.json",
	}
	awsip    AWSIPRanges
	googleip GoogleIPRanges
	azureip  AzureIPRanges
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudip",
	Short: "Program to fetch public IP ranges of AWS, Azure and Google",
	Long: `Get a list of all public IP address ranges (v4 or v6) for the three major cloud vendors:

Amazon AWS, Microsoft Azure and Google

By default, the ranges are printed to the console/screen. If you would like to save them in a file, the
output format is CSV, and you can use the "--file" flag to specify a file name.

Example:

	cloudip --vendor aws --iptype 4 --file AWS_IP_Ranges.csv`,
	Run: func(cmd *cobra.Command, args []string) {
		client := resty.New()

		switch vendor {
		case "aws":
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				Get(ipsources["aws"])

			if err != nil {
				fmt.Printf("unable to connect to AWS - %s", err)
			}

			if err := json.Unmarshal([]byte(resp.String()), &awsip); err != nil {
				fmt.Printf("JSON parse error on IP info - %s", err)
			}

			if len(file) > 0 {
				output, err := easycsv.NewCSV(fmt.Sprintf("%s", file))
				if err != nil {
					fmt.Printf("unable to create CSV file - %s", err)
				}

				output.Write("Prefix,Region,Service,Network Border Group\n")

				if iptype == 4 {
					for _, iprange := range awsip.Prefixes {
						output.Write(fmt.Sprintf("%s,%s,%s,%s\n", iprange.IPPrefix, iprange.Region, iprange.Service, iprange.NetworkBorderGroup))
					}
				}

				if iptype == 6 {
					for _, iprange6 := range awsip.Ipv6Prefixes {
						output.Write(fmt.Sprintf("%s,%s,%s,%s\n", iprange6.Ipv6Prefix, iprange6.Region, iprange6.Service, iprange6.NetworkBorderGroup))
					}
				}

				output.End()
			} else {
				if iptype == 4 {
					for _, iprange := range awsip.Prefixes {
						fmt.Printf("%s\n", iprange.IPPrefix)
					}
				}

				if iptype == 6 {
					for _, iprange6 := range awsip.Ipv6Prefixes {
						fmt.Printf("%s\n", iprange6.Ipv6Prefix)
					}
				}
			}
		case "google":
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				Get(ipsources["google"])

			if err != nil {
				fmt.Printf("unable to connect to Google - %s", err)
			}

			if err := json.Unmarshal([]byte(resp.String()), &googleip); err != nil {
				fmt.Printf("JSON parse error on IP info - %s", err)
			}

			if len(file) > 0 {
				output, err := easycsv.NewCSV(fmt.Sprintf("%s", file))
				if err != nil {
					fmt.Printf("unable to create CSV file - %s", err)
				}

				output.Write("Prefix\n")

				for _, iprange := range googleip.Prefixes {
					if iptype == 4 {
						if len(iprange.Ipv4Prefix) > 0 && len(iprange.Ipv6Prefix) <= 0 {
							output.Write(fmt.Sprintf("%s\n", iprange.Ipv4Prefix))
						}
					}

					if iptype == 6 {
						if len(iprange.Ipv4Prefix) <= 0 && len(iprange.Ipv6Prefix) > 0 {
							output.Write(fmt.Sprintf("%s\n", iprange.Ipv6Prefix))
						}
					}
				}

				output.End()
			} else {
				for _, iprange := range googleip.Prefixes {
					if iptype == 4 {
						if len(iprange.Ipv4Prefix) > 0 && len(iprange.Ipv6Prefix) <= 0 {
							fmt.Printf("%s\n", iprange.Ipv4Prefix)
						}
					}

					if iptype == 6 {
						if len(iprange.Ipv4Prefix) <= 0 && len(iprange.Ipv6Prefix) > 0 {
							fmt.Printf("%s\n", iprange.Ipv6Prefix)
						}
					}
				}
			}
		case "azure":
			resp, err := client.R().
				SetHeader("Content-Type", "application/json").
				Get(ipsources["azure"])

			if err != nil {
				fmt.Printf("unable to connect to Azure - %s", err)
			}

			if err := json.Unmarshal([]byte(resp.String()), &azureip); err != nil {
				fmt.Printf("JSON parse error on IP info - %s", err)
			}

			if len(file) > 0 {
				output, err := easycsv.NewCSV(fmt.Sprintf("%s", file))
				if err != nil {
					fmt.Printf("unable to create CSV file - %s", err)
				}

				output.Write("Name,ID,Change Number,Region,Region ID,Platform,System Service,Prefixes,Network Features\n")

				for _, iprange := range azureip.Values {
					ip4 := []string{}
					ip6 := []string{}

					for _, prefixes := range iprange.Properties.AddressPrefixes {
						if iptype == 4 {
							if IsIPv4(prefixes) {
								ip4 = append(ip4, prefixes)
							}
						}

						if iptype == 6 {
							if IsIPv6(prefixes) {
								ip6 = append(ip6, prefixes)
							}
						}
					}

					if iptype == 4 {
						output.Write(fmt.Sprintf("%s,%s,%d,%s,%d,%s,%s,\"%s\",\"%s\"\n", iprange.Name, iprange.ID, iprange.Properties.ChangeNumber,
							iprange.Properties.Region, iprange.Properties.RegionID, iprange.Properties.Platform, iprange.Properties.SystemService,
							sliceToString(ip4), sliceToString(iprange.Properties.NetworkFeatures)))
					}

					if iptype == 6 {
						output.Write(fmt.Sprintf("%s,%s,%d,%s,%d,%s,%s,\"%s\",\"%s\"\n", iprange.Name, iprange.ID, iprange.Properties.ChangeNumber,
							iprange.Properties.Region, iprange.Properties.RegionID, iprange.Properties.Platform, iprange.Properties.SystemService,
							sliceToString(ip6), sliceToString(iprange.Properties.NetworkFeatures)))
					}
				}

				output.End()
			} else {
				for _, iprange := range azureip.Values {
					// ip4 := []string{}
					// ip6 := []string{}

					for _, prefixes := range iprange.Properties.AddressPrefixes {
						if iptype == 4 {
							if IsIPv4(prefixes) {
								fmt.Printf("%s\n", prefixes)
								// ip4 = append(ip4, prefixes)
							}
						}

						if iptype == 6 {
							if IsIPv6(prefixes) {
								fmt.Printf("%s\n", prefixes)
								// ip6 = append(ip6, prefixes)
							}
						}
					}
				}
			}
		}
	},
}

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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cloudip.yaml)")
	rootCmd.Flags().StringVarP(&vendor, "vendor", "v", "", "Cloud vendor to export IP's from - aws|azure|google")
	rootCmd.Flags().IntVarP(&iptype, "iptype", "i", 4, "IP Type to export - 4|6|all")
	rootCmd.Flags().StringVarP(&file, "file", "f", "", "CSV filename to save the output to")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.MarkFlagRequired("vendor")
	rootCmd.MarkFlagRequired("iptype")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
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

		// Search config in home directory with name ".cloudip" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cloudip")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

func sliceToString(slice []string) string {
	var str string

	for _, item := range slice {
		str += fmt.Sprintf("%s, ", item)
	}

	return strings.TrimRight(str, ", ")
}

func stringToSlice(str string) []string {
	var slice []string

	list := strings.FieldsFunc(str, func(r rune) bool { return strings.ContainsRune(",;", r) })
	for _, item := range list {
		slice = append(slice, strings.TrimSpace(item))
	}

	return slice
}
