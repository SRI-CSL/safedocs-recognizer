/*
* Copyright SRI International 2019-2022 All Rights Reserved.
* This material is based upon work supported by the Defense Advanced Research Projects Agency (DARPA) under Contract No. HR001119C0074.
 */

package cmd

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "run a process component ",
	Long:  `This subcommand runs a process component`,
	Run:   runProcessCmd,
}

func init() {
	rootCmd.AddCommand(processCmd)
	processCmd.Flags().String("subset", "", "document subset to process")
	processCmd.Flags().String("tag", "", "docker tag to run")
	processCmd.MarkFlagRequired("tag")
	processCmd.Flags().String("universe", "n/a", "mark the processing with a universe tag, defaults to n/a")
	processCmd.Flags().Bool("baseline", false, "consider results as part of baseline")
	processCmd.Flags().Int("processMax", -1, "only process a certain number at a time")
	processCmd.Flags().Int("workerCount", 8, "number of parallel workers")
}

var port atomic.Value

func startFileServer(dir string) {
	listener, err := net.Listen("tcp4", "0.0.0.0:0")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.Dir(dir)))
	log.Printf("Serving directory %s\n", dir)
	log.Println("Using port:", listener.Addr().(*net.TCPAddr).Port)
	port.Store(listener.Addr().(*net.TCPAddr).Port)
	err = http.Serve(listener, nil)
	log.Println(err)
}

type docResult struct {
	Doc string
}

func runProcessCmd(cmd *cobra.Command, args []string) {
	postgresConn := viper.Get("postgresConn").(string)
	postgresConnHost := viper.Get("postgresConnHost").(string)
	docsURL := viper.Get("docsURL").(string)
	subset, _ := cmd.Flags().GetString("subset")
	baseline, _ := cmd.Flags().GetBool("baseline")
	processMax, _ := cmd.Flags().GetInt("processMax")
	tag, _ := cmd.Flags().GetString("tag")
	workerCount, _ := cmd.Flags().GetInt("workerCount")
	universe, _ := cmd.Flags().GetString("universe")

	go startFileServer("./localdocs")
	time.Sleep(1 * time.Second)

	conn, err := pgx.Connect(context.Background(), postgresConnHost)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	jobCount := 0
	jobsAlreadyProcessed := 0

	docsProcessed := make(map[string]int)
	parser := tag[3:]
	rows, err := conn.Query(context.Background(), alreadyProcessed, parser, baseline)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var result docResult
		err = rows.Scan(&result.Doc)
		if err != nil {
			log.Fatal(err)
		}
		url, err := url.Parse(result.Doc)
		if err != nil {
			continue
		}
		docsProcessed[strings.TrimLeft(url.Path, "/")] = 1
	}
	rows.Close()

	fmt.Printf("%d already processed for %s\n", len(docsProcessed), parser)

	var docs []string
	docIndexURL := fmt.Sprintf("http://127.0.0.1:%d/sd_index.gz", port.Load().(int))
	resp, err := http.Get(docIndexURL)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fz, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer fz.Close()
	body, err := ioutil.ReadAll(fz)
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range strings.Split(string(body), "\n") {
		if _, ok := docsProcessed[line]; ok {
			continue
		}
		if strings.Contains(line, subset) {
			docs = append(docs, strings.Replace(line, " ", "%20", -1))
		}
	}
	fmt.Printf("%d docs set for processing\n", len(docs))

	numJobs := len(docs)
	jobs := make(chan BatchJob, numJobs)
	results := make(chan string, numJobs)
	for w := 0; w < workerCount; w++ {
		go worker(w, jobs, results)
	}

	containerBatchSize := 50
	for i := 0; i < len(docs); i += containerBatchSize {
		if processMax > -1 && jobCount >= processMax*containerBatchSize {
			log.Println("processMax: submitted maximum jobs, waiting for completion...")
			break
		}

		batchJob := BatchJob{}

		for j := 0; j < containerBatchSize; j++ {
			if len(docs) > i+j {
				mrDocURL := fmt.Sprintf("http://%s:%d/%s", docsURL, port.Load().(int), docs[i+j])
				batchJob.Meta.DocURL += mrDocURL + " "
			}
		}

		batchJob.Tag = tag
		batchJob.Meta.PostgresConn = postgresConn
		batchJob.Meta.IsBaseline = strconv.FormatBool(baseline)
		batchJob.Meta.Universe = universe

		jobs <- batchJob
		jobCount++
	}

	finishedJobs := 0
	if jobCount > 0 {
		for {
			<-results
			finishedJobs++
			log.Println(finishedJobs, "containers completed")
			if finishedJobs >= jobCount {
				break
			}
		}
	} else {
		log.Println("no jobs were queued")
		log.Println(jobsAlreadyProcessed, "jobs were not started because they have already been run to completion")
	}
}

// BatchJob batch job spec
type BatchJob struct {
	Payload string
	Tag     string
	Meta    Meta
}

// Meta metadata for job
type Meta struct {
	IsBaseline   string `json:"MR_IS_BASELINE"`
	DocURL       string `json:"MR_DOC_URL"`
	PostgresConn string `json:"MR_POSTGRES_CONN"`
	Universe     string `json:"MR_UNIVERSE"`
}

func worker(id int, jobs <-chan BatchJob, results chan<- string) {
	for j := range jobs {
		log.Println("worker: " + strconv.Itoa(id) + " took job for " + strconv.Itoa(len(strings.Split(j.Meta.DocURL, " "))) + " documents")
		cmd := exec.Command("docker", "run", "--add-host=host.docker.internal:host-gateway", "--rm",
			"-e", "MR_DOC_URL="+j.Meta.DocURL, "-e", "MR_POSTGRES_CONN="+j.Meta.PostgresConn,
			"-e", "MR_IS_BASELINE="+j.Meta.IsBaseline, "-e", "MR_UNIVERSE="+j.Meta.Universe,
			j.Tag)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}

		results <- j.Meta.DocURL
	}
}

var alreadyProcessed = `
SELECT doc FROM consensus WHERE parser = $1 and baseline = $2
`
