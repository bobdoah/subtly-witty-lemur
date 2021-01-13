package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/philhofer/tcx"
	"github.com/urfave/cli/v2"
)

func readTcx(filepath string) error {
	db, err := tcx.ReadFile(filepath)
	if err != nil {
		return err
	}
	nacts := len(db.Acts.Act)
	if nacts > 0 {
		act := db.Acts.Act[0]
		fmt.Printf("id: %s sport: %s\n", act.Id.Format(time.RFC3339), act.Sport)
	}

	return nil
}

func main() {
	app := &cli.App{
		Name:  "greet",
		Usage: "fight the loneliness!",
		Action: func(c *cli.Context) error {
			if c.NArg() > 0 {
				var i int
				for i = 0; i < c.Args().Len(); i++ {
					readTcx(c.Args().Get(i))
				}
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
