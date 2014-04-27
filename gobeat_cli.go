package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
)

// settings manages global state from the application. It is either retrieved or
// created before command invocation, and saved to disk after execution.
var settings *gobeatSettings

func main() {
	s, err := retrieveSettings()
	if err != nil {
		printError(err)
	}
	settings = s

	app := setupCliApp()

	if err := app.Run(os.Args); err != nil {
		printError(err)
	}
}

// printError handles the exiting of the program and displaying all errors. No-
// op if error is nil.
func printError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: "+err.Error())
		os.Exit(1)
	}
}

// setupCliApp initializes a new *cli.App and populates its fields and flags.
func setupCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "gobeat"
	app.Usage = `gobeat Tweets scores of game matches from an account configured
	    server-side.`
	app.Author = "Alex Toombs"

	populateCommands(app)
	return app
}

// populateCommands sets up all commands on the new command line application.
func populateCommands(app *cli.App) {
	app.Commands = []cli.Command{
		cli.Command{
			Name:        "target",
			ShortName:   "t",
			Description: "`target` sets the URL of the server that gobeat talks to.",
			Usage:       "target [url]",
			Action: func(c *cli.Context) {
				if len(c.Args()) == 0 {
					fmt.Printf("Current target: %s\n", settings.TargetURL)
				} else {
					settings.TargetURL = c.Args().First()

					// Attempt to parse.
					u, err := settings.URL()
					if err != nil {
						printError(err)
					}
					fmt.Printf("Set target to %s\n", u.String())

					if err := settings.save(); err != nil {
						printError(err)
					}
				}
			},
		},
		cli.Command{
			Name:        "user",
			ShortName:   "u",
			Description: "`user` sets the current user.",
			Usage:       "user [username]",
			Action: func(c *cli.Context) {
				if len(c.Args()) == 0 {
					fmt.Printf("Current user: %s\n", settings.User)
				} else {
					settings.User = c.Args().First()
					fmt.Printf("Set user to %s\n", settings.User)

					if err := settings.save(); err != nil {
						printError(err)
					}
				}
			},
		},
		cli.Command{
			Name:        "result",
			ShortName:   "r",
			Description: "`result` sends a result to be tweeted.",
			Usage:       "result [opponent] [score]",
			Action: func(c *cli.Context) {
				if len(c.Args()) == 0 {
					printError(fmt.Errorf("missing opponent name and score."))
				} else if len(c.Args()) == 1 {
					printError(fmt.Errorf("missing opponent name and score."))
				}
				opponent := c.Args().First()
				score := c.Args().Get(1)

				u, err := settings.URL()
				if err != nil {
					printError(err)
				}

				if err := postResult(u, opponent, score); err != nil {
					printError(err)
				}

				fmt.Println("Successfully posted result. Congratulations!")
			},
		},
	}
}

// postResult posts a match result to the configured target.
func postResult(u *url.URL, opponent, score string) error {
	if u == nil || u.String() == "" {
		return fmt.Errorf("cannot post with empty URL")
	}

	client := http.Client{}
	req, err := http.NewRequest("POST", u.String(), formatResult(opponent, score))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	errStr := "on request: got code %d"
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	default:
		return fmt.Errorf(errStr, resp.StatusCode)
	}
	return nil
}

// formatResult formats the body posted to the server.
func formatResult(opponent, score string) *strings.Reader {
	return strings.NewReader(fmt.Sprintf("%s beat %s at %s with score %s",
		settings.User, opponent, settings.Game, score))
}

// retrieveSettings attempts to locate the settings of the app, contained in
// '~/.gobeat' by default.
func retrieveSettings() (*gobeatSettings, error) {
	f, err := os.Open(gobeatPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		} else {
			// File does not exist, so we create a new settings object.
			s := new(gobeatSettings)
			if err := s.assignDefaults(); err != nil {
				return nil, err
			}
			return s, nil
		}
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &settings); err != nil {
		return nil, err
	}

	if err := settings.assignDefaults(); err != nil {
		return nil, err
	}
	return settings, nil
}

const settingsFile = ".gobeat"

// gobeatPath is the full path to where the gobeat settings file resides.
var gobeatPath = filepath.Join(os.Getenv("HOME"), settingsFile)

// gobeatSettings is marshalled to disk to set configuration about target.
type gobeatSettings struct {
	// TargetURL is the URL that the gobeat server is serving at. Set with the
	// 'gobeat target' command.
	TargetURL string `json:"target_url"`

	// User is the command line user's ID. Populated from os/user.Current().
	User string `json:"user"`

	// Game is the type of game (e.g., ping pong) played. Defaults to "ping
	// pong".
	// TODO(alex): allow users to modify this.
	Game string `json:"game"`
}

// assignDefaults populates the settings object with default values.
func (g *gobeatSettings) assignDefaults() error {
	// Provide a default value for username by looking up current user.
	if g.User == "" {
		user, err := user.Current()
		if err != nil {
			return err
		}
		g.User = user.Name
	}

	// Defaults to ping pong for now.
	if g.Game == "" {
		g.Game = "ping pong"
	}
	return nil
}

// save saves to disk a settings file in '~/.gobeat'.
func (g *gobeatSettings) save() error {
	b, err := json.Marshal(g)
	if err != nil {
		return err
	}

	tmpPath := filepath.Join(os.TempDir(), "temp_gobeat")
	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	// Move into correct path.
	return os.Rename(tmpPath, gobeatPath)
}

// URL returns the fully-resolved URL from the gobeat settings.
func (g *gobeatSettings) URL() (*url.URL, error) {
	return url.Parse(g.TargetURL)
}
