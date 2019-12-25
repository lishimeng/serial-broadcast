package cmd

import (
	"github.com/spf13/cobra"
)

type Params struct {
	Port uint16
	SerialName string
	Baud int
}

var rootCmd = &cobra.Command{
	Use: "serial broadcast",
	Short: "transport serial to tcp",
	Long: "transport serial to tcp",
	Run: handler,
}

var p = Params{}

var cb func(p Params)

func init() {
	rootCmd.Flags().StringVarP(&p.SerialName, "serial", "s", "", "serial name")
	rootCmd.Flags().Uint16VarP(&p.Port, "port", "p", 0, "tcp port")
	rootCmd.Flags().IntVarP(&p.Baud, "baud", "b", 115200, "serial baud")
}

func handler(cmd *cobra.Command, _ []string) {
	if len(p.SerialName) == 0 || p.Port <= 0 {
		_ = cmd.Help()
		return
	}
	if cb != nil {
		cb(p)
	}
}

func Exec(h func(p Params)) (err error) {

	cb = h
	err = rootCmd.Execute()

	return
}