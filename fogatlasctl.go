package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/ghodss/yaml"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
	apiclient "github.com/fogatlas/client-go/client"
	"github.com/fogatlas/client-go/client/operations"
	"github.com/fogatlas/client-go/models"
)

type confFile struct {
	Applications     []models.Application      `json:"applications,omitempty"`
	Microservices    []models.Microservice     `json:"microservices,omitempty"`
	Relationships    []models.Relationship     `json:"relationships,omitempty"`
	Nodes            []models.Node             `json:"nodes,omitempty"`
	Regions          []models.Region           `json:"regions,omitempty"`
	ExternalEdpoints []models.ExternalEndpoint `json:"externalendpoints,tempty"`
	DynamicNodes     []models.DynamicNode      `json:"dynamicnode,omitempty"`
	Deployments      []models.Deployment       `json:"deployments,omitempty"`
}

func main() {
	cli.AppHelpTemplate = `NAME:
     {{.Name}} - {{.Usage}}{{ "\n"}}
  USAGE:
     {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
     {{if len .Authors}}
  AUTHOR:
     {{range .Authors}}{{ . }}{{end}}
     {{end}}{{if .Commands}}
  COMMANDS:
  {{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n"}}  {{end}}{{end}}{{end}}{{if .VisibleFlags}}
  GLOBAL OPTIONS:
     {{range .VisibleFlags}}{{.}}
     {{end}}{{end}}{{if .Copyright }}
  COPYRIGHT:
     {{.Copyright}}
     {{end}}{{if .Version}}
  VERSION:
     {{.Version}}
     {{end}}
  `

	cli.CommandHelpTemplate = `NAME:
     {{.HelpName}} - {{.Usage}}{{ "\n"}}
  USAGE:
     {{.HelpName}} {{if .VisibleFlags}}[command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{ "\n"}}{{if .VisibleFlags}}
  OPTIONS:
     {{range .VisibleFlags}}--{{.Name}}=<value>{{"\t"}}{{.Usage}}
     {{end}}{{end}}
  `

	app := cli.NewApp()
	app.Name = "fogatlasctl"
	app.Version = "1.3.0"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "FBK CREATE-NET",
			Email: "",
		},
	}
	app.Usage = "Command line interface for FogAtlas"
	app.ArgsUsage = "resource"
	app.Commands = []cli.Command{
		cli.Command{
			Name:      "get",
			Usage:     "retrieve information on a resource",
			ArgsUsage: "{applications|deployments|microservices|nodes|regions|relationships|externalendpoints|dynamicnodes}",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Value: "127.0.0.1:8080",
					Usage: "API endpoint",
				},
				cli.StringFlag{
					Name:  "id",
					Value: "",
					Usage: "identifier of the resource to be retrieved",
				},
				cli.StringFlag{
					Name:  "region_id",
					Value: "",
					Usage: "identifier of the region the resource belongs to (valid only for nodes, externalendpoints, relationships and dynamic nodes)",
				},
				cli.StringFlag{
					Name:  "node_id",
					Value: "",
					Usage: "identifier of the node the resource belongs to (valid only for microservices)",
				},
				cli.StringFlag{
					Name:  "status",
					Value: "",
					Usage: "status of the deployment (valid only for deployments)",
				},
			},
			SkipFlagParsing: false,
			HideHelp:        false,
			Hidden:          false,
			HelpName:        "fogatlasctl get",
			Action: func(c *cli.Context) error {
				return handleGet(c)
			},
			OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
				return err
			},
		},
		cli.Command{
			Name:      "put",
			Usage:     "create/update a resource",
			ArgsUsage: "{applications|deployments|microservices|nodes|regions|relationships|externalendpoints|dynamicnodes}",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Value: "127.0.0.1:8080",
					Usage: "API endpoint",
				},
				cli.StringFlag{
					Name:  "id",
					Value: "",
					Usage: "identifier of the resource to be created/updated",
				},
				cli.StringFlag{
					Name:  "file",
					Value: "",
					Usage: "filename containing the resource to be created/updated in json format. Check FogAtlas API documentation.",
				},
			},
			SkipFlagParsing: false,
			HideHelp:        false,
			Hidden:          false,
			HelpName:        "fogatlasctl put",
			Action: func(c *cli.Context) error {
				return handlePut(c)
			},
			OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
				return err
			},
		},
		cli.Command{
			Name:      "patch",
			Usage:     "update a resource",
			ArgsUsage: "{deployments}",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Value: "127.0.0.1:8080",
					Usage: "API endpoint",
				},
				cli.StringFlag{
					Name:  "id",
					Value: "",
					Usage: "identifier of the resource to be created/updated",
				},
				cli.StringFlag{
					Name:  "status",
					Value: "",
					Usage: "value of the status. Check FogAtlas API documentation.",
				},
			},
			SkipFlagParsing: false,
			HideHelp:        false,
			Hidden:          false,
			HelpName:        "fogatlasctl patch",
			Action: func(c *cli.Context) error {
				return handlePatch(c)
			},
			OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
				return err
			},
		},
		cli.Command{
			Name:      "delete",
			Usage:     "delete a resource",
			ArgsUsage: "{applications|microservices|nodes|regions|relationships|externalendpoints|dynamicnodes|deployments}",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Value: "127.0.0.1:8080",
					Usage: "API endpoint",
				},
				cli.StringFlag{
					Name:  "id",
					Value: "",
					Usage: "identifier of the resource to be deleted",
				},
			},
			SkipFlagParsing: false,
			HideHelp:        false,
			Hidden:          false,
			HelpName:        "fogatlasctl delete",
			Action: func(c *cli.Context) error {
				return handleDelete(c)
			},
			OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
				return err
			},
		},
		cli.Command{
			Name:      "putAll",
			Usage:     "create/update a set of resources",
			ArgsUsage: "{}",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Value: "127.0.0.1:8080",
					Usage: "API endpoint",
				},
				cli.StringFlag{
					Name:  "file",
					Value: "",
					Usage: "yaml file that describes the resources to be loaded",
				},
			},
			SkipFlagParsing: false,
			HideHelp:        false,
			Hidden:          false,
			HelpName:        "fogatlasctl putAll",
			Action: handlePutAll,
			OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
				return err
			},
		},
		cli.Command{
			Name:      "deleteAll",
			Usage:     "delete all resources of the given type",
			ArgsUsage: "{applications|microservices|nodes|regions|relationships|externalendpoints|dynamicnodes|deployments}",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Value: "127.0.0.1:8080",
					Usage: "API endpoint",
				},
			},
			SkipFlagParsing: false,
			HideHelp:        false,
			Hidden:          false,
			HelpName:        "fogatlasctl deleteAll",
			Action: handleDeleteAll,
			OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
				return err
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func handleGet(c *cli.Context) error {
	var resp interface{}
	var err error
	schemes := []string{"http"}
	transport := httptransport.New(c.String("endpoint"), "/api/v2.0.0", schemes)
	client := apiclient.New(transport, strfmt.Default)
	switch resource := c.Args().Get(0); resource {
	case "applications":
		if c.String("id") != "" {
			params := operations.NewGetApplicationsIDParams()
			params.ID = c.String("id")
			resp, err = client.Operations.GetApplicationsID(params)
			if err != nil {
				return fmt.Errorf("Error: get applications failed: %s", err)
			}
		} else {
			params := operations.NewGetApplicationsParams()
			resp, err = client.Operations.GetApplications(params)
			if err != nil {
				return fmt.Errorf("Error: get applications failed: %s", err)
			}
		}
	case "deployments":
		if c.String("id") != "" {
			params := operations.NewGetDeploymentsNameParams()
			params.Name = c.String("id")
			resp, err = client.Operations.GetDeploymentsName(params)
			if err != nil {
				return fmt.Errorf("Error: get deployments failed: %s", err)
			}
		} else {
			params := operations.NewGetDeploymentsParams()
			if c.String("status") != "" {
				status := c.String("status")
				params.Status = &status
			}
			resp, err = client.Operations.GetDeployments(params)
			if err != nil {
				return fmt.Errorf("Error: get deployments failed: %s", err)
			}
		}
	case "microservices":
		if c.String("id") != "" {
			params := operations.NewGetMicroservicesIDParams()
			params.ID = c.String("id")
			resp, err = client.Operations.GetMicroservicesID(params)
			if err != nil {
				return fmt.Errorf("Error: get microservices failed: %s", err)
			}
		} else {
			params := operations.NewGetMicroservicesParams()
			if c.String("node_id") != "" {
				node_id := c.String("node_id")
				params.NodeID = &node_id
			}
			resp, err = client.Operations.GetMicroservices(params)
			if err != nil {
				return fmt.Errorf("Error: get microservices failed: %s", err)
			}
		}
	case "nodes":
		if c.String("id") != "" {
			params := operations.NewGetNodesIDParams()
			params.ID = c.String("id")
			resp, err = client.Operations.GetNodesID(params)
			if err != nil {
				return fmt.Errorf("Error: get nodes failed: %s", err)
			}
		} else {
			params := operations.NewGetNodesParams()
			if c.String("region_id") != "" {
				region_id := c.String("region_id")
				params.RegionID = &region_id
			}
			resp, err = client.Operations.GetNodes(params)
			if err != nil {
				return fmt.Errorf("Error: get nodes failed: %s", err)
			}
		}
	case "regions":
		if c.String("id") != "" {
			params := operations.NewGetRegionsIDParams()
			params.ID = c.String("id")
			resp, err = client.Operations.GetRegionsID(params)
			if err != nil {
				return fmt.Errorf("Error: get regions failed: %s", err)
			}
		} else {
			params := operations.NewGetRegionsParams()
			resp, err = client.Operations.GetRegions(params)
			if err != nil {
				return fmt.Errorf("Error: get regions failed: %s", err)
			}
		}
	case "relationships":
		if c.String("id") != "" {
			params := operations.NewGetRelationshipsIDParams()
			params.ID = c.String("id")
			resp, err = client.Operations.GetRelationshipsID(params)
			if err != nil {
				return fmt.Errorf("Error: get relationships failed: %s", err)
			}
		} else {
			params := operations.NewGetRelationshipsParams()
			if c.String("region_id") != "" {
				region_id := c.String("region_id")
				params.RegionID = &region_id
			}
			resp, err = client.Operations.GetRelationships(params)
			if err != nil {
				return fmt.Errorf("Error: get relationships failed: %s", err)
			}
		}
	case "externalendpoints":
		if c.String("id") != "" {
			params := operations.NewGetExternalendpointsIDParams()
			params.ID = c.String("id")
			resp, err = client.Operations.GetExternalendpointsID(params)
			if err != nil {
				return fmt.Errorf("Error: get external endpoints failed: %s", err)
			}
		} else {
			params := operations.NewGetExternalendpointsParams()
			if c.String("region_id") != "" {
				region_id := c.String("region_id")
				params.RegionID = &region_id
			}
			resp, err = client.Operations.GetExternalendpoints(params)
			if err != nil {
				return fmt.Errorf("Error: get external endpoints failed: %s", err)
			}
		}
	case "dynamicnodes":
		if c.String("id") != "" {
			params := operations.NewGetDynamicnodesIDParams()
			params.ID = c.String("id")
			resp, err = client.Operations.GetDynamicnodesID(params)
			if err != nil {
				return fmt.Errorf("Error: get dynamicnodes failed: %s", err)
			}
		} else {
			params := operations.NewGetDynamicnodesParams()
			if c.String("region_id") != "" {
				region_id := c.String("region_id")
				params.RegionID = &region_id
			}
			resp, err = client.Operations.GetDynamicnodes(params)
			if err != nil {
				return fmt.Errorf("Error: get dynamicnodes failed: %s", err)
			}
		}
	default:
		return fmt.Errorf("Error: resource specificed (%s) is unkwnown", resource)
	}
	printData(resp)
	return nil
}

