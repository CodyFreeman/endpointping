package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type UrlStatus struct {
	url        string
	StatusCode int
}

type UrlStatusCollection struct {
	urls []UrlStatus
}

func (pc *UrlStatusCollection) addPath(path UrlStatus) {
	pc.urls = append(pc.urls, path)
}

func (pc *UrlStatusCollection) getPaths() []UrlStatus {
	return pc.urls
}

func eHandler(err error, msg string) {
	if err != nil {
		fmt.Println(msg)
		os.Exit(1)
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func getEndpointList(path string) []string {
	fh, err := os.Open(path)
	defer eHandler(fh.Close(), "Could not close file handler for reading endpoint list")
	eHandler(err, "Could not open input file")

	data, err := ioutil.ReadFile(path)
	eHandler(err, "Could not read input file")

	return strings.Split(string(data), ",")
}

func fuzzUrl(url string, client http.Client) UrlStatus {
	response, err := client.Get(url)
	defer eHandler(response.Body.Close(), "Could not close response")
	eHandler(err, fmt.Sprintf("Error connecting to: %s", url))

	return UrlStatus{url, response.StatusCode}
}

func fuzzUrls(urls []string) UrlStatusCollection {
	var wg sync.WaitGroup
	var collection UrlStatusCollection

	client := http.Client{
		Timeout: time.Second * 30,
	}

	wg.Add(len(urls))
	for _, url := range urls {
		go func(url string) {
			defer wg.Done()
			collection.addPath(fuzzUrl(url, client))
		}(url)
	}
	wg.Wait()
	return collection
}

func displayIntro() {

	fmt.Println("    _______           __               __         __   ")
	fmt.Println("   |    ___|.-----.--|  |.-----.-----.|__|.-----.|  |_ ")
	fmt.Println("   |    ___||     |  _  ||  _  |  _  ||  ||     ||   _|")
	fmt.Println("   |_______||__|__|_____||   __|_____||__||__|__||____|")
	fmt.Println("                         |__|")
	fmt.Println("                   ______ __")
	fmt.Println("                  |   __ \\__|.-----.-----.")
	fmt.Println("                  |    __/  ||     |  _  |")
	fmt.Println("                  |___|  |__||__|__|___  |")
	fmt.Println("                                   |_____|")

}

func displayHelp() {
	fmt.Println("    _   _  ____  __    ____    ")
	fmt.Println("   ( )_( )( ___)(  )  (  _ \\  ")
	fmt.Println("    ) _ (  )__)  )(__  )___/   ")
	fmt.Println("   (_) (_)(____)(____)(__)     ")
	fmt.Println("  ===========================  ")
	fmt.Println("")
	flag.PrintDefaults()
	fmt.Println("")
}

func parseFlags() (bool, string, string) {
	help := flag.Bool("h", false, "Help message")
	endpointListPath := flag.String("l", "", "Path to list of comma separated URLs")
	outputFilePath := flag.String("o", "result.txt", "Path to output file")

	flag.Parse()
	validateFlags(*help, *endpointListPath, *outputFilePath)

	return *help, *endpointListPath, *outputFilePath
}

func validateFlags(help bool, endpointListPath string, outputFilePath string) {
	// TODO: REWRITE TO SWITCH?
	if help {
		displayHelp()
		os.Exit(0)
	}

	if len(outputFilePath) <= 0 {
		fmt.Println("Please provide Path to output file")
	}

	if !fileExists(endpointListPath) {
		fmt.Println("Please provide path to list of endpoints. Use -h for help.")
		os.Exit(0)
	}
}

func writeOutput(collection UrlStatusCollection, outputFile string) {
	fh, err := os.Create(outputFile)
	eHandler(err, "Could not create output file")
	defer func() {
		err := fh.Close()
		eHandler(err, "Could not close file")
	}()

	for _, urlStatus := range collection.getPaths() {
		_, err := fh.WriteString(fmt.Sprintf("%v   -   %s\n", urlStatus.StatusCode, urlStatus.url))
		eHandler(err, "Failed to write line")
	}

	eHandler(fh.Sync(), "Could not sync with file")
}

func init() {
	displayIntro()
}

func main() {
	// Parsing flags
	_, endpointListPath, outputFilePath := parseFlags()
	// Checking urls
	results := fuzzUrls(getEndpointList(endpointListPath))
	// Writing output to file
	writeOutput(results, outputFilePath)
}