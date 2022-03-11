package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/releasr/pkg"
	"github.com/packagrio/releasr/pkg/config"
	"github.com/packagrio/releasr/pkg/version"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

var goos string
var goarch string

func main() {
	app := &cli.App{
		Name:     "releasr",
		Usage:    "Language agnostic tool to package a git repo. Commit any local changes and create a git tag. It should also generate artifacts. Nothing should be pushed to remote repository",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			packagrUrl := "github.com/packagrio/releasr"

			versionInfo := fmt.Sprintf("%s.%s-%s", goos, goarch, version.VERSION)

			subtitle := packagrUrl + utils.LeftPad2Len(versionInfo, " ", 53-len(packagrUrl))

			fmt.Fprintf(c.App.Writer, fmt.Sprintf(utils.StripIndent(
				`
			 ____   __    ___  __ _   __    ___  ____ 
			(  _ \ / _\  / __)(  / ) / _\  / __)(  _ \
			 ) __//    \( (__  )  ( /    \( (_ \ )   /
			(__)  \_/\_/ \___)(__\_)\_/\_/ \___/(__\_)
			%s

			`), subtitle))
			return nil
		},

		Commands: []cli.Command{
			{
				Name:  "start",
				Usage: "Start releasr pipeline",
				Action: func(c *cli.Context) error {

					configuration, _ := config.Create()
					
					if c.IsSet("scm") {
						configuration.Set(config.PACKAGR_SCM, c.String("scm"))
					}
					if c.IsSet("package_type") {
						configuration.Set(config.PACKAGR_PACKAGE_TYPE, c.String("package_type"))
					}

					fmt.Println("package type:", configuration.GetString(config.PACKAGR_PACKAGE_TYPE))
					fmt.Println("scm:", configuration.GetString(config.PACKAGR_SCM))

					pipeline := pkg.Pipeline{}
					err := pipeline.Start(configuration)
					if err != nil {
						fmt.Printf("FATAL: %+v\n", err)
						os.Exit(1)
					}

					return nil
				},

				Flags: []cli.Flag{
					//TODO: currently not applicable
					//&cli.StringFlag{
					//	Name:  "runner",
					//	Value: "default", // can be :none, :circleci or :shippable (check the readme for why other hosted providers arn't supported.)
					//	Usage: "The cloud CI runner that is running this PR. (Used to determine the Environmental Variables to parse)",
					//},

					&cli.StringFlag{
						Name:  "scm",
						Value: "default",
						Usage: "The scm for the code, for setting additional SCM specific metadata",
					},

					&cli.StringFlag{
						Name:  "package_type",
						Value: "generic",
						Usage: "The type of package being built.",
					},

					&cli.BoolFlag{
						Name:  "dry_run",
						Usage: "When dry run is enabled, no data is written to file system",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