func handlePatch(c *cli.Context) error {
	schemes := []string{"http"}
	transport := httptransport.New(c.String("endpoint"), "/api/v2.0.0", schemes)
	client := apiclient.New(transport, strfmt.Default)
	switch resource := c.Args().Get(0); resource {
	case "deployments":
		if c.String("id") == "" || c.String("status") == "" {
			return fmt.Errorf("Error: options --id and --status are required")
		}
		params := operations.NewPatchDeploymentsNameParams()

		var stat models.PatchStatus
		stat.Status = c.String("status")
		params.PatchStatus = &stat
		params.Name = c.String("id")
		resp, err := client.Operations.PatchDeploymentsName(params)
		if err != nil {
			return fmt.Errorf("Error: patch deployments failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())
	default:
		return fmt.Errorf("Error: resource specificed (%s) is unkwnown", resource)
	}
	return nil
}

func handlePut(c *cli.Context) error {
	schemes := []string{"http"}
	transport := httptransport.New(c.String("endpoint"), "/api/v2.0.0", schemes)
	client := apiclient.New(transport, strfmt.Default)
	switch resource := c.Args().Get(0); resource {
	case "applications":
		if c.String("id") == "" || c.String("file") == "" {
			return fmt.Errorf("Error: options --id and --file are required")
		}
		params := operations.NewPutApplicationsIDParams()
		str, err := getFromFile(c.String("file"))
		if err != nil {
			return fmt.Errorf("Error: unable to read file %s: %s", c.String("file"), err)
		}
		var d models.Application
		b := []byte(str)
		err = json.Unmarshal(b, &d)
		if err != nil {
			return fmt.Errorf("Error: wrong file format: %s", err)
		}
		params.Application = &d
		params.ID = c.String("id")
		resp, err := client.Operations.PutApplicationsID(params)
		if err != nil {
			return fmt.Errorf("Error: put applications failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "deployments":
		if c.String("id") == "" || c.String("file") == "" {
			return fmt.Errorf("Error: options --id and --file are required")
		}
		params := operations.NewPutDeploymentsNameParams()
		str, err := getFromFile(c.String("file"))
		if err != nil {
			return fmt.Errorf("Error: unable to read file %s: %s", c.String("file"), err)
		}
		var d models.Deployment
		b := []byte(str)
		err = json.Unmarshal(b, &d)
		if err != nil {
			return fmt.Errorf("Error: wrong file format: %s", err)
		}
		params.Deployment = &d
		params.Name = c.String("id")
		resp, err := client.Operations.PutDeploymentsName(params)
		if err != nil {
			return fmt.Errorf("Error: put deployments failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())

	case "microservices":
		if c.String("id") == "" || c.String("file") == "" {
			return fmt.Errorf("Error: options --id and --file are required")
		}
		params := operations.NewPutMicroservicesIDParams()
		str, err := getFromFile(c.String("file"))
		if err != nil {
			return fmt.Errorf("Error: unable to read file %s: %s", c.String("file"), err)
		}
		var d models.Microservice
		b := []byte(str)
		err = json.Unmarshal(b, &d)
		if err != nil {
			return fmt.Errorf("Error: wrong file format: %s", err)
		}
		params.Microservice = &d
		params.ID = c.String("id")
		resp, err := client.Operations.PutMicroservicesID(params)
		if err != nil {
			return fmt.Errorf("Error: put microservices failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "nodes":
		if c.String("id") == "" || c.String("file") == "" {
			return fmt.Errorf("Error: options --id and --file are required")
		}
		params := operations.NewPutNodesIDParams()
		str, err := getFromFile(c.String("file"))
		if err != nil {
			return fmt.Errorf("Error: unable to read file %s: %s", c.String("file"), err)
		}
		var d models.Node
		b := []byte(str)
		err = json.Unmarshal(b, &d)
		if err != nil {
			return fmt.Errorf("Error: wrong file format: %s", err)
		}
		params.Node = &d
		params.ID = c.String("id")
		resp, err := client.Operations.PutNodesID(params)
		if err != nil {
			return fmt.Errorf("Error: put nodes failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "regions":
		if c.String("id") == "" || c.String("file") == "" {
			return fmt.Errorf("Error: options --id and --file are required")
		}
		params := operations.NewPutRegionsIDParams()
		str, err := getFromFile(c.String("file"))
		if err != nil {
			return fmt.Errorf("Error: unable to read file %s: %s", c.String("file"), err)
		}
		var d models.Region
		b := []byte(str)
		err = json.Unmarshal(b, &d)
		if err != nil {
			return fmt.Errorf("Error: wrong file format: %s", err)
		}
		params.Region = &d
		params.ID = c.String("id")
		resp, err := client.Operations.PutRegionsID(params)
		if err != nil {
			return fmt.Errorf("Error: put regions failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "relationships":
		if c.String("id") == "" || c.String("file") == "" {
			return fmt.Errorf("Error: options --id and --file are required")
		}
		params := operations.NewPutRelationshipsIDParams()
		str, err := getFromFile(c.String("file"))
		if err != nil {
			return fmt.Errorf("Error: unable to read file %s: %s", c.String("file"), err)
		}
		var d models.Relationship
		b := []byte(str)
		err = json.Unmarshal(b, &d)
		if err != nil {
			return fmt.Errorf("Error: wrong file format: %s", err)
		}
		params.Relationship = &d
		params.ID = c.String("id")
		resp, err := client.Operations.PutRelationshipsID(params)
		if err != nil {
			return fmt.Errorf("Error: put relationships failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "externalendpoints":
		if c.String("id") == "" || c.String("file") == "" {
			return fmt.Errorf("Error: options --id and --file are required")
		}
		params := operations.NewPutExternalendpointsIDParams()
		str, err := getFromFile(c.String("file"))
		if err != nil {
			return fmt.Errorf("Error: unable to read file %s: %s", c.String("file"), err)
		}
		var d models.ExternalEndpoint
		b := []byte(str)
		err = json.Unmarshal(b, &d)
		if err != nil {
			return fmt.Errorf("Error: wrong file format: %s", err)
		}
		params.Externalendpoint = &d
		params.ID = c.String("id")
		resp, err := client.Operations.PutExternalendpointsID(params)
		if err != nil {
			return fmt.Errorf("Error: put external endpoints failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "dynamicnodes":
		if c.String("id") == "" || c.String("file") == "" {
			return fmt.Errorf("Error: options --id and --file are required")
		}
		params := operations.NewPutDynamicnodesIDParams()
		str, err := getFromFile(c.String("file"))
		if err != nil {
			return fmt.Errorf("Error: unable to read file %s: %s", c.String("file"), err)
		}
		var d models.DynamicNode
		b := []byte(str)
		err = json.Unmarshal(b, &d)
		if err != nil {
			return fmt.Errorf("Error: wrong file format: %s", err)
		}
		params.Dynamicnode = &d
		params.ID = c.String("id")
		resp, err := client.Operations.PutDynamicnodesID(params)
		if err != nil {
			return fmt.Errorf("Error: put dynamicnodes failed (%s)", err)
		}
		fmt.Printf("%s\n", resp.Error())
	default:
		return fmt.Errorf("Error: resource specificed (%s) is unkwnown", resource)
	}
	return nil
}

func handleDelete(c *cli.Context) error {
	schemes := []string{"http"}
	transport := httptransport.New(c.String("endpoint"), "/api/v2.0.0", schemes)
	client := apiclient.New(transport, strfmt.Default)
	switch resource := c.Args().Get(0); resource {
	case "applications":
		if c.String("id") == "" {
			return fmt.Errorf("Error: option --id is required")
		}
		params := operations.NewDeleteApplicationsIDParams()
		params.ID = c.String("id")
		resp, err := client.Operations.DeleteApplicationsID(params)
		if err != nil {
			return fmt.Errorf("Error: delete applications failed: %s", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "microservices":
		if c.String("id") == "" {
			return fmt.Errorf("Error: option --id is required")
		}
		params := operations.NewDeleteMicroservicesIDParams()
		params.ID = c.String("id")
		resp, err := client.Operations.DeleteMicroservicesID(params)
		if err != nil {
			return fmt.Errorf("Error: delete microservices failed: %s", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "nodes":
		if c.String("id") == "" {
			return fmt.Errorf("Error: option --id is required")
		}
		params := operations.NewDeleteNodesIDParams()
		params.ID = c.String("id")
		resp, err := client.Operations.DeleteNodesID(params)
		if err != nil {
			return fmt.Errorf("Error: delete nodes failed: %s", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "regions":
		if c.String("id") == "" {
			return fmt.Errorf("Error: option --id is required")
		}
		params := operations.NewDeleteRegionsIDParams()
		params.ID = c.String("id")
		resp, err := client.Operations.DeleteRegionsID(params)
		if err != nil {
			return fmt.Errorf("Error: delete regions failed: %s", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "relationships":
		if c.String("id") == "" {
			return fmt.Errorf("Error: option --id is required")
		}
		params := operations.NewDeleteRelationshipsIDParams()
		params.ID = c.String("id")
		resp, err := client.Operations.DeleteRelationshipsID(params)
		if err != nil {
			return fmt.Errorf("Error: delete relationships failed: %s", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "external endpoints":
		if c.String("id") == "" {
			return fmt.Errorf("Error: option --id is required")
		}
		params := operations.NewDeleteExternalendpointsIDParams()
		params.ID = c.String("id")
		resp, err := client.Operations.DeleteExternalendpointsID(params)
		if err != nil {
			return fmt.Errorf("Error: delete external endpoints failed: %s", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "dynamicnodes":
		if c.String("id") == "" {
			return fmt.Errorf("Error: option --id is required")
		}
		params := operations.NewDeleteDynamicnodesIDParams()
		params.ID = c.String("id")
		resp, err := client.Operations.DeleteDynamicnodesID(params)
		if err != nil {
			return fmt.Errorf("Error: delete dynamicnodes failed: %s", err)
		}
		fmt.Printf("%s\n", resp.Error())
	case "deployments":
		if c.String("id") == "" {
			return fmt.Errorf("Error: option --id is required")
		}
		params := operations.NewDeleteDeploymentsNameParams()
		params.Name = c.String("id")
		resp, err := client.Operations.DeleteDeploymentsName(params)
		if err != nil {
			return fmt.Errorf("Error: delete deployment failed: %s", err)
		}
		fmt.Printf("%s\n", resp.Error())
	default:
		return fmt.Errorf("Error: resource specificed (%s) is unkwnown", resource)
	}
	return nil
}

func handlePutAll(c *cli.Context) error {
	if c.String("file") == "" {
		return fmt.Errorf("Error: option --file is required")
	}

	schemes := []string{"http"}
	transport := httptransport.New(c.String("endpoint"), "/api/v2.0.0", schemes)
	client := apiclient.New(transport, strfmt.Default)

	conf := &confFile{}
	if err := parseYAML(c.String("file"), conf); err != nil {
		return err
	}

	for _, app := range conf.Applications {
		params := operations.NewPutApplicationsIDParams()
		params.Application = &app
		params.ID = app.ID
		resp, err := client.Operations.PutApplicationsID(params)
		if err != nil {
			fmt.Printf("error while sending request: %s", err)
			continue
		}
		if resp != nil {
			fmt.Println(resp.Error())
		}
	}

	for _, reg := range conf.Regions {
		params := operations.NewPutRegionsIDParams()
		params.Region = &reg
		params.ID = reg.ID
		resp, err := client.Operations.PutRegionsID(params)
		if err != nil {
			fmt.Printf("error while sending request: %s", err)
			continue
		}
		if resp != nil {
			fmt.Println(resp.Error())
		}
	}

	for _, depl := range conf.Deployments {
		params := operations.NewPutDeploymentsNameParams()
		params.Deployment = &depl
		params.Name = depl.Name
		resp, err := client.Operations.PutDeploymentsName(params)
		if err != nil {
			fmt.Printf("error while sending request: %s", err)
			continue
		}
		if resp != nil {
			fmt.Println(resp.Error())
		}
	}

	for _, ms := range conf.Microservices {
		params := operations.NewPutMicroservicesIDParams()
		params.Microservice = &ms
		params.ID = params.Microservice.Name
		resp, err := client.Operations.PutMicroservicesID(params)
		if err != nil {
			fmt.Printf("error while sending request: %s", err)
			continue
		}
		if resp != nil {
			fmt.Println(resp.Error())
		}
	}

	for _, node := range conf.Nodes {
		params := operations.NewPutNodesIDParams()
		params.Node = &node
		params.ID = node.ID
		resp, err := client.Operations.PutNodesID(params)
		if err != nil {
			fmt.Printf("error while sending request: %s", err)
			continue
		}
		if resp != nil {
			fmt.Println(resp.Error())
		}
	}

	for _, rel := range conf.Relationships {
		params := operations.NewPutRelationshipsIDParams()
		params.Relationship = &rel
		params.ID = rel.ID
		resp, err := client.Operations.PutRelationshipsID(params)
		if err != nil {
			fmt.Printf("error while sending request: %s", err)
			continue
		}
		if resp != nil {
			fmt.Println(resp.Error())
		}
	}

	for _, ee := range conf.ExternalEdpoints {
		params := operations.NewPutExternalendpointsIDParams()
		params.Externalendpoint = &ee
		params.ID = ee.ID
		resp, err := client.Operations.PutExternalendpointsID(params)
		if err != nil {
			fmt.Printf("error while sending request: %s", err)
			continue
		}
		if resp != nil {
			fmt.Println(resp.Error())
		}
	}

	for _, dnode := range conf.DynamicNodes {
		params := operations.NewPutDynamicnodesIDParams()
		params.Dynamicnode = &dnode
		params.ID = dnode.ID
		resp, err := client.Operations.PutDynamicnodesID(params)
		if err != nil {
			fmt.Printf("error while sending request: %s", err)
			continue
		}
		if resp != nil {
			fmt.Println(resp.Error())
		}
	}

	return nil
}

func getFromFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseYAML(filename string, obj *confFile) error {
	strData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	b := []byte(strData)
	if err := yaml.Unmarshal(b, obj); err != nil {
		return err
	}
	return nil
}

func handleDeleteAll(c *cli.Context) error {
	var resp interface{}
	var err error
	schemes := []string{"http"}
	transport := httptransport.New(c.String("endpoint"), "/api/v2.0.0", schemes)
	client := apiclient.New(transport, strfmt.Default)

	switch 	resource := c.Args().Get(0); resource {
	case "applications":
		params := operations.NewGetApplicationsParams()
		resp, err = client.Operations.GetApplications(params)
		if err != nil {
			return fmt.Errorf("Error: get applications failed: %s", err)
		}
		for _, app := range resp.(*operations.GetApplicationsOK).Payload.Applications {
			params := operations.NewDeleteApplicationsIDParams()
			params.ID = app.ID
			resp, err := client.Operations.DeleteApplicationsID(params)
			if err != nil {
				return fmt.Errorf("Error: delete applications failed: %s", err)
			}
			fmt.Printf("%s\n", resp.Error())
		}
	case "deployments":
		params := operations.NewGetDeploymentsParams()
		resp, err = client.Operations.GetDeployments(params)
		if err != nil {
			return fmt.Errorf("Error: get deployments failed: %s", err)
		}
		for _, depl := range resp.(*operations.GetDeploymentsOK).Payload.Deployments {
			params := operations.NewDeleteDeploymentsNameParams()
			params.Name = depl.Name
			resp, err := client.Operations.DeleteDeploymentsName(params)
			if err != nil {
				fmt.Printf("Error: delete deployment failed: %s", err)
			}
			fmt.Printf("%s\n", resp.Error())
		}

	case "microservices":
		params := operations.NewGetMicroservicesParams()
		resp, err = client.Operations.GetMicroservices(params)
		if err != nil {
			return fmt.Errorf("Error: get microservices failed: %s", err)
		}
		for _, ms := range resp.(*operations.GetMicroservicesOK).Payload.Microservices {
			params := operations.NewDeleteMicroservicesIDParams()
			params.ID = ms.ID
			resp, err := client.Operations.DeleteMicroservicesID(params)
			if err != nil {
				fmt.Printf("Error: delete microservices failed: %s", err)
			}
			fmt.Printf("%s\n", resp.Error())
		}

	case "nodes":
		params := operations.NewGetNodesParams()
		resp, err = client.Operations.GetNodes(params)
		if err != nil {
			return fmt.Errorf("Error: get nodes failed: %s", err)
		}
		for _, node := range resp.(*operations.GetNodesOK).Payload.Nodes {
			params := operations.NewDeleteNodesIDParams()
			params.ID = node.ID
			resp, err := client.Operations.DeleteNodesID(params)
			if err != nil {
				fmt.Printf("Error: delete nodes failed: %s", err)
			}
			fmt.Printf("%s\n", resp.Error())
		}

	case "regions":
		params := operations.NewGetRegionsParams()
		resp, err = client.Operations.GetRegions(params)
		if err != nil {
			return fmt.Errorf("Error: get regions failed: %s", err)
		}
		for _, reg := range resp.(*operations.GetRegionsOK).Payload.Regions {
			params := operations.NewDeleteRegionsIDParams()
			params.ID = reg.ID
			resp, err := client.Operations.DeleteRegionsID(params)
			if err != nil {
				fmt.Printf("Error: delete regions failed: %s", err)
			}
			fmt.Printf("%s\n", resp.Error())
		}

	case "relationships":
		params := operations.NewGetRelationshipsParams()
		resp, err = client.Operations.GetRelationships(params)
		if err != nil {
			return fmt.Errorf("Error: get relationships failed: %s", err)
		}
		for _, rel := range resp.(*operations.GetRelationshipsOK).Payload.Relationships {
			params := operations.NewDeleteRelationshipsIDParams()
			params.ID = rel.ID
			resp, err := client.Operations.DeleteRelationshipsID(params)
			if err != nil {
				fmt.Printf("Error: delete relationships failed: %s", err)
			}
			fmt.Printf("%s\n", resp.Error())
		}

	case "externalendpoints":
		params := operations.NewGetExternalendpointsParams()
		resp, err = client.Operations.GetExternalendpoints(params)
		if err != nil {
			return fmt.Errorf("Error: get external endpoints failed: %s", err)
		}
		for _, ee := range resp.(*operations.GetExternalendpointsOK).Payload.Externalendpoints {
			params := operations.NewDeleteExternalendpointsIDParams()
			params.ID = ee.ID
			resp, err := client.Operations.DeleteExternalendpointsID(params)
			if err != nil {
				fmt.Printf("Error: delete external endpoints failed: %s", err)
			}
			fmt.Printf("%s\n", resp.Error())
		}

	case "dynamicnodes":
		params := operations.NewGetDynamicnodesParams()
		if c.String("region_id") != "" {
			region_id := c.String("region_id")
			params.RegionID = &region_id
		}
		resp, err = client.Operations.GetDynamicnodes(params)
		if err != nil {
			return fmt.Errorf("Error: get dynamicnodes failed: %s", err)
		}
		for _, dnode := range resp.(*operations.GetDynamicnodesOK).Payload.Dynamicnodes {
			params := operations.NewDeleteDynamicnodesIDParams()
			params.ID = dnode.ID
			resp, err := client.Operations.DeleteDynamicnodesID(params)
			if err != nil {
				fmt.Printf("Error: delete dynamicnodes failed: %s", err)
			}
			fmt.Printf("%s\n", resp.Error())
		}

	default:
		return fmt.Errorf("Error: resource specificed (%s) is unkwnown", resource)
	}

	return nil
}

func printData(any interface{}) {
	if resp, ok := any.(*operations.GetApplicationsOK); ok {
		var data [][]string
		for _, app := range resp.Payload.Applications {
			msids := ""
			for _, ms := range app.Microservices {
				if msids == "" {
					msids = ms.MicroserviceID
				} else {
					msids = msids + "," + ms.MicroserviceID
				}
			}
			str := []string{app.ID, app.Name, app.Description, app.Status, msids}
			data = append(data, str)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Description", "Status", "Microservice Id"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetApplicationsIDOK); ok {
		var data [][]string
		app := resp.Payload
		msids := ""
		for _, ms := range app.Microservices {
			if msids == "" {
				msids = ms.MicroserviceID
			} else {
				msids = msids + "," + ms.MicroserviceID
			}
		}
		str := []string{app.ID, app.Name, app.Description, app.Status, msids}
		data = append(data, str)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Description", "Status", "Microservice Id"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetDeploymentsOK); ok {
		var data [][]string
		var datams [][]string
		var datadf [][]string
		for _, depl := range resp.Payload.Deployments {
			for _, ms := range depl.Microservices {
				strms := []string{depl.Name, ms.Name, ms.Description, ms.CPURequired, ms.MemoryRequired, ms.DiskRequired,
					ms.RegionID, ms.RegionRequired, strconv.FormatFloat(ms.PriceRequired, 'f', 2, 64),
					strconv.FormatFloat(ms.PriceComputed, 'f', 2, 64), ms.DeploymentDescriptor}
				datams = append(datams, strms)
			}
			for _, df := range depl.Dataflows {
				strdf := []string{depl.Name, df.SourceID, df.DestinationID, strconv.FormatInt(df.BandwidthRequired, 10),
					strconv.FormatInt(df.LatencyRequired, 10)}
				datadf = append(datadf, strdf)
			}
			str := []string{depl.Name, depl.Description, depl.Status, depl.ExternalendpointID}
			data = append(data, str)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Description", "Status", "ExternaEndpointID"})
		table.AppendBulk(data)
		table.Render()
		fmt.Printf("Microservices Requirements\n")
		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Depl. Name", "Name", "Description", "CPURequired", "MemoryRequired", "DiskRequired",
			"RegionID", "RegionRequired", "PriceRequired", "PriceComputed", "Deployment Descriptor"})
		table.AppendBulk(datams)
		table.Render()
		fmt.Printf("Dataflows\n")
		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Depl. Name", "SourceID", "DestinationID", "BandwidthRequired", "LatencyRequired"})
		table.AppendBulk(datadf)
		table.Render()
	}
	if resp, ok := any.(*operations.GetDeploymentsNameOK); ok {
		var data [][]string
		var datams [][]string
		var datadf [][]string
		depl := resp.Payload
		for _, ms := range depl.Microservices {
			strms := []string{depl.Name, ms.Name, ms.Description, ms.CPURequired, ms.MemoryRequired, ms.DiskRequired,
				ms.RegionID, ms.RegionRequired, strconv.FormatFloat(ms.PriceRequired, 'f', 2, 64),
				strconv.FormatFloat(ms.PriceComputed, 'f', 2, 64), ms.DeploymentDescriptor}
			datams = append(datams, strms)
		}
		for _, df := range depl.Dataflows {
			strdf := []string{depl.Name, df.SourceID, df.DestinationID, strconv.FormatInt(df.BandwidthRequired, 10),
				strconv.FormatInt(df.LatencyRequired, 10)}
			datadf = append(datadf, strdf)
		}
		str := []string{depl.Name, depl.Description, depl.Status, depl.ExternalendpointID}
		data = append(data, str)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Description", "Status", "ExternalEndpointID"})
		table.AppendBulk(data)
		table.Render()
		fmt.Printf("Microservices Requirements\n")
		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Depl. Name", "Name", "Description", "CPURequired", "MemoryRequired", "DiskRequired",
			"RegionID", "RegionRequired", "PriceRequired", "PriceComputed", "Deployment Descriptor"})
		table.AppendBulk(datams)
		table.Render()
		fmt.Printf("Dataflows\n")
		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Depl. Name", "SourceID", "DestinationID", "BandwidthRequired", "LatencyRequired"})
		table.AppendBulk(datadf)
		table.Render()
	}
	if resp, ok := any.(*operations.GetMicroservicesOK); ok {
		var data [][]string
		for _, ms := range resp.Payload.Microservices {
			str := []string{ms.ID, ms.Name, ms.Description, ms.ApplicationID, ms.NodeID, ms.RegionID, ms.Status}
			data = append(data, str)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Description", "ApplicationID", "NodeID", "RegionID", "Status"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetMicroservicesIDOK); ok {
		var data [][]string
		ms := resp.Payload
		str := []string{ms.ID, ms.Name, ms.Description, ms.ApplicationID, ms.NodeID, ms.RegionID, ms.Status}
		data = append(data, str)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Description", "ApplicationID", "NodeID", "RegionID", "Status"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetNodesOK); ok {
		var data [][]string
		for _, node := range resp.Payload.Nodes {
			str := []string{node.ID, node.Architecture, node.Version, node.Distribution, node.RegionID, node.CPUCapacity, node.CPUAvailable,
				node.MemoryCapacity, node.MemoryAvailable, node.DiskCapacity, node.DiskAvailable, node.Status}
			data = append(data, str)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Architecture", "Version", "Distribution", "RegionID", "CPUCapacity", "CPUAvailable",
			"MemoryCapacity", "MemoryAvailable", "DiskCapacity", "DiskAvailable", "Status"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetNodesIDOK); ok {
		var data [][]string
		node := resp.Payload
		str := []string{node.ID, node.Architecture, node.Version, node.Distribution, node.RegionID, node.CPUCapacity, node.CPUAvailable,
			node.MemoryCapacity, node.MemoryAvailable, node.DiskCapacity, node.DiskAvailable, node.Status}
		data = append(data, str)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Architecture", "Version", "Distribution", "RegionID", "CPUCapacity", "CPUAvailable",
			"MemoryCapacity", "MemoryAvailable", "DiskCapacity", "DiskAvailable", "Status"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetRegionsOK); ok {
		var data [][]string
		for _, reg := range resp.Payload.Regions {
			relids := ""
			for _, rel := range reg.Relationships {
				if relids == "" {
					relids = rel.RelationshipID
				} else {
					relids = relids + "," + rel.RelationshipID
				}
			}
			var cpuPrice, memPrice, diskPrice string
			if reg.Prices != nil {
				cpuPrice = strconv.FormatFloat(reg.Prices.CPU.MinPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.CPU.MaxPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.CPU.Scarcity, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.CPU.UnitPrice, 'f', 2, 64)
				memPrice = strconv.FormatFloat(reg.Prices.Memory.MinPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.Memory.MaxPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.Memory.Scarcity, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.Memory.UnitPrice, 'f', 2, 64)
				diskPrice = strconv.FormatFloat(reg.Prices.Disk.MinPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.Disk.MaxPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.Disk.Scarcity, 'f', 2, 64) + "," +
					strconv.FormatFloat(reg.Prices.Disk.UnitPrice, 'f', 2, 64)
			}
			str := []string{reg.ID, reg.Description, reg.Location, strconv.FormatInt(reg.Tier, 10), cpuPrice, memPrice, diskPrice, relids}
			data = append(data, str)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Description", "Location", "Tier", "CPUPrice", "MemPrice", "DiskPrice", "Relationship Id"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetRegionsIDOK); ok {
		var data [][]string
		reg := resp.Payload
		relids := ""
		for _, rel := range reg.Relationships {
			if relids == "" {
				relids = rel.RelationshipID
			} else {
				relids = relids + "," + rel.RelationshipID
			}
		}
		cpuPrice := strconv.FormatFloat(reg.Prices.CPU.MinPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.CPU.MaxPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.CPU.Scarcity, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.CPU.UnitPrice, 'f', 2, 64)
		memPrice := strconv.FormatFloat(reg.Prices.Memory.MinPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.Memory.MaxPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.Memory.Scarcity, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.Memory.UnitPrice, 'f', 2, 64)
		diskPrice := strconv.FormatFloat(reg.Prices.Disk.MinPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.Disk.MaxPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.Disk.Scarcity, 'f', 2, 64) + "," +
			strconv.FormatFloat(reg.Prices.Disk.UnitPrice, 'f', 2, 64)
		str := []string{reg.ID, reg.Description, reg.Location, strconv.FormatInt(reg.Tier, 10), cpuPrice, memPrice, diskPrice, relids}
		data = append(data, str)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Description", "Location", "Tier", "CPUPrice", "MemPrice", "DiskPrice", "Relationship Id"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetRelationshipsOK); ok {
		var data [][]string
		for _, rel := range resp.Payload.Relationships {
			var bwPrice, latPrice string
			if rel.Prices != nil {
				bwPrice = strconv.FormatFloat(rel.Prices.Bandwidth.MinPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(rel.Prices.Bandwidth.MaxPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(rel.Prices.Bandwidth.Scarcity, 'f', 2, 64) + "," +
					strconv.FormatFloat(rel.Prices.Bandwidth.UnitPrice, 'f', 2, 64)
				latPrice = strconv.FormatFloat(rel.Prices.Latency.MinPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(rel.Prices.Latency.MaxPrice, 'f', 2, 64) + "," +
					strconv.FormatFloat(rel.Prices.Latency.Scarcity, 'f', 2, 64) + "," +
					strconv.FormatFloat(rel.Prices.Latency.UnitPrice, 'f', 2, 64)
			}
			str := []string{rel.ID, rel.EndpointA, rel.EndpointB, rel.RegionID, strconv.FormatInt(rel.BandwidthCapacity, 10),
				strconv.FormatInt(rel.BandwidthAvailable, 10), strconv.FormatInt(rel.Latency, 10), bwPrice, latPrice, rel.Status}
			data = append(data, str)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "EndpointA", "EndpointB", "RegionID", "BandwidthCapacity", "BandwidthAvailable",
			"Latency", "BandwidthPrice", "LatencyPrice", "Status"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetRelationshipsIDOK); ok {
		var data [][]string
		rel := resp.Payload
		bwPrice := strconv.FormatFloat(rel.Prices.Bandwidth.MinPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(rel.Prices.Bandwidth.MaxPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(rel.Prices.Bandwidth.Scarcity, 'f', 2, 64) + "," +
			strconv.FormatFloat(rel.Prices.Bandwidth.UnitPrice, 'f', 2, 64)
		latPrice := strconv.FormatFloat(rel.Prices.Latency.MinPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(rel.Prices.Latency.MaxPrice, 'f', 2, 64) + "," +
			strconv.FormatFloat(rel.Prices.Latency.Scarcity, 'f', 2, 64) + "," +
			strconv.FormatFloat(rel.Prices.Latency.UnitPrice, 'f', 2, 64)
		str := []string{rel.ID, rel.EndpointA, rel.EndpointB, rel.RegionID, strconv.FormatInt(rel.BandwidthCapacity, 10),
			strconv.FormatInt(rel.BandwidthAvailable, 10), strconv.FormatInt(rel.Latency, 10), bwPrice, latPrice, rel.Status}
		data = append(data, str)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "EndpointA", "EndpointB", "RegionID", "BandwidthCapacity", "BandwidthAvailable",
			"Latency", "BandwidthPrice", "LatencyPrice", "Status"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetExternalendpointsOK); ok {
		var data [][]string
		for _, th := range resp.Payload.Externalendpoints {
			str := []string{th.ID, th.Name, th.Description, th.Type, th.Location, th.RegionID, th.IPAddress}
			data = append(data, str)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Description", "Type", "Location", "RegionID", "IPAddress"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetExternalendpointsIDOK); ok {
		var data [][]string
		th := resp.Payload
		str := []string{th.ID, th.Name, th.Description, th.Type, th.Location, th.RegionID, th.IPAddress}
		data = append(data, str)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Description", "Type", "Location", "RegionID", "IPAddress"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetDynamicnodesOK); ok {
		var data [][]string
		for _, dyn := range resp.Payload.Dynamicnodes {
			str := []string{dyn.ID, dyn.IPAddress, dyn.NodeID, dyn.RegionID}
			data = append(data, str)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "IPAddress", "NodeID", "RegionID"})
		table.AppendBulk(data)
		table.Render()
	}
	if resp, ok := any.(*operations.GetDynamicnodesIDOK); ok {
		var data [][]string
		dyn := resp.Payload
		str := []string{dyn.ID, dyn.IPAddress, dyn.NodeID, dyn.RegionID}
		data = append(data, str)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "IPAddress", "NodeID", "RegionID"})
		table.AppendBulk(data)
		table.Render()
	}

}
