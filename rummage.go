package Rummage

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	ipfsapi "github.com/ipfs/go-ipfs-api"
)

//function to setup the local connections to ipfs, eth gateway etc.
//to be used at gateway server running the web interface
func ConnectServer(Infura string, HLI string) (*ipfsapi.Shell, *ethclient.Client) {
	err := setDirectories()
	if err != nil {
		log.Println(err)
	}
	sh := ipfsapi.NewShell("localhost:5001")
	cli, err := ethclient.Dial(Infura)
	if err != nil {
		log.Println(err)
	}
	HLI = HLI
	return sh, cli
}

//function to setup the local connections to ipfs, eth gateway, gateway server address etc.
//to be used by clients using the CLI application
func ConnectClient(Infura string, HLI string, ip string, port int, passW string) (*ipfsapi.Shell, *ethclient.Client) {
	err := setDirectories()
	if err != nil {
		log.Println(err)
	}
	sh := ipfsapi.NewShell("localhost:5001")
	cli, err := ethclient.Dial(Infura)
	if err != nil {
		log.Println(err)
	}
	serverPort = port
	serverIP = ip
	passWord = passW
	HLI, err = getHLI()
	if err != nil || HLI == "" {
		log.Println(err)
		//fallback to config file if server does not respond
		HLI = HLI
	}
	return sh, cli
}

//initiates a crawl for a name and content type t
//to be used on server (and therefore performs name publishing step locally)
//files are stored locally. Future releases will use tmpdir
func DoCrawlServer(name string, t string) error {

	//start crawling process
	d, id, err := crawlInput(t, name)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Success retrieving file...")

	err = ioutil.WriteFile("./retrieved_files/"+id+".pdf", d, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Success saving file locally...")

	//starts adding to the index by extracting keywords using OCR
	content, err := ExtractPdfDataOCR("./retrieved_files/" + id)
	if err != nil {
		log.Println(&pdfreadfail{"./retrieved_files/" + id + ".pdf"})
		return &pdfreadfail{"./retrieved_files/" + id + ".pdf"}
	}

	log.Println("Success extracting keywords...")

	err = CreateIndexEntryServer1(content, id)
	if err != nil {
		log.Println(&noIndexAdd{})
		return &noIndexAdd{}
	}

	log.Println("Successful indexing.")
	return nil
}

//initiates a crawl for a name and content type t
//to be used at client and uses the server to update the HLI entry on IPNS
//files are stored locally. Future releases will use tmpdir
func DoCrawlClient(name string, t string) error {

	//start crawling process
	d, id, err := crawlInput(t, name)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Success retrieving file...")

	err = ioutil.WriteFile("./retrieved_files/"+id+".pdf", d, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Success saving file locally...")

	//starts adding to the index by extracting keywords using OCR
	content, err := ExtractPdfDataOCR("./retrieved_files/" + id)
	if err != nil {
		log.Println(&pdfreadfail{"./retrieved_files/" + id + ".pdf"})
		return &pdfreadfail{"./retrieved_files/" + id + ".pdf"}
	}

	log.Println("Success extracting keywords...")

	err = CreateIndexEntryClient1(content, id)
	if err != nil {
		log.Println(&noIndexAdd{})
		return &noIndexAdd{}
	}

	log.Println("Successful indexing.")
	return nil
}

//search to be used from the gateway server
func DoSearch1(query string) ([]QueryResult, error) {

	searchTerms := strings.Split(query, " ")

	//get the latest HLI file
	latestHLI, err := Shell.Resolve(HLI)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	cidHLI := strings.Split(latestHLI, "s/")[1]

	//retrieve the HLI file
	cat, err := Shell.Cat(cidHLI)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result, err := ioutil.ReadAll(cat)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = cat.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	err = ioutil.WriteFile("./HLI/HLI.csv", result, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	f, err := os.Open("./HLI/HLI.csv")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	csvr := csv.NewReader(f)
	records, _ := csvr.ReadAll()

	//retrieve the index location for each keyword
	var indexLocations []string

	for j := 0; j < len(searchTerms); j++ {
		i := sort.Search(len(records), func(i int) bool { return searchTerms[j] <= records[i][0] })
		if i < len(records) && records[i][0] == searchTerms[j] {
			indexLocations = append(indexLocations, records[i][1])
		} else {
			indexLocations = append(indexLocations, "-")
		}
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	qResult, err := resultsWordServer1(searchTerms, indexLocations)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return qResult, nil
}

//search to be used from the client
func DoSearchClient(searchTerms []string) error {

	//get the latest HLI file
	latestHLI, err := Shell.Resolve(HLI)
	if err != nil {
		log.Println(err)
		return err
	}

	cidHLI := strings.Split(latestHLI, "s/")[1]

	//retrieve the HLI file
	cat, err := Shell.Cat(cidHLI)
	if err != nil {
		log.Println(err)
		return err
	}

	go updatePin(cidHLI)

	result, err := ioutil.ReadAll(cat)
	if err != nil {
		log.Println(err)
		return err
	}

	err = cat.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	err = ioutil.WriteFile("./HLI/HLI.csv", result, 0644)
	if err != nil {
		log.Println(err)
		return err
	}

	f, err := os.Open("./HLI/HLI.csv")
	if err != nil {
		log.Println(err)
		return err
	}

	csvr := csv.NewReader(f)
	records, _ := csvr.ReadAll()

	//retrieve the index location for each keyword
	var indexLocations []string

	for j := 0; j < len(searchTerms); j++ {
		i := sort.Search(len(records), func(i int) bool { return searchTerms[j] <= records[i][0] })
		if i < len(records) && records[i][0] == searchTerms[j] {
			indexLocations = append(indexLocations, records[i][1])
		} else {
			indexLocations = append(indexLocations, "-")
		}
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
		return err
	}

	qResult, err := resultsWordServer1(searchTerms, indexLocations)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println("Found", len(qResult), "results: ")

	//print out the results of the searchterms
	printKeyword := " "
	for k := 0; k < len(qResult); k++ {
		if printKeyword != qResult[k].SearchTerm {
			fmt.Println()
			fmt.Println(qResult[k].SearchTerm + " : ")
		}
		fmt.Println(qResult[k].CID + " (https://gateway.ipfs.io/ipfs/" + qResult[k].CID + ")")
	}

	return nil
}
