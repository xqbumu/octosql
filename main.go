package main

import (
	"context"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.com/cube2222/octosql/aggregates"
	"github.com/cube2222/octosql/datasources/json"
	"github.com/cube2222/octosql/execution"
	"github.com/cube2222/octosql/functions"
	"github.com/cube2222/octosql/logical"
	"github.com/cube2222/octosql/parser"
	"github.com/cube2222/octosql/physical"
	"github.com/cube2222/octosql/table_valued_functions"
	"github.com/xqbumu/sqlparser"
)

func main() {
	statement, err := sqlparser.Parse(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	logicalPlan, err := parser.ParseNode(statement.(sqlparser.SelectStatement))
	if err != nil {
		log.Fatal(err)
	}
	env := physical.Environment{
		Aggregates: map[string][]physical.AggregateDescriptor{
			"count": aggregates.CountOverloads,
		},
		Functions: map[string][]physical.FunctionDescriptor{
			"=": functions.Equals,
		},
		Datasources: &physical.DatasourceRepository{
			Datasources: map[string]func(name string) (physical.DatasourceImplementation, error){
				"json": json.Creator,
			},
		},
		TableValuedFunctions: map[string][]physical.TableValuedFunctionDescriptor{
			"tumble": table_valued_functions.Tumble,
		},
		PhysicalConfig:  nil,
		VariableContext: nil,
	}
	// TODO: Wrap panics into errors in subfunction.
	physicalPlan := logicalPlan.Typecheck(
		context.Background(),
		env,
		logical.Environment{
			CommonTableExpressions: map[string]physical.Node{},
		},
	)
	spew.Dump(physicalPlan.Schema)
	executionPlan, err := physicalPlan.Materialize(
		context.Background(),
		env,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := executionPlan.Run(
		execution.ExecutionContext{
			Context:         context.Background(),
			VariableContext: nil,
		},
		func(ctx execution.ProduceContext, record execution.Record) error {
			log.Println(record.String())
			return nil
		},
		func(ctx execution.ProduceContext, msg execution.MetadataMessage) error {
			log.Printf("%+v", msg)
			return nil
		},
	); err != nil {
		log.Fatal(err)
	}
}
