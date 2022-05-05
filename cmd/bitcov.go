/*
* Copyright SRI International 2019-2022 All Rights Reserved.
* This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
 */

package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"sort"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type bitcovDiff struct {
	Filename  string  `json:"filename"`
	NewHits   int64   `json:"new_hits"`
	MatchRate float64 `json:"match_rate"`
}

type ByMatchRate []bitcovDiff

func (a ByMatchRate) Len() int           { return len(a) }
func (a ByMatchRate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByMatchRate) Less(i, j int) bool { return a[i].MatchRate > a[j].MatchRate }

var bitcovVisCmd = &cobra.Command{
	Use:   "bitcov-vis",
	Short: "generate a visualization for non-baseline files",
	Long:  `This subcommand calculates the difference for non-baseline files with a given bitcov model and produces a visualization`,
	Run:   runBitcovVisCmd,
}

var bitcovDiffCmd = &cobra.Command{
	Use:   "bitcov-diff",
	Short: "diff non-baseline files with a bitcov model",
	Long:  `This subcommand calculates the difference for non-baseline files with a given bitcov model`,
	Run:   runBitCovModelDiffCmd,
}

var bitcovCmd = &cobra.Command{
	Use:   "bitcov",
	Short: "generate a bitcov model",
	Long:  `This subcommand generates a bitcov model given a parser`,
	Run:   runBitCovModelCmd,
}

func init() {
	rootCmd.AddCommand(bitcovCmd)
	bitcovCmd.Flags().String("parser", "", "parser from which to build the model")
	bitcovCmd.MarkFlagRequired("parser")
	bitcovCmd.Flags().String("universe", "", "mark the processing with a universe tag")
	bitcovCmd.MarkFlagRequired("universe")

	rootCmd.AddCommand(bitcovDiffCmd)
	bitcovDiffCmd.Flags().String("parser", "", "parser from which to build the model")
	bitcovDiffCmd.Flags().String("model", "", "filename of model output by the bitcov command (png)")
	bitcovDiffCmd.MarkFlagRequired("parser")
	bitcovDiffCmd.MarkFlagRequired("model")

	rootCmd.AddCommand(bitcovVisCmd)
	bitcovVisCmd.Flags().String("parser", "", "parser from which to build the model")
	bitcovVisCmd.Flags().String("model", "", "filename of model output by the bitcov command (png)")
	bitcovVisCmd.MarkFlagRequired("parser")
	bitcovVisCmd.MarkFlagRequired("model")
}

func runBitCovModelCmd(cmd *cobra.Command, args []string) {
	parser, _ := cmd.Flags().GetString("parser")
	universe, _ := cmd.Flags().GetString("universe")
	postgresConnHost := viper.Get("postgresConnHost")
	conn, err := pgx.Connect(context.Background(), postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())
	rows, err := conn.Query(context.Background(), baselineDocsQuery, parser, universe)
	if err != nil {
		log.Fatal(err)
	}
	var model []int
	covLines := 0
	for rows.Next() {
		var filename string
		var bitcov []byte
		err = rows.Scan(&filename, &bitcov)
		if err != nil {
			log.Fatal(err)
		}
		// parse png, if decompressed length not same, bail, 'or' everything into one pdf
		img, err := png.Decode(bytes.NewReader(bitcov))
		if err != nil {
			log.Print(filename)
			log.Println(err)
			continue
		}
		if covLines == 0 {
			covLines = img.Bounds().Max.X
			// initialize model with this parsers LoC length
			model = make([]int, covLines)
			for i := range model {
				model[i] = 1
			}
		} else if covLines != img.Bounds().Max.X {
			log.Fatal("bitcov length doesn't match for " + filename)
		}
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				c := img.At(x, y)
				r, g, b, _ := c.RGBA()
				if r == 0 && g == 0 && b == 0 {
					model[x] = 0
				}
			}
		}
	}
	//output model as png
	modelImg := image.NewGray(image.Rect(0, 0, covLines, 1))
	for y := 0; y < 1; y++ {
		for x := 0; x < covLines; x++ {
			var val uint8
			val = 255
			if model[x] == 0 {
				val = 0
			}
			modelImg.Set(x, y, color.Gray{
				Y: val,
			})
		}
	}

	filename := parser + "_" + universe + "_model.png"
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, modelImg); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	log.Println("wrote model to " + filename)
}

