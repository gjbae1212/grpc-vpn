package main

import (
	"github.com/gjbae1212/grpc-vpn/auth"
	"log"
	"os"
	"runtime"

	"github.com/gjbae1212/grpc-vpn/server"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:    "run",
		Short:  "Start vpn-server",
		Long:   "Start vpn-server",
		PreRun: startPreRun(),
		Run:    startRun(),
	}
)

func startPreRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		if runtime.GOOS != "linux" {
			log.Printf(color.RedString("`vpn-server` is only to support LINUX."))
			os.Exit(1)
		}

		if os.Getuid() != 0 {
			log.Printf("%s %s", color.RedString("[RETRY][COMMAND]"),
				color.CyanString("`sudo vpn-server run`"))
			os.Exit(1)
		}
	}
}

func startRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		// apply default params
		var opts []server.Option
		if defaultConfig.SubNet != "" {
			opts = append(opts, server.WithVpnSubNet(defaultConfig.SubNet))
		}
		if defaultConfig.Port != "" {
			opts = append(opts, server.WithGrpcPort(defaultConfig.Port))
		}
		if defaultConfig.JwtSalt != "" {
			opts = append(opts, server.WithVpnJwtSalt(defaultConfig.JwtSalt))
		}
		if defaultConfig.TlsCertification != "" {
			opts = append(opts, server.WithGrpcTlsCertification(defaultConfig.TlsCertification))
		}
		if defaultConfig.TlsPem != "" {
			opts = append(opts, server.WithGrpcTlsPem(defaultConfig.TlsPem))
		}
		if defaultConfig.JwtExpiration > 0 {
			opts = append(opts, server.WithVpnJwtExpiration(defaultConfig.JwtExpiration))
		}

		// apply auth interceptors
		var authMethods []auth.ServerAuthMethod
		if auth1, ok := defaultConfig.GoogleConfig.ServerAuth(); ok {
			authMethods = append(authMethods, auth1)
		}
		if auth2, ok := defaultConfig.AwsConfig.ServerAuth(); ok {
			authMethods = append(authMethods, auth2)
		}
		opts = append(opts, server.WithAuthMethods(authMethods))

		// create server
		server, err := server.NewVpnServer(opts...)
		if err != nil {
			log.Panicln(color.RedString("[ERR] %s", err.Error()))
		}
		if err := server.Run(); err != nil {
			log.Panicln(color.RedString("[ERR] %s", err.Error()))
		}
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}
