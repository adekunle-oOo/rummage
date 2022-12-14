package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

type Configuration struct {
	ServerIP   string
	ServerPort int
	EthGateway string
	HLI        string
	PassW      string
	ServerAddr string
}

var (
	configuration Configuration
	app1          = cli.NewApp()
)

//load parameters from config file
//requires config file to be present
func setConfig() error {
	file, err := os.Open("./config1.json")
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		return err
	}
	return nil
}

func info() {
	app1.Name = "Rummage"
	app1.Usage = "Decentralised Search for IPFS"
	app1.Authors = []*cli.Author{
		{Name: "Navin V. Keizer", Email: "navin.keizer.15@ucl.ac.uk"},
		{Name: "Puneet K. Bindlish", Email: "p.k.bindlish@vu.nl"},
	}
	app1.Version = "0.0.1"

}

func commands() {
	ty := ""
	app1.Commands = []*cli.Command{
		{
			Name:        "search",
			Usage:       "Performs a decentralised search on IPFS",
			Description: "Retrieves the index to find which pages contain the keywords",

			Action: func(c *cli.Context) error {
				if c.Args().Len() < 1 {
					return &Rummage.IncorrrectInput{}
				}
				//ensure swarm is connected to gateway peer
				err := Rummage.Shell.SwarmConnect(context.Background(), configuration.ServerAddr)
				if err != nil {
					log.Println(err)
				}
				fmt.Println("searching...")
				err = Rummage.DoSearchClient(c.Args().Slice())
				if err != nil {
					return err
				}
				return nil
			},
		},

		{
			Name:        "crawl",
			Usage:       "Crawls a page to add to the decentralised index",
			Description: "Crawls the page, extracts keywords using OCR, and adds to the index stored on IPFS",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "type",
					Aliases:     []string{"t"},
					Usage:       "Input domain type (default CID)",
					Destination: &ty,
				},
			},

			Before: func(c *cli.Context) error {
				fmt.Println("Start crawl...")
				//ensure swarm is connected to gateway peer
				err := Rummage.Shell.SwarmConnect(context.Background(), configuration.ServerAddr)
				if err != nil {
					log.Println(err)
				}
				return nil
			},

			Action: func(c *cli.Context) error {
				if c.Args().Len() != 1 {
					return &Rummage.IncorrrectInput{}
				}
				id := c.Args().Get(0)
				if ty == "" || ty == "CID" {
					err := Rummage.DoCrawlClient(id, "CID")
					if err != nil {
						return err
					}
					return nil
				} else if ty == "ENS" {
					err := Rummage.DoCrawlClient(id, "ENS")
					if err != nil {
						return err
					}
					return nil
				} else if ty == "DNS" {
					err := Rummage.DoCrawlClient(id, "DNS")
					if err != nil {
						return err
					}
					return nil
				} else if ty == "IPNS" {
					err := Rummage.DoCrawlClient(id, "IPNS")
					if err != nil {
						return err
					}
					return nil
				}
				return &Rummage.IncorrrectInput{}
			},
		},
	}
}

//main CLI program
func main() {
	err := setConfig()
	if err != nil {
		log.Fatal(err)
	}
	Rummage.Shell, Rummage.Client = Rummage.ConnectClient(configuration.EthGateway,
		configuration.HLI, configuration.ServerIP, configuration.ServerPort, configuration.PassW)
	info()
	commands()

	err = app1.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
