package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gilliek/go-opml/opml"
	"github.com/go-ini/ini"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const Version = "0.4.0"

func addFeed(name, url string) error {
	sanitizedName := strings.TrimSpace(name)
	sanitizedURL := strings.TrimSpace(url)

	if sanitizedName == "" {
		return errors.New("feed name cannot be empty")
	}
	if sanitizedURL == "" {
		return errors.New("feed URL cannot be empty")
	}

	cfgPath, err := configFilePath()
	if err != nil {
		return err
	}

	cfg, err := ini.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config for update: %w", err)
	}

	feedsSection := cfg.Section("feeds")
	if feedsSection.HasKey(sanitizedName) {
		return fmt.Errorf("feed named '%s' already exists", sanitizedName)
	}

	feedsSection.Key(sanitizedName).SetValue(sanitizedURL)

	if err := cfg.SaveTo(cfgPath); err != nil {
		return fmt.Errorf("persist feed configuration: %w", err)
	}

	Config.Feeds = append(Config.Feeds, Feed{Name: sanitizedName, URL: sanitizedURL})

	return nil
}

func main() {
	syscall.Umask(0)
	if err := LoadConfig(); err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "opens the config file with $EDITOR",
				Action: func(cCtx *cli.Context) error {
					editor, ok := os.LookupEnv("EDITOR")
					if len(editor) == 0 || !ok {
						return errors.New("$EDITOR environment variable is not set")
					}
					cfgPath, err := configFilePath()
					if err != nil {
						return err
					}
					cmd := exec.Command(editor, cfgPath)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					return cmd.Run()
				},
			},
			{
				Name:    "refetch",
				Aliases: []string{"r"},
				Usage:   "delete and refetch given feed(s) or all feeds if no argument is given",
				Action: func(cCtx *cli.Context) error {
					if err := InitialiseNewArticleDirectory(); err != nil {
						return err
					}
					if cCtx.Args().Len() == 0 {
						UpdateAllFeeds(true)
						return nil
					}
					for i := 0; i < cCtx.Args().Len(); i++ {
						name := cCtx.Args().Get(i)
						if _, err := UpdateFeed(name, true); err != nil {
							log.Error(err)
						}
					}
					return nil
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "update given feed(s) or all feeds if no argument is given",
				Action: func(cCtx *cli.Context) error {
					if err := InitialiseNewArticleDirectory(); err != nil {
						return err
					}
					if cCtx.Args().Len() == 0 {
						UpdateAllFeeds(false)
						return nil
					}
					for i := 0; i < cCtx.Args().Len(); i++ {
						name := cCtx.Args().Get(i)
						if _, err := UpdateFeed(name, false); err != nil {
							log.Error(err)
						}
					}
					return nil
				},
			},
			{
				Name:    "open",
				Aliases: []string{"o"},
				Usage:   "open given feed's directory or root feeds directory if no argument is given",
				Action: func(cCtx *cli.Context) error {
					path := Config.FeedDirectory
					if cCtx.Args().Len() > 0 {
						path = filepath.Join(Config.FeedDirectory, cCtx.Args().Get(0))
					}
					cmd := exec.Command(Config.Viewer, path)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					return cmd.Run()
				},
			},
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a given feed to config",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Len() != 2 {
						return errors.New("exactly two arguments are required, first being feed name, second being URL")
					}
					return addFeed(cCtx.Args().Get(0), cCtx.Args().Get(1))
				},
			},
			{
				Name:    "import",
				Aliases: []string{"i"},
				Usage:   "import an OPML file",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Len() != 1 {
						return errors.New("argument specifying OPML file path or URL is required")
					}
					doc, err := opml.NewOPMLFromFile(cCtx.Args().Get(0))
					if err != nil {
						doc, err = opml.NewOPMLFromURL(cCtx.Args().Get(0))
						if err != nil {
							return err
						}
					}
					for _, outline := range doc.Body.Outlines {
						if len(outline.XMLURL) > 0 {
							var title string
							if len(outline.Title) > 0 {
								title = outline.Title
							} else if len(outline.Text) > 0 {
								title = outline.Text
							} else {
								continue
							}
							if err := addFeed(strings.ReplaceAll(title, " ", "-"), outline.XMLURL); err != nil {
								log.Errorf("Failed to add feed titled '%s': %v", title, err)
								continue
							}
						}
						for _, innerOutline := range outline.Outlines {
							if len(innerOutline.XMLURL) > 0 {
								var title string
								if len(innerOutline.Title) > 0 {
									title = innerOutline.Title
								} else if len(innerOutline.Text) > 0 {
									title = innerOutline.Text
								} else {
									continue
								}
								if err := addFeed(strings.ReplaceAll(title, " ", "-"), innerOutline.XMLURL); err != nil {
									log.Errorf("Failed to add feed titled '%s': %v", title, err)
									continue
								}
							}
						}
					}
					return nil
				},
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "display the version",
				Action: func(cCtx *cli.Context) error {
					log.Info(Version)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
	}
}
