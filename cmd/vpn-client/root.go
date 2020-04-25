package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gjbae1212/grpc-vpn/auth"
	"github.com/gjbae1212/grpc-vpn/internal"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var (
	rootCmd = &cobra.Command{
		Use:   "vpn-client",
		Short: color.GreenString(`vpn-client is the VPN client communicating through GRPC.`),
		Long:  color.GreenString(`vpn-client is the VPN client communicating through GRPC.`),
	}

	defaultConfig config
)

type config struct {
	Addr             string
	Port             string
	TlsCertification string
	Auth             auth.Config
}

type commandRun func(cmd *cobra.Command, args []string)

// Execute runs main command.
func Execute() {
	rootCmd.Execute()
}

func initConfig() {
	cfgPath := viper.GetString("config")

	if cfgPath == "" {
		log.Println(color.RedString("[ERR] Not Found Config file"))
		os.Exit(1)
	}

	if err := setConfig(cfgPath); err != nil {
		log.Println(color.RedString("[ERR] setConfig %s", err))
		os.Exit(1)
	}
}

func setConfig(cfgPath string) error {
	cfgAbsPath, err := filepath.Abs(cfgPath)
	if err != nil {
		return err
	}

	yml, err := ioutil.ReadFile(cfgAbsPath)
	if err != nil {
		return err
	}

	conf := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(yml, &conf); err != nil {
		return err
	}

	for name, value := range conf {
		switch name {
		case "vpn":
			for k, v := range value.(map[interface{}]interface{}) {
				switch k.(string) {
				case "port":
					defaultConfig.Port = internal.InterfaceToString(v)
				case "addr":
					defaultConfig.Addr = internal.InterfaceToString(v)
				case "tls_certification":
					defaultConfig.TlsCertification = internal.InterfaceToString(v)
				default:
					return fmt.Errorf("[ERR] unknown config %s", k)
				}
			}
		case "auth":
			for k, v := range value.(map[interface{}]interface{}) {
				switch k {
				case "google_openid":
					defaultConfig.Auth.GoogleOpenId = &auth.GoogleOpenIDConfig{}
					for kk, vv := range v.(map[interface{}]interface{}) {
						switch kk.(string) {
						case "client_id":
							defaultConfig.Auth.GoogleOpenId.ClientId = internal.InterfaceToString(vv)
						case "client_secret":
							defaultConfig.Auth.GoogleOpenId.ClientSecret = internal.InterfaceToString(vv)
						default:
							return fmt.Errorf("[ERR] unknown config %s", kk)
						}
					}
				case "aws_iam":
					defaultConfig.Auth.AwsIAM = &auth.AwsIamConfig{}
					for kk, vv := range v.(map[interface{}]interface{}) {
						switch kk.(string) {
						case "access_key":
							defaultConfig.Auth.AwsIAM.ClientAccessKey = internal.InterfaceToString(vv)
						case "secret_access_key":
							defaultConfig.Auth.AwsIAM.ClientSecretAccessKey = internal.InterfaceToString(vv)
						default:
							return fmt.Errorf("[ERR] unknown config %s", kk)
						}
					}
				default:
					return fmt.Errorf("[ERR] unknown config %s", k)
				}
			}
		default:
			return fmt.Errorf("[ERR] unknown config %s", name)
		}
	}

	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path(yaml)")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}
