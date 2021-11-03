package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var consensusReportCmd = &cobra.Command{
	Use:   "consensus-report",
	Short: "generate a json report specifying file validity",
	Long:  `This tool uses a consensus vote for determining file validity`,
	Run:   runConsensusReportCmd,
}

func init() {
	rootCmd.AddCommand(consensusReportCmd)
	consensusReportCmd.Flags().String("subset", "", "subset string matching files")
	consensusReportCmd.MarkFlagRequired("subset")
	consensusReportCmd.Flags().Bool("baseline", false, "report on files considered baseline files or not")
	consensusReportCmd.Flags().Bool("stderr", false, "include STDERR in json report")
}

func runConsensusReportCmd(cmd *cobra.Command, args []string) {
	subset, _ := cmd.Flags().GetString("subset")
	baseline, _ := cmd.Flags().GetBool("baseline")
	stderr, _ := cmd.Flags().GetBool("stderr")
	postgresConnHost := viper.Get("postgresConnHost")
	conn, err := pgx.Connect(context.Background(), postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())
	rows, err := conn.Query(context.Background(), reportQueryDocFilter, baseline, "%"+subset+"%")
	if err != nil {
		log.Fatal(err)
	}

	var entries map[string]reportEntry
	entries = make(map[string]reportEntry)

	for rows.Next() {
		var result result
		err = rows.Scan(&result.Filename, &result.Doc, &result.Parser, &result.Status, &result.Stderr, &result.Digest)
		if err != nil {
			log.Fatal(err)
		}
		if _, ok := entries[result.Doc]; !ok {
			entry := reportEntry{}
			entry.Status = "valid"
			entries[result.Doc] = entry
		}
		entry := entries[result.Doc]
		entry.Digest = result.Digest
		entry.Testfile = result.Filename
		if result.Status == "rejected" {
			entry.ParserFailureCount++
			entry.Notes += result.Parser + " rejected doc, "
			if stderr {
				entry.ErrorDetail += result.Parser + "::\n" + result.Stderr + " \n"
			}
		}
		if result.Status == "valid" {
			entry.ParserValidCount++
		}
		// if entry.ParserFailureCount > 1 {
		// 	entry.Status = "rejected"
		// }
		entries[result.Doc] = entry
		// log.Println(result)
	}
	rows.Close()

	var reportEntries []reportEntry
	// var lackingQuorem []reportEntry
	for _, v := range entries {
		if v.ParserValidCount+v.ParserFailureCount < 4 {
			v.Notes += "one or more parsers didn't or failed to run against pdf"
		}
		if v.ParserValidCount <= v.ParserFailureCount {
			v.Status = "rejected"
		}
		// require a quorem before adding the result
		// if v.ParserValidCount+v.ParserFailureCount >= 3 {
		reportEntries = append(reportEntries, v)
		// } else {
		// lackingQuorem = append(lackingQuorem, v)
		// }
	}
	sort.Slice(reportEntries, func(i, j int) bool {
		return reportEntries[i].Testfile < reportEntries[j].Testfile
	})
	jsonReport, err := json.MarshalIndent(reportEntries, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	filename := "report.json"
	err = ioutil.WriteFile(filename, jsonReport, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(filename, " written to disk")

	// jsonReport, err = json.MarshalIndent(lackingQuorem, "", "  ")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// filename = "lackingQuorem.json"
	// err = ioutil.WriteFile(filename, jsonReport, 0644)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(filename, " written to disk")
}

type result struct {
	Filename string
	Doc      string
	Parser   string
	Status   string
	Stderr   string
	Digest   string
}

type reportEntry struct {
	Testfile           string `json:"testfile"`
	Status             string `json:"status"`
	Notes              string `json:"notes"`
	ErrorDetail        string `json:"errordetails"`
	Digest             string `json:"digest"`
	ParserFailureCount int    `json:"-"`
	ParserValidCount   int    `json:"-"`
}

var reportQueryDocFilter = `
SELECT substring(doc from '(?:.+/)(.+)') AS filename, 
	   doc, parser, status, stderr, digest  
FROM consensus
WHERE baseline = $1 AND doc LIKE $2 and parser in ('poppler', 'qpdf', 'caradoc', 'mupdf', 'poppler2009evaltwovanilla')
GROUP BY doc, parser, status, digest, stderr
ORDER BY doc, status
`
