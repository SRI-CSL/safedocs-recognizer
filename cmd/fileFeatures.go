/*
* Copyright SRI International 2019-2022 All Rights Reserved.
* This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
 */

package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/jmoiron/sqlx"

	// use the pq driver for sqlx
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fileFeaturesCmd = &cobra.Command{
	Use:   "file-features",
	Short: "print a histogram of file features",
	Long:  `This subcommand prints a histogram of file features`,
	Run:   runFileFeaturesCmd,
}

func init() {
	rootCmd.AddCommand(fileFeaturesCmd)
	fileFeaturesCmd.Flags().String("file", "", "file for histogram")
	fileFeaturesCmd.MarkFlagRequired("file")
}

func runFileFeaturesCmd(cmd *cobra.Command, args []string) {
	file, _ := cmd.Flags().GetString("file")
	postgresConnHost := viper.Get("postgresConnHost")
	// docsURL := viper.Get("docsURL")
	db, err := sqlx.Open("postgres", postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Queryx(histoQuery, "%"+file+"%")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	first := true
	for rows.Next() {
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			log.Fatal(err)
		}
		if first {
			fmt.Println("filename: ", results["filename"])
			first = false
		}
		spacer := "     "
		count := strconv.FormatInt(results["count"].(int64), 10)

		fmt.Println(results["count"], spacer[0:len(spacer)-len(count)], results["feature"])
	}
}

var histoQuery = `
with file as (
	select features_list,features,doc from file_features where doc like $1 limit 1
)
select value->>'type' as feature,
	   count(*) as count,
	   substring((select doc from file) from '(?:.+/)(.+)') AS filename,
	   (select features from file) as features
from jsonb_array_elements
(
	(select features_list from file)
)
group by feature
order by count desc
`
