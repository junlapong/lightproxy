package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/octavore/naga/service"
)

func (a *App) addCommands(c *service.Config) {
	c.AddCommand(&service.Command{
		Keyword:    "init",
		ShortUsage: "initialize the config file",
		Usage:      "Initialize a default config file if it doesn't already exist, and print its location",
		Run:        a.cmdInitConfig,
	})

	c.AddCommand(&service.Command{
		Keyword:    "config",
		ShortUsage: "prints the config file",
		Usage:      "Prints the config files",
		Run:        a.cmdPrintConfig,
	})

	c.AddCommand(&service.Command{
		Keyword:    "set-dest <domain> <port>",
		ShortUsage: "map <domain> to <port>",
		Usage:      "Map <domain> to <port>",
		Run:        a.cmdSetHost,
	})

	c.AddCommand(&service.Command{
		Keyword:    "version",
		ShortUsage: "print version",
		Usage:      "Print version",
		Run: func(*service.CommandContext) {
			fmt.Println("lightproxy", version)
		},
	})
}

func (a *App) cmdInitConfig(ctx *service.CommandContext) {
	fi, err := os.Stat(a.configPath())
	if fi != nil && err == nil {
		fmt.Printf("found init file: %s\n", a.configPath())
		return
	}
	if !os.IsNotExist(err) {
		ctx.Fatal("unknown error: %s", err)
	}

	err = os.MkdirAll(a.configDir(), os.ModePerm)
	if err != nil {
		ctx.Fatal("failed to create dir %s: %s", a.configDir(), err)
	}

	f, err := os.Create(a.configPath())
	defer f.Close()
	if err != nil {
		ctx.Fatal("failed to create dir %s: %s", a.configDir(), err)
	}

	b, err := json.MarshalIndent(&Config{
		Addr:    ":7999",
		TLD:     "test",
		Entries: []Entry{},
	}, "", "  ")
	if err != nil {
		ctx.Fatal("failed to to create config.json file: %s", err)
	}

	_, err = f.Write(b)
	if err != nil {
		ctx.Fatal("failed to to create config.json file: %s", err)
	}
	fmt.Printf("created init file: %s\n", a.configPath())
}

func (a *App) cmdPrintConfig(ctx *service.CommandContext) {
	err := a.loadConfig()
	if err != nil {
		ctx.Fatal(err.Error())
	}
	b, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		ctx.Fatal(err.Error())
	}
	fmt.Printf("found config %s:\n\n", a.configPath())
	fmt.Println(string(b))
}

func (a *App) cmdSetHost(ctx *service.CommandContext) {
	ctx.RequireExactlyNArgs(2)
	err := a.loadConfig()
	if err != nil {
		// todo: more helpful error if config.json does not exist
		ctx.Fatal(err.Error())
	}

	host := ctx.Args[0]
	port, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		ctx.Fatal("expected port to be an int")
	}

	dest := fmt.Sprintf("localhost:%d", port)
	found := false
	for _, e := range a.config.Entries {
		if e.Source == host {
			fmt.Printf("replacing existing entry for %s: %s\n", host, *e.DestHost)
			e.DestHost = &dest
			found = true
		}
	}
	if !found {
		a.config.Entries = append(a.config.Entries, Entry{
			Source:   host,
			DestHost: &dest,
		})
	}
	b, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		ctx.Fatal(err.Error())
	}
	err = ioutil.WriteFile(a.configPath(), b, os.ModePerm)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	fmt.Printf("registered: %s => %s\n", host, dest)
}
