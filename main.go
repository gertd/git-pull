package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// command line args
const (
	GitOrg   = "org"
	GitToken = "token"
)

// environment variables
const (
	GitOrgEnv   = "GIT_ORG"
	GitTokenEnv = "GIT_TOKEN"
)

// command line usage
const (
	GitOrgUsage   = "GitHub organization"
	GitTokenUsage = "GitHub private access token"
)

// command literals
const (
	AppName  = "git-pull"
	AppUsage = "pull all GitHub repos in current directory"
	GitCmd   = "git"
	GitPull  = "pull"
)

func main() {

	app := cli.NewApp()
	app.Name = AppName
	app.Usage = AppUsage
	app.Action = gitPull
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   GitOrg,
			Value:  "",
			Usage:  GitOrgUsage,
			EnvVar: GitOrgEnv,
		},
		cli.StringFlag{
			Name:   GitToken,
			Value:  "",
			Usage:  GitTokenUsage,
			EnvVar: GitTokenEnv,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}

	os.Exit(1)
}

func gitPull(c *cli.Context) error {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.GlobalString(GitToken)},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	page := 1

	cmdName := GitCmd

	for {

		opt := &github.RepositoryListByOrgOptions{
			Type:        "all",
			ListOptions: github.ListOptions{PerPage: 10, Page: page},
		}

		repos, resp, err := client.Repositories.ListByOrg(c.GlobalString(GitOrg), opt)
		if err != nil {
			log.Fatalln(err)
		}

		for _, v := range repos {

			// check if directory/.git exists
			checkPath := "../" + *v.FullName + "/.git"

			if _, err := os.Stat(checkPath); os.IsNotExist(err) {

				fmt.Printf("%s ", *v.FullName)
				fmt.Printf("does not exists\n")

			} else {

				// fmt.Printf("exists\n")

				os.Chdir("../" + *v.FullName)

				pwd, err := os.Getwd()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Printf("%s\n", pwd)

				cmdArgs := []string{GitPull, *v.CloneURL}

				cmd := exec.Command(cmdName, cmdArgs...)
				cmdReader, err := cmd.StdoutPipe()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error: creating StdoutPipe for Cmd", err)
					return err
				}

				scanner := bufio.NewScanner(cmdReader)
				go func() {
					for scanner.Scan() {
						fmt.Printf("%s\n", scanner.Text())
					}
				}()

				err = cmd.Start()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error: starting Cmd", err)
					return err
				}

				err = cmd.Wait()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error: waiting for Cmd", err)
					// return err
				}

				os.Chdir("..")

			}
		}

		if resp.NextPage == 0 {
			break
		}

		page = resp.NextPage
	}
	return nil
}
