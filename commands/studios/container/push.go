package container

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hofstadter-io/hof/pkg/studios/container"
)

var PushLong = `Uploads the local copy and makes it the latest copy in Studios`

var PushCmd = &cobra.Command{

	Use: "push [name]",

	Short: "Send the latest version on Studios",

	Long: PushLong,

	Run: func(cmd *cobra.Command, args []string) {

		var name string
		if 0 < len(args) {
			name = args[0]
		}

		/*
			fmt.Println("hof containers push:",
				name,
			)
		*/

		err := container.Push(name)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	},
}
