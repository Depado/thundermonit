package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	"github.com/samsarahq/thunder/graphql"
	_ "github.com/samsarahq/thunder/graphql/graphiql/statik"
	"github.com/samsarahq/thunder/graphql/introspection"
	"github.com/samsarahq/thunder/graphql/schemabuilder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Depado/thundermonit/cmd"
	"github.com/Depado/thundermonit/domain"
	graphqlabs "github.com/Depado/thundermonit/graphql"
	"github.com/Depado/thundermonit/storer"
	"github.com/Depado/thundermonit/uc"
)

// Build number and versions injected at compile time, set yours
var (
	Version = "unknown"
	Build   = "unknown"
)

// Main command that will be run when no other command is provided on the
// command-line
var rootCmd = &cobra.Command{
	Use:   "thundermonit",
	Short: "thundermonit",
	Run:   func(cmd *cobra.Command, args []string) { run() },
}

// Version command that will display the build number and version (if any)
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show build and version",
	Run:   func(cmd *cobra.Command, args []string) { fmt.Printf("Build: %s\nVersion: %s\n", Build, Version) },
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Generate and write the schema.json file on disk",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var output []byte
		requestHandler := graphqlabs.RequestHandler{Interactor: uc.NewInteractor(storer.NewMemoryStorer())}
		builder := schemabuilder.NewSchema()
		requestHandler.RegisterRepo(builder)
		requestHandler.RegisterService(builder)
		requestHandler.RegisterCI(builder)
		introspection.AddIntrospectionToSchema(builder.MustBuild())
		if output, err = introspection.ComputeSchemaJSON(*builder); err != nil {
			logrus.WithError(err).Fatal("Couldn't compute the schema to JSON")
		}
		if err = ioutil.WriteFile("schema.json", output, 0644); err != nil {
			logrus.WithError(err).Fatal("Couldn't write the schema on disk")
		}
		logrus.WithField("file", "schema.json").Info("Computed and built the schema")
	},
}

func run() {
	var err error
	var graphiqlfs http.FileSystem

	// Instantiate GraphiQL file system
	if graphiqlfs, err = fs.New(); err != nil {
		logrus.WithError(err).Fatal("Couldn't create GraphiQL file system")
	}

	fixtures := []*domain.Service{
		{
			ID: 0, Name: "goploader", URL: "https://gpldr.in", CI: &domain.CI{
				API: "https://drone.depado.eu",
				URL: "https://drone.depado.eu/Depado/goploader",
			}, Repo: &domain.Repo{Type: "github"},
		}, {
			ID: 1, Name: "gomonit", URL: "https://monit.depado.eu", CI: &domain.CI{
				API: "https://drone.depado.eu",
				URL: "https://drone.depado.eu/Depado/gomonit",
			}, Repo: &domain.Repo{Type: "github"},
		},
	}

	requestHandler := graphqlabs.RequestHandler{
		Interactor: uc.NewInteractor(
			storer.NewMemoryStorer(fixtures...),
		),
	}

	builder := schemabuilder.NewSchema()
	requestHandler.RegisterRepo(builder)
	requestHandler.RegisterService(builder)
	requestHandler.RegisterCI(builder)
	schema := builder.MustBuild()
	introspection.AddIntrospectionToSchema(schema)

	r := gin.Default()
	r.GET("/graphql", gin.WrapH(graphql.Handler(schema)))
	r.StaticFS("/graphiql/", graphiqlfs)
	if err = r.Run(":3030"); err != nil {
		logrus.WithError(err).Fatal("Couldn't start server")
	}
}

func main() {
	// Initialize Cobra and Viper
	cobra.OnInitialize(cmd.Initialize)
	cmd.AddLoggerFlags(rootCmd)
	cmd.AddConfigurationFlag(rootCmd)
	cmd.AddServerFlags(rootCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(schemaCmd)

	// Run the command
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("Couldn't start")
	}
}
