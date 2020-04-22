package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/gjbae1212/grpc-vpn/auth"
	auth_aws_iam "github.com/gjbae1212/grpc-vpn/auth/aws_iam"
	auth_google_openid "github.com/gjbae1212/grpc-vpn/auth/google_openid"
	"github.com/gjbae1212/grpc-vpn/internal"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var (
	rootCmd = &cobra.Command{
		Use:   "vpn-server",
		Short: color.GreenString(`vpn-server is the VPN server communicating through GRPC.`),
		Long:  color.GreenString(`vpn-server is the VPN server communicating through GRPC.`),
	}

	defaultConfig config
)

type config struct {
	Port             string
	SubNet           string
	LogPath          string
	JwtSalt          string
	TlsCertification string
	TlsPem           string
	Auth             auth.ServerConfig
}

type commandRun func(cmd *cobra.Command, args []string)

// Execute runs main command.
func Execute() {
	rootCmd.Execute()
}

func initConfig() {
	cfgPath := viper.GetString("config")

	if cfgPath == "" {
		log.Panicln(color.RedString("[ERR] Not Found Config file"))
	}

	if err := setConfig(cfgPath); err != nil {
		log.Panicln(color.RedString("[ERR] setConfig %s", err))
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
				case "subnet":
					defaultConfig.SubNet = internal.InterfaceToString(v)
				case "log_path":
					defaultConfig.LogPath = internal.InterfaceToString(v)
				case "jwt_salt":
					defaultConfig.JwtSalt = internal.InterfaceToString(v)
				case "tls_certification":
					defaultConfig.TlsCertification = internal.InterfaceToString(v)
				case "tls_pem":
					defaultConfig.TlsPem = internal.InterfaceToString(v)
				default:
					return fmt.Errorf("[ERR] unknown config %s", k)
				}
			}
		case "auth":
			for k, v := range value.(map[interface{}]interface{}) {
				switch k {
				case "google_openid":
					defaultConfig.Auth.GoogleOpenId = &auth_google_openid.Config{}
					for kk, vv := range v.(map[interface{}]interface{}) {
						switch kk.(string) {
						case "client_id":
							defaultConfig.Auth.GoogleOpenId.ClientId = internal.InterfaceToString(vv)
						case "client_secret":
							defaultConfig.Auth.GoogleOpenId.ClientSecret = internal.InterfaceToString(vv)
						case "redirect_url":
							defaultConfig.Auth.GoogleOpenId.RedirectURL = internal.InterfaceToString(vv)
						default:
							return fmt.Errorf("[ERR] unknown config %s", kk)
						}
					}
				case "aws_iam":
					defaultConfig.Auth.AwsIAM = &auth_aws_iam.Config{}
					for kk, vv := range v.(map[interface{}]interface{}) {
						switch kk.(string) {
						case "access_key":
							defaultConfig.Auth.AwsIAM.AccessKey = internal.InterfaceToString(vv)
						case "secret_access_key":
							defaultConfig.Auth.AwsIAM.SecretAccessKey = internal.InterfaceToString(vv)
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
