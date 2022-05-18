package main

import (
	"log"
	"os"
	"runtime"

	"github.com/limitcool/dl/downloader"
	"github.com/urfave/cli/v2"
)

func main() {
	Process := runtime.NumCPU()
	log.Println("Process:", Process)
	app := &cli.App{
		Name:    "dl",
		Usage:   "File concurrency downloader",
		Version: "0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Usage:    "url",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "output",
				Aliases: []string{"o"},
			},
			&cli.IntFlag{
				Name:  "thread",
				Usage: "thread",
				Value: Process,
			},
		},
		Action: func(c *cli.Context) error {
			url := c.String("url")
			output := c.String("output")
			thread := c.Int("thread")

			return downloader.NewDownloader(thread).Download(url, output)
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
