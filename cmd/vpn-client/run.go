package main

import (
	"log"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/gjbae1212/grpc-vpn/client"
	"github.com/mitchellh/go-ps"
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

		// if already process is running, exit.
		processes, err := ps.Processes()
		if err != nil {
			log.Printf(color.RedString(err.Error()))
			os.Exit(1)
		}

		processMap := map[string]int{}
		var processName string
		for _, process := range processes {
			if process.Pid() == os.Getpid() {
				processName = process.Executable()
			}
			processMap[process.Executable()] += 1
		}
		if processMap[processName] >= 2 {
			log.Printf(color.RedString("[ERR] ALREADY %s RUNNING", processName))
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
		if defaultConfig.SelfSignedCertification != "" {
			opts = append(opts, client.WithSelfSignedCertification(defaultConfig.SelfSignedCertification))
		}
		opts = append(opts, client.WithGRPCInsecure(defaultConfig.Insecure))

		// aws authentication
		method1, ok1 := defaultConfig.Auth.ClientAuthForAwsIAM()
		if ok1 {
			opts = append(opts, client.WithAuthMethod(method1))
		}

		// google authentication
		method2, ok2 := defaultConfig.Auth.ClientAuthForGoogleOpenID()
		if ok2 {
			opts = append(opts, client.WithAuthMethod(method2))
		}

		// if both method1 and method2 is empty.
		if !ok1 && !ok2 {
			method3, _ := defaultConfig.Auth.ClientAuthForTest()
			opts = append(opts, client.WithAuthMethod(method3))
		}

		client, err := client.NewVpnClient(opts...)
		if err != nil {
			log.Println(color.RedString("[ERR] %s", err.Error()))
			os.Exit(1)
		}

		if err := client.Run(); err != nil {
			log.Println(color.RedString("[ERR] %s", err.Error()))
			os.Exit(1)
		}
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
}
