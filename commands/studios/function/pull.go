package function

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hofstadter-io/hof/pkg/studios/function"
)

var PullLong = `Replaces the local copy with the latest copy in Studios`

var PullCmd = &cobra.Command{

	Use: "pull",

	Short: "Get the latest version from Studios",

	Long: PullLong,

	Run: func(cmd *cobra.Command, args []string) {

		// fmt.Println("hof function pull:")

		err := function.Pull()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	},
}
