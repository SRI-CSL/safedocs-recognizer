package cmd

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"

	// use the pq driver for sqlx
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var schemaStatsCmd = &cobra.Command{
	Use:   "schema-stats",
	Short: "print information about the recognizer schema",
	Long:  `This subcommand prints a summary of each table`,
	Run:   runSchemaStatsCmd,
}

func init() {
	rootCmd.AddCommand(schemaStatsCmd)
}

func runSchemaStatsCmd(cmd *cobra.Command, args []string) {
	postgresConnHost := viper.Get("postgresConnHost")
	// docsURL := viper.Get("docsURL")
	db, err := sqlx.Open("postgres", postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Queryx(schemaStatsQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(results["relation"], " ", results["total_size"])
	}
}

var schemaStatsQuery = `
SELECT nspname || '.' || relname AS "relation",
    pg_size_pretty(pg_total_relation_size(C.oid)) AS "total_size"
  FROM pg_class C
  LEFT JOIN pg_namespace N ON (N.oid = C.relnamespace)
  WHERE nspname NOT IN ('pg_catalog', 'information_schema')
    AND C.relkind <> 'i'
    AND nspname !~ '^pg_toast'
  ORDER BY pg_total_relation_size(C.oid) DESC;
`
