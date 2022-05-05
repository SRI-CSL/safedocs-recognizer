/*
* Copyright SRI International 2019-2022 All Rights Reserved.
* This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
 */

package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type flatCfgDiff struct {
	Filename string   `json:"filename"`
	NewEdges []string `json:"new_edges"`
}

type ByNewEdges []flatCfgDiff

func (a ByNewEdges) Len() int           { return len(a) }
func (a ByNewEdges) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNewEdges) Less(i, j int) bool { return len(a[i].NewEdges) > len(a[j].NewEdges) }

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

var flattenedCfgModelCmd = &cobra.Command{
	Use:   "flat-cfg",
	Short: "generate a flat cfg model for the parser",
	Long:  `This subcommand generates a flat cfg model of the given parser`,
	Run:   runFlatCfgCmd,
}

var flattenedCfgDiffModelCmd = &cobra.Command{
	Use:   "flat-cfg-diff",
	Short: "diff non-baseline files with a flat-cfg model",
	Long:  `This subcommand calculates the difference for non-baseline files with a given flat-cfg model`,
	Run:   runFlatCfgDiffCmd,
}

func init() {
	rootCmd.AddCommand(flattenedCfgModelCmd)
	flattenedCfgModelCmd.Flags().String("parser", "", "parser from which to build the model")
	flattenedCfgModelCmd.MarkFlagRequired("parser")
	flattenedCfgModelCmd.Flags().String("universe", "", "mark the processing with a universe tag")
	flattenedCfgModelCmd.MarkFlagRequired("universe")

	rootCmd.AddCommand(flattenedCfgDiffModelCmd)
	flattenedCfgDiffModelCmd.Flags().String("parser", "", "parser from which to build the model")
	flattenedCfgDiffModelCmd.MarkFlagRequired("parser")
	flattenedCfgDiffModelCmd.Flags().String("model", "", "filename of model output by the flat-cfg command")
	flattenedCfgDiffModelCmd.MarkFlagRequired("model")
}

func runFlatCfgDiffCmd(cmd *cobra.Command, args []string) {
	parser, _ := cmd.Flags().GetString("parser")
	modelArg, _ := cmd.Flags().GetString("model")
	modelFile, err := os.Open(modelArg)
	if err != nil {
		log.Fatal(err)
	}
	defer modelFile.Close()
	model := make(map[string]int)
	scanner := bufio.NewScanner(modelFile)
	for scanner.Scan() {
		model[scanner.Text()] = 1
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	diffs := []flatCfgDiff{}
	postgresConnHost := viper.Get("postgresConnHost")
	conn, err := pgx.Connect(context.Background(), postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())
	rows, err := conn.Query(context.Background(), nonBaselineFlatCfgQuery, parser)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var filename string
		var cfg string
		err = rows.Scan(&filename, &cfg)
		if err != nil {
			log.Fatal(err)
		}
		diff := flatCfgDiff{}
		diff.Filename = filename
		graphAst, err := gographviz.ParseString(cfg)
		if err != nil {
			os.Stderr.WriteString(filename)
			os.Stderr.WriteString(err.Error())
			continue
		}
		graph := gographviz.NewGraph()
		err = gographviz.Analyse(graphAst, graph)
		if err != nil {
			os.Stderr.WriteString(filename)
			os.Stderr.WriteString(err.Error())
			continue
		}
		for _, e := range graph.Edges.Edges {
			srcNode := graph.Nodes.Lookup[e.Src]
			dstNode := graph.Nodes.Lookup[e.Dst]
			a := strings.Split(srcNode.Attrs["label"], "\\n")[0]
			b := strings.Split(dstNode.Attrs["label"], "\\n")[0]
			a = strings.Trim(a, "\"")
			b = strings.Trim(b, "\"")
			thisPath := a + "-->" + b
			if _, exists := model[thisPath]; !exists {
				exists := false
				for _, ne := range diff.NewEdges {
					if ne == thisPath {
						exists = true
						break
					}
				}
				if !exists {
					diff.NewEdges = append(diff.NewEdges, thisPath)
				}
			}
		}
		if len(diff.NewEdges) > 0 {
			diffs = append(diffs, diff)
		}
	}

	sort.Sort(ByNewEdges(diffs))
	output, err := json.MarshalIndent(diffs, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}

func runFlatCfgCmd(cmd *cobra.Command, args []string) {
	parser, _ := cmd.Flags().GetString("parser")
	universe, _ := cmd.Flags().GetString("universe")
	postgresConnHost := viper.Get("postgresConnHost")
	conn, err := pgx.Connect(context.Background(), postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())
	rows, err := conn.Query(context.Background(), flatCfgQuery, parser, universe)
	if err != nil {
		log.Fatal(err)
	}

	model := make(map[string]int)
	f, err := os.Create(parser + "_" + universe + "_flat_cfg_model.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	for rows.Next() {
		var filename string
		var cfg string
		err = rows.Scan(&filename, &cfg)
		if err != nil {
			log.Fatal(err)
		}
		graphAst, err := gographviz.ParseString(cfg)
		if err != nil {
			os.Stderr.WriteString(filename)
			os.Stderr.WriteString(err.Error())
			continue
		}
		graph := gographviz.NewGraph()
		err = gographviz.Analyse(graphAst, graph)
		if err != nil {
			os.Stderr.WriteString(filename)
			os.Stderr.WriteString(err.Error())
			continue
		}
		for _, e := range graph.Edges.Edges {
			srcNode := graph.Nodes.Lookup[e.Src]
			dstNode := graph.Nodes.Lookup[e.Dst]
			a := strings.Split(srcNode.Attrs["label"], "\\n")[0]
			b := strings.Split(dstNode.Attrs["label"], "\\n")[0]
			a = strings.Trim(a, "\"")
			b = strings.Trim(b, "\"")
			thisPath := a + "-->" + b
			if _, exists := model[thisPath]; !exists {
				model[thisPath] = 1
				_, err := w.WriteString(thisPath + "\n")
				check(err)
			}
		}
		log.Println(len(model))
	}
	w.Flush()
}

var flatCfgQuery = `
SELECT
	substring(doc from '(?:.+/)(.+)') as filename,
	cfg 
FROM consensus 
WHERE parser = $1 AND tag = $2 AND baseline = true AND status = 'valid'
`

var nonBaselineFlatCfgQuery = `
SELECT
	substring(doc from '(?:.+/)(.+)') as filename,
	cfg 
FROM consensus 
WHERE parser = $1 AND baseline = false
`
