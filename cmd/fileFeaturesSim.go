/*
* Copyright SRI International 2019-2022 All Rights Reserved.
* This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
 */

package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fileFeaturesSimCmd = &cobra.Command{
	Use:   "file-features-sim",
	Short: "print a list of similar structured files",
	Long:  `This subcommand prints a list of similarly structured files`,
	Run:   runFileFeaturesSimCmd,
}

func init() {
	rootCmd.AddCommand(fileFeaturesSimCmd)
	fileFeaturesSimCmd.Flags().String("subset", "", "subset string for file similarity")
	fileFeaturesSimCmd.MarkFlagRequired("subset")
}

type fileFeatureInfo struct {
	Name           string
	Size           int64
	Type           string
	Offset         int64
	RelativeOffset int64 `json:"relative_offset"`
}

type featureCount struct {
	Filename          string
	Type              string
	Count             int64
	TotalElementCount int64
}

type overview struct {
	Filename      string `json:"filename"`
	TotalFeatures int64  `json:"total_features"`
	Magic         string `json:"magic"`
	Length        int64  `json:"length"`
}

func runFileFeaturesSimCmd(cmd *cobra.Command, args []string) {
	subset, _ := cmd.Flags().GetString("subset")
	postgresConnHost := viper.Get("postgresConnHost")
	conn, err := pgx.Connect(context.Background(), postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())
	rows, err := conn.Query(context.Background(), allDocsQuery, "%"+subset+"%")
	if err != nil {
		log.Fatal(err)
	}

	var matches map[string][]overview = make(map[string][]overview)

	for rows.Next() {
		var doc string
		var filename string
		var featuresCount int64
		var featuresList []fileFeatureInfo
		var magic string
		var length int64
		err = rows.Scan(&doc, &filename, &featuresCount, &featuresList, &magic, &length)
		if err != nil {
			log.Fatal(err)
		}
		var o overview
		o.Filename = filename
		o.TotalFeatures = featuresCount
		o.Magic = magic
		o.Length = length
		structureHash := calculateHash(featuresList)
		if o.TotalFeatures == 0 {
			continue
		}
		_, ok := matches[structureHash]
		if !ok {
			matches[structureHash] = []overview{}
		}
		matches[structureHash] = append(matches[structureHash], o)
	}
	rows.Close()

	var matchesOrdered map[string][]overview = make(map[string][]overview)
	for k, v := range matches {
		if len(v) >= 2 {
			// add to another map with descriptive key names for in order printing
			matchesOrdered[fmt.Sprintf("%09d-%s", len(v), k)] = v
		}
		delete(matches, k)
	}
	outputJSON, err := json.MarshalIndent(matchesOrdered, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	filename := "structure_matches.json"
	err = ioutil.WriteFile(filename, outputJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(filename, " written to disk")
}

func calculateHash(featuresList []fileFeatureInfo) string {
	str := ""
	for _, e := range featuresList {
		str += e.Type
	}
	sha256 := sha256.Sum256([]byte(str))
	return fmt.Sprintf("%x", sha256)
}

var allDocsQuery = `
select 
	doc, 
	substring(doc from '(?:.+/)(.+)') as filename,
	jsonb_array_length(features_list) as features_count,
	features_list,
	magic,
	features->'length' as length
from file_features 
where doc like $1
order by jsonb_array_length(features_list) desc
`
