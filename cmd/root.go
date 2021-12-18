package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/cube2222/octosql/aggregates"
	"github.com/cube2222/octosql/config"
	"github.com/cube2222/octosql/datasources/csv"
	"github.com/cube2222/octosql/datasources/json"
	"github.com/cube2222/octosql/datasources/plugins"
	"github.com/cube2222/octosql/execution"
	"github.com/cube2222/octosql/execution/nodes"
	"github.com/cube2222/octosql/functions"
	"github.com/cube2222/octosql/helpers/graph"
	"github.com/cube2222/octosql/logical"
	"github.com/cube2222/octosql/optimizer"
	"github.com/cube2222/octosql/outputs/batch"
	"github.com/cube2222/octosql/outputs/stream"
	"github.com/cube2222/octosql/parser"
	"github.com/cube2222/octosql/parser/sqlparser"
	"github.com/cube2222/octosql/physical"
	"github.com/cube2222/octosql/plugins/executor"
	"github.com/cube2222/octosql/plugins/manager"
	"github.com/cube2222/octosql/plugins/repository"
	"github.com/cube2222/octosql/table_valued_functions"
)

var emptyYamlNode = func() yaml.Node {
	var out yaml.Node
	if err := yaml.Unmarshal([]byte("{}"), &out); err != nil {
		fatalf("[BUG] Couldn't create empty yaml node: %s", err)
	}
	return out
}()

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "octosql",
	Args:  cobra.ExactArgs(1),
	Short: "",
	Long:  ``,
	Example: `octosql "SELECT * FROM myfile.json"
octosql "SELECT * FROM ` + "`mydir/myfile.json`" + `
octosql "SELECT * FROM plugins.plugins"`,
	Run: func(cmd *cobra.Command, args []string) {
		debug.SetGCPercent(1000)

		pluginManager := &manager.PluginManager{}

		pluginExecutor := executor.PluginExecutor{
			Manager: pluginManager,
		}
		defer func() {
			if err := pluginExecutor.Close(); err != nil {
				fatalf("couldn't close plugin executor: %s", err)
			}
		}()

		cfg, err := config.Read()
		if err != nil {
			fatalf("couldn't read config: %s", err)
		}

		installedPlugins, err := pluginManager.ListInstalledPlugins()
		if err != nil {
			fatalf("Couldn't list installed plugins: %s", err)
		}

		resolvedVersions := map[string]*semver.Version{}

		// Fill in plugin versions.
	dbLoop:
		for i := range cfg.Databases {
			if cfg.Databases[i].Version == nil {
				constraint, _ := semver.NewConstraint("*")
				cfg.Databases[i].Version = config.NewYamlUnmarshallableVersionConstraint(constraint)
			}
			for _, plugin := range installedPlugins {
				if plugin.Reference != cfg.Databases[i].Type {
					continue
				}
				for _, version := range plugin.Versions {
					if cfg.Databases[i].Version.Raw().Check(version.Number) {
						resolvedVersions[cfg.Databases[i].Name] = version.Number
						continue dbLoop
					}
				}
				break
			}
			fatalf("Database '%s' plugin '%s' used in configuration is not instaled with the required version.", cfg.Databases[i].Name, cfg.Databases[i].Type.String())
		}

		databases := make(map[string]func() (physical.Database, error))
		for _, dbConfig := range cfg.Databases {
			once := sync.Once{}
			curDbConfig := dbConfig
			var db physical.Database
			var err error

			databases[curDbConfig.Name] = func() (physical.Database, error) {
				once.Do(func() {
					db, err = pluginExecutor.RunPlugin(context.Background(), curDbConfig.Type, curDbConfig.Name, resolvedVersions[curDbConfig.Name], curDbConfig.Config)
				})
				if err != nil {
					return nil, fmt.Errorf("couldn't run %s plugin %s: %w", curDbConfig.Type, curDbConfig.Name, err)
				}
				return db, nil
			}
		}
		{
			once := sync.Once{}
			var repositories []repository.Repository
			var err error
			databases["plugins"] = func() (physical.Database, error) {
				once.Do(func() {
					repositories, err = repository.GetRepositories(context.Background())
				})
				if err != nil {
					return nil, fmt.Errorf("couldn't get repositories: %w", err)
				}
				return plugins.Creator(context.Background(), pluginManager, repositories)
			}
		}

		for _, metadata := range installedPlugins {
			if _, ok := databases[metadata.Reference.Name]; ok {
				continue
			}
			curMetadata := metadata

			once := sync.Once{}
			var db physical.Database
			var err error

			databases[curMetadata.Reference.Name] = func() (physical.Database, error) {
				once.Do(func() {
					db, err = pluginExecutor.RunPlugin(context.Background(), curMetadata.Reference, curMetadata.Reference.Name, metadata.Versions[0].Number, emptyYamlNode)
				})
				if err != nil {
					return nil, fmt.Errorf("couldn't run default plugin %s database: %w", curMetadata.Reference, err)
				}
				return db, nil
			}
		}

		statement, err := sqlparser.Parse(args[0])
		if err != nil {
			fatalf("couldn't parse query: %s", err)
		}
		logicalPlan, outputOptions, err := parser.ParseNode(statement.(sqlparser.SelectStatement), true)
		if err != nil {
			fatalf("couldn't parse query: %s", err)
		}
		env := physical.Environment{
			Aggregates: map[string][]physical.AggregateDescriptor{
				"array_agg":          aggregates.ArrayOverloads,
				"array_agg_distinct": aggregates.DistinctAggregateOverloads(aggregates.ArrayOverloads),
				"count":              aggregates.CountOverloads,
				"count_distinct":     aggregates.DistinctAggregateOverloads(aggregates.CountOverloads),
				"sum":                aggregates.SumOverloads,
				"sum_distinct":       aggregates.DistinctAggregateOverloads(aggregates.SumOverloads),
				"avg":                aggregates.AverageOverloads,
				"avg_distinct":       aggregates.DistinctAggregateOverloads(aggregates.AverageOverloads),
				"max":                aggregates.MaxOverloads,
				"min":                aggregates.MinOverloads,
			},
			Functions: functions.FunctionMap(),
			Datasources: &physical.DatasourceRepository{
				Databases: databases,
				FileHandlers: map[string]func(name string) (physical.DatasourceImplementation, physical.Schema, error){
					"json": json.Creator,
					"csv":  csv.Creator,
				},
			},
			PhysicalConfig:  nil,
			VariableContext: nil,
		}
		tableValuedFunctions := map[string]logical.TableValuedFunctionDescription{
			"max_diff_watermark": table_valued_functions.MaxDiffWatermark,
			"tumble":             table_valued_functions.Tumble,
			"range":              table_valued_functions.Range,
			"poll":               table_valued_functions.Poll,
		}
		uniqueNameGenerator := map[string]int{}
		var physicalPlan physical.Node
		var mapping map[string]string
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("typecheck error: %s\n", r)
					os.Exit(1)
				}
			}()
			physicalPlan, mapping = logicalPlan.Typecheck(
				context.Background(),
				env,
				logical.Environment{
					CommonTableExpressions: map[string]logical.CommonTableExpression{},
					TableValuedFunctions:   tableValuedFunctions,
					UniqueNameGenerator:    uniqueNameGenerator,
				},
			)
		}()
		reverseMapping := logical.ReverseMapping(mapping)

		var executionPlan execution.Node
		var orderByExpressions []execution.Expression
		var outSchema physical.Schema
		if describe {
			for i := range physicalPlan.Schema.Fields {
				physicalPlan.Schema.Fields[i].Name = reverseMapping[physicalPlan.Schema.Fields[i].Name]
			}
			executionPlan = &DescribeNode{
				Schema: physicalPlan.Schema,
			}
			outSchema = DescribeNodeSchema
			outputOptions.Limit = 0
			outputOptions.OrderByExpressions = nil
			outputOptions.OrderByDirections = nil
		} else {
			if optimize {
				physicalPlan = optimizer.Optimize(physicalPlan)
			}

			if explain >= 1 {
				file, err := os.CreateTemp(os.TempDir(), "octosql-describe-*.png")
				if err != nil {
					fatalf("couldn't create temporary file: %s", err)
				}
				os.WriteFile("describe.txt", []byte(graph.Show(physical.DescribeNode(physicalPlan, true)).String()), os.ModePerm)
				cmd := exec.Command("dot", "-Tpng")
				cmd.Stdin = strings.NewReader(graph.Show(physical.DescribeNode(physicalPlan, explain >= 2)).String())
				cmd.Stdout = file
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fatalf("couldn't render graph: %s", err)
				}
				if err := file.Close(); err != nil {
					fatalf("couldn't close temporary file: %s", err)
				}
				if err := open.Start(file.Name()); err != nil {
					fatalf("couldn't open graph: %s", err)
				}
				return
			}

			executionPlan, err = physicalPlan.Materialize(
				context.Background(),
				env,
			)
			if err != nil {
				fatalf("couldn't materialize physical plan: %s", err)
			}

			orderByExpressions := make([]execution.Expression, len(outputOptions.OrderByExpressions))
			for i := range outputOptions.OrderByExpressions {
				physicalExpr := outputOptions.OrderByExpressions[i].Typecheck(context.Background(), env.WithRecordSchema(physicalPlan.Schema), logical.Environment{
					CommonTableExpressions: map[string]logical.CommonTableExpression{},
					TableValuedFunctions:   tableValuedFunctions,
					UniqueVariableNames: &logical.VariableMapping{
						Mapping: mapping,
					},
					UniqueNameGenerator: uniqueNameGenerator,
				})
				execExpr, err := physicalExpr.Materialize(context.Background(), env.WithRecordSchema(physicalPlan.Schema))
				if err != nil {
					fatalf("couldn't materialize output order by expression with index %d: %v", i, err)
				}
				orderByExpressions[i] = execExpr
			}

			outFields := make([]physical.SchemaField, len(physicalPlan.Schema.Fields))
			copy(outFields, physicalPlan.Schema.Fields)
			outSchema = physical.Schema{
				Fields:    outFields,
				TimeField: physicalPlan.Schema.TimeField,
			}
			for i := range outFields {
				outFields[i].Name = reverseMapping[outFields[i].Name]
			}
		}

		var sink interface {
			Run(execCtx execution.ExecutionContext) error
		}

		switch os.Getenv("OCTOSQL_OUTPUT") {
		case "live_table":
			sink = batch.NewOutputPrinter(
				executionPlan,
				orderByExpressions,
				logical.DirectionsToMultipliers(outputOptions.OrderByDirections),
				outputOptions.Limit,
				outSchema,
				batch.NewTableFormatter,
				true,
			)
		case "batch_table":
			sink = batch.NewOutputPrinter(
				executionPlan,
				orderByExpressions,
				logical.DirectionsToMultipliers(outputOptions.OrderByDirections),
				outputOptions.Limit,
				outSchema,
				batch.NewTableFormatter,
				false,
			)
		case "stream_native":
			if len(orderByExpressions) > 0 {
				executionPlan = nodes.NewBatchOrderBy(
					executionPlan,
					orderByExpressions,
					logical.DirectionsToMultipliers(outputOptions.OrderByDirections),
				)
			}
			if outputOptions.Limit > 0 {
				fatalf("LIMIT clause not supported with stream output.")
			}

			sink = stream.NewOutputPrinter(
				executionPlan,
				stream.NewNativeFormat(outSchema),
			)
		default:
			sink = batch.NewOutputPrinter(
				executionPlan,
				orderByExpressions,
				logical.DirectionsToMultipliers(outputOptions.OrderByDirections),
				outputOptions.Limit,
				outSchema,
				batch.NewTableFormatter,
				true,
			)
		}

		if err := sink.Run(
			execution.ExecutionContext{
				Context:         context.Background(),
				VariableContext: nil,
			},
		); err != nil {
			fatalf("couldn't run query: %s", err)
		}
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

var describe bool
var explain int
var optimize bool

func init() {
	rootCmd.Flags().BoolVar(&describe, "describe", false, "Describe query output schema.")
	rootCmd.Flags().IntVar(&explain, "explain", 0, "Describe query output schema.")
	rootCmd.Flags().BoolVar(&optimize, "optimize", true, "Whether OctoSQL should optimize the query.")
}