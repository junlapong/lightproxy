package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/octavore/naga/service"
)

func (a *App) addCommands(c *service.Config) {
	c.AddCommand(&service.Command{
		Keyword:    "init",
		ShortUsage: "Initialize the config file",
		Usage:      "Initialize a default config file if it doesn't already exist, and print its location",
		Run:        a.cmdInitConfig,
	})

	c.AddCommand(&service.Command{
		Keyword:    "config",
		ShortUsage: "Prints the config file",
		Usage:      "Prints the config file",
		Run:        a.cmdPrintConfig,
	})

	c.AddCommand(&service.Command{
		Keyword:    "set-dest <domain> <port>",
		ShortUsage: "Map <domain> to <port>",
		Usage:      "Map <domain> to <port>",
		Run:        a.cmdSetHost,
	})

	c.AddCommand(&service.Command{
		Keyword:    "set-dir <domain> <folder>",
		ShortUsage: "Map <domain> to files in <folder>",
		Usage:      "Map <domain> to files in <folder>",
		Run:        a.cmdSetHostFolder,
	})

	c.AddCommand(&service.Command{
		Keyword:    "rm-dest <domain>",
		ShortUsage: "Remove mapping for <domain>",
		Usage:      "Remove mapping for <domain>",
		Run:        a.cmdRmHost,
	})

	c.AddCommand(&service.Command{
		Keyword:    "version",
		ShortUsage: "Print version",
		Usage:      "Print version",
		Run: func(*service.CommandContext) {
			fmt.Println("lightproxy", version)
		},
	})
}

func (a *App) cmdInitConfig(ctx *service.CommandContext) {
	err := a.configManager.ensure()
	if err != nil {
		ctx.Fatal(err.Error())
	}
}

func (a *App) cmdPrintConfig(ctx *service.CommandContext) {
	configPath, _, exists := a.configManager.configPath()
	if !exists {
		fmt.Println("config file does not exist")
		fmt.Println("run `lightproxy init` to initialize at " + configPath)
		return
	}
	config, err := a.configManager.ensureAndLoad()
	if err != nil {
		ctx.Fatal(err.Error())
	}
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		ctx.Fatal(err.Error())
	}
	fmt.Printf("found config file: %s\n\n", configPath)
	fmt.Println(string(b))
}

func (a *App) cmdSetHost(ctx *service.CommandContext) {
	ctx.RequireExactlyNArgs(2)
	config, err := a.configManager.ensureAndLoad()
	if err != nil {
		ctx.Fatal(err.Error())
	}

	host := ctx.Args[0]
	port, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		ctx.Fatal("expected port to be an int")
	}

	dest := fmt.Sprintf("localhost:%d", port)
	found := false
	for _, e := range config.Entries {
		if e.Source == host {
			fmt.Printf("replacing existing entry for %s: %s\n", host, e.DestHost)
			e.DestHost = dest
			e.DestFolder = ""
			found = true
		}
	}
	if !found {
		config.Entries = append(config.Entries, &Entry{
			Source:   host,
			DestHost: dest,
		})
	}
	err = a.configManager.writeConfig(config)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	fmt.Printf("registered: %s => %s\n", host, dest)
}

func (a *App) cmdSetHostFolder(ctx *service.CommandContext) {
	ctx.RequireExactlyNArgs(2)
	config, err := a.configManager.ensureAndLoad()
	if err != nil {
		ctx.Fatal(err.Error())
	}

	host, dir := ctx.Args[0], ctx.Args[1]
	absDir, err := filepath.Abs(dir)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	fmt.Println(absDir)
	found := false
	for _, e := range config.Entries {
		if e.Source == host {
			fmt.Printf("replacing existing entry for %s: %s\n", host, e.DestHost)
			e.DestHost = ""
			e.DestFolder = absDir
			found = true
		}
	}
	if !found {
		config.Entries = append(config.Entries, &Entry{
			Source:     host,
			DestFolder: absDir,
		})
	}
	err = a.configManager.writeConfig(config)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	fmt.Printf("registered: %s => %s\n", host, absDir)
}

func (a *App) cmdRmHost(ctx *service.CommandContext) {
	ctx.RequireExactlyNArgs(1)
	config, err := a.configManager.ensureAndLoad()
	if err != nil {
		ctx.Fatal(err.Error())
	}

	host := ctx.Args[0]
	entries := []*Entry{}
	for _, e := range config.Entries {
		if e.Source != host {
			entries = append(entries, e)
		}
	}
	config.Entries = entries
	err = a.configManager.writeConfig(config)
	if err != nil {
		ctx.Fatal(err.Error())
	}
	fmt.Printf("removed: %s\n", host)
}
