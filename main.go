package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	. "github.com/tendermint/go-common"
	pcm "github.com/tendermint/go-process"
)

func main() {
	app := cli.NewApp()
	app.Name = "devtool"
	app.Usage = "devtool [command] [args...]"
	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "List dependencies and show info",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config",
					Value: "config.json",
				},
			},
			Action: func(c *cli.Context) {
				cmdList(app, c)
			},
		},
	}
	app.Run(os.Args)

}

//--------------------------------------------------------------------------------

type Config struct {
	Repos []Repo `json:"repos"`
}

type Repo struct {
	Path string `json:"path"`
}

//--------------------------------------------------------------------------------

func cmdList(app *cli.App, c *cli.Context) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		Exit("$GOPATH must be set")
	}

	configBytes, err := ReadFile(c.String("config"))
	if err != nil {
		Exit(Fmt("Error reading config file: %v", err))
	}

	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		Exit(Fmt("Error parsing config file: %v", err))
	}

	if len(config.Repos) == 0 {
		Exit(Fmt("Config file has no repos"))
	}

	for _, repo := range config.Repos {
		printRepoInfo(goPath, repo)
	}
}

func printRepoInfo(goPath string, repo Repo) {
	path := Fmt("%v/src/%v", goPath, repo.Path)
	resBranch, success, err := pcm.Run(path, "git", []string{"symbolic-ref", "--short", "-q", "HEAD"})
	if !success {
		fmt.Printf("%-40s    %v\n%v\n", repo.Path, Red(err), Red(strings.Trim(resBranch, "\n")))
		return
	}
	fmt.Printf("%-40s    %10s\n", repo.Path, Green(strings.Trim(resBranch, "\n")))
	resStatus, _, err := pcm.Run(path, "git", []string{"status", "--short"})
	if err != nil {
		Exit(Fmt("Error fetching git status for %v: %v", repo.Path, err))
	}
	status := strings.Trim(resStatus, "\n")
	if status != "" {
		fmt.Println(Yellow(status))
	}
}