var baselineDocsQuery = `
SELECT 
	substring(doc from '(?:.+/)(.+)') as filename,
	bitcov
FROM consensus
WHERE baseline = true AND parser = $1 AND tag = $2 AND status = 'valid'
`

func runBitcovVisCmd(cmd *cobra.Command, args []string) {
	parser, _ := cmd.Flags().GetString("parser")
	modelArg, _ := cmd.Flags().GetString("model")
	postgresConnHost := viper.Get("postgresConnHost")
	conn, err := pgx.Connect(context.Background(), postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	// open model file
	modelFile, err := os.Open(modelArg)
	if err != nil {
		log.Fatal(err)
	}
	defer modelFile.Close()
	modelPng, err := png.Decode(bufio.NewReader(modelFile))
	if err != nil {
		log.Fatal(err)
	}
	model := make([]int, modelPng.Bounds().Max.X)

	for y := modelPng.Bounds().Min.Y; y < modelPng.Bounds().Max.Y; y++ {
		for x := modelPng.Bounds().Min.X; x < modelPng.Bounds().Max.X; x++ {
			c := modelPng.At(x, y)
			r, g, b, _ := c.RGBA()
			if r == 0 && g == 0 && b == 0 {
				model[x] = 0
			} else {
				model[x] = 1
			}
		}
	}

	// count := 0
	// conn.QueryRow(context.Background(), nonBaselineDocsCountQuery, parser).Scan(&count)
	// width := modelPng.Bounds().Max.X
	// height := count / 10
	// visImg := image.NewNRGBA(image.Rect(0, 0, width, height))

	rows, err := conn.Query(context.Background(), nonBaselineDocsQuery, parser)
	if err != nil {
		log.Fatal(err)
	}
	// diffVisRows := make([]int, 0)
	row := 0
	var newHitFiles int
	for rows.Next() {
		// highlight "location" of new hits
		// diffVis := make([]int, modelPng.Bounds().Max.X)
		// log.Println(modelPng.Bounds().Max.X)
		log.Println(row)
		// if row >= height {
		// 	break
		// }
		var filename string
		var bitcov []byte
		err = rows.Scan(&filename, &bitcov)
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(bitcov))
		if err != nil {
			log.Fatal(err)
		}
		if len(model) != img.Bounds().Max.X {
			log.Fatal("bitcov model length doesn't match for " + filename)
		}
		var newHits int64
		var matchingHits int64
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				c := img.At(x, y)
				r, g, b, _ := c.RGBA()
				// diffVis[x] = 1
				// val := 0
				// log.Println(r, g, b, a)
				if r == 0 && g == 0 && b == 0 {
					if model[x] != 0 {
						// this file hit a line of code that the model did not
						newHits++
						r = 65535
						// log.Println("new hit")
						// diffVis[x] = 0
						// val = 1
					}
					if model[x] == 0 {
						// match
						matchingHits++
					}
				}
				// visImg.Set(x, row, color.NRGBA{
				// 	R: uint8(r & 255),
				// 	G: uint8(g & 255),
				// 	B: uint8(b & 255),
				// 	A: uint8(a & 255),
				// })
			}
		}
		if newHits > 0 {
			newHitFiles++
			// for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			// 	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			// 		c := img.At(x, y)
			// 		r, g, b, _ := c.RGBA()
			// 		if r == 0 && g == 0 && b == 0 {
			// 			if model[x] != 0 {
			// 				// this file hit a line of code that the model did not
			// 				r = 65535
			// 			}
			// 		}
			// 		// visImg.Set(x, row, color.NRGBA{
			// 		// 	R: uint8(r & 255),
			// 		// 	G: uint8(g & 255),
			// 		// 	B: uint8(b & 255),
			// 		// 	A: uint8(a & 255),
			// 		// })
			// 	}
			// }
		}
		row++
		// diffVisRows = append(diffVisRows, diffVis...)
		// log.Println(len(diffVisRows))
	}

	width := modelPng.Bounds().Max.X
	height := newHitFiles
	visImg := image.NewNRGBA(image.Rect(0, 0, width, height))

	rows, err = conn.Query(context.Background(), nonBaselineDocsQuery, parser)
	if err != nil {
		log.Fatal(err)
	}
	row = 0
	for rows.Next() {
		// highlight "location" of new hits
		log.Println(row)
		var filename string
		var bitcov []byte
		err = rows.Scan(&filename, &bitcov)
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(bitcov))
		if err != nil {
			log.Fatal(err)
		}
		if len(model) != img.Bounds().Max.X {
			log.Fatal("bitcov model length doesn't match for " + filename)
		}
		var newHits int64
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				c := img.At(x, y)
				r, g, b, _ := c.RGBA()
				if r == 0 && g == 0 && b == 0 {
					if model[x] != 0 {
						// this file hit a line of code that the model did not
						newHits++
						r = 65535
					}
				}
			}
		}
		if newHits > 0 {
			for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
				for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
					c := img.At(x, y)
					r, g, b, a := c.RGBA()
					if r == 0 && g == 0 && b == 0 {
						if model[x] != 0 {
							// this file hit a line of code that the model did not
							r = 65535
						}
					}
					visImg.Set(x, row, color.NRGBA{
						R: uint8(r & 255),
						G: uint8(g & 255),
						B: uint8(b & 255),
						A: uint8(a & 255),
					})
				}
			}
			row++
		}
	}

	f, err := os.Create(parser + "_" + modelArg + "_vis.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, visImg); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func runBitCovModelDiffCmd(cmd *cobra.Command, args []string) {
	parser, _ := cmd.Flags().GetString("parser")
	modelArg, _ := cmd.Flags().GetString("model")
	postgresConnHost := viper.Get("postgresConnHost")
	conn, err := pgx.Connect(context.Background(), postgresConnHost.(string))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	// open model file
	modelFile, err := os.Open(modelArg)
	if err != nil {
		log.Fatal(err)
	}
	defer modelFile.Close()
	modelPng, err := png.Decode(bufio.NewReader(modelFile))
	if err != nil {
		log.Fatal(err)
	}
	model := make([]int, modelPng.Bounds().Max.X)

	for y := modelPng.Bounds().Min.Y; y < modelPng.Bounds().Max.Y; y++ {
		for x := modelPng.Bounds().Min.X; x < modelPng.Bounds().Max.X; x++ {
			c := modelPng.At(x, y)
			r, g, b, _ := c.RGBA()
			if r == 0 && g == 0 && b == 0 {
				model[x] = 0
			} else {
				model[x] = 1
			}
		}
	}

	diffs := []bitcovDiff{}
	modelHits := 0
	for x := range model {
		if model[x] == 0 {
			modelHits++
		}
	}

	rows, err := conn.Query(context.Background(), nonBaselineDocsQuery, parser)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var filename string
		var bitcov []byte
		err = rows.Scan(&filename, &bitcov)
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(bitcov))
		if err != nil {
			log.Fatal(err)
		}
		if len(model) != img.Bounds().Max.X {
			log.Fatal("bitcov model length doesn't match for " + filename)
		}
		var newHits int64
		var matchingHits int64
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				c := img.At(x, y)
				r, g, b, _ := c.RGBA()
				if r == 0 && g == 0 && b == 0 {
					if model[x] != 0 {
						// this file hit a line of code that the model did not
						newHits++
					}
					if model[x] == 0 {
						// match
						matchingHits++
					}
				}
			}
		}
		matchRate := float64(matchingHits) / float64(modelHits)
		// log.Println(filename + " new hits: " + strconv.FormatInt(newHits, 10) + ", match rate: " + strconv.FormatFloat(matchRate, 'f', 4, 64))
		var diff bitcovDiff
		diff.Filename = filename
		diff.NewHits = newHits
		diff.MatchRate = matchRate
		diffs = append(diffs, diff)
	}

	sort.Sort(ByMatchRate(diffs))
	output, err := json.MarshalIndent(diffs, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}

var nonBaselineDocsQuery = `
SELECT 
	substring(doc from '(?:.+/)(.+)') as filename,
	bitcov
FROM consensus
WHERE baseline = false AND parser = $1
`

var nonBaselineDocsCountQuery = `
SELECT count(*)
FROM consensus
WHERE baseline = false AND parser = $1
`
