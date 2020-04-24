package cmd

import (
	"log"
	"os"
	"runtime"

	"github.com/gjbae1212/grpc-vpn/pkg/vpn-client/client"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:    "run",
		Short:  "Start vpn-client",
		Long:   "Start vpn-client",
		PreRun: startPreRun(),
		Run:    startRun(),
	}
)

func startPreRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		if runtime.GOOS == "windows" {
			log.Printf(color.RedString("Window OS doesn't support."))
			os.Exit(1)
		}
		if os.Getuid() != 0 {
			log.Printf("%s %s", color.RedString("[RETRY][COMMAND]"),
				color.CyanString("`sudo vpn-client run`"))
			os.Exit(1)
		}
	}
}

func startRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		// apply default params
		var opts []client.Option
		if defaultConfig.Addr != "" {
			opts = append(opts, client.WithServerAddr(defaultConfig.Addr))
		}
		if defaultConfig.Port != "" {
			opts = append(opts, client.WithServerPort(defaultConfig.Port))
		}
		if defaultConfig.TlsCertification != "" {
			opts = append(opts, client.WithTlsCertification(defaultConfig.TlsCertification))
		}

		// TODO: AUTH(GOOGLE, AWS)

		client, err := client.NewVpnClient(opts...)
		if err != nil {
			log.Panicln(color.RedString("[ERR] %s", err.Error()))
		}

		if err := client.Run(); err != nil {
			log.Panicln(color.RedString("[ERR] %s", err.Error()))
		}
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}
