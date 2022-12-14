package Rummage

import (
	"bytes"
	"encoding/csv"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract"
)

const (
	getMessage    = "GET,x"
	StopCharacter = "\r\n\r\n"
)

//update the ipns record through the HLI server
func serverHLIUpdate(newcid string) error {

	addr := strings.Join([]string{serverIP, strconv.Itoa(serverPort)}, ":")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(err)
		return err
	}
	//defer conn.Close()
	_, err = conn.Write([]byte(setMessage(newcid)))
	_, err = conn.Write([]byte(StopCharacter))
	if err != nil {
		log.Println(err)
		return err
	}
	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(string(buff[:n]) + "...")
	err = conn.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

//set message for server HLI update
func setMessage(cid string) string {
	return "SET," + passWord + "," + cid
}

//function for extracting keywords from pdf with tesseract OCR
func ExtractPdfDataOCR(name string) ([]string, error) {
	var keywords string
	filename := name + ".pdf"
	doc, err := fitz.New(filename)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//defer doc.Close()

	client := gosseract.NewClient()
	//defer client.Close()

	// Extract pages as images
	for n := 0; n < doc.NumPage(); n++ {

		img, err := doc.Image(n)
		if err != nil {
			log.Println(err)
			continue
		}

		buf := new(bytes.Buffer)
		err = jpeg.Encode(buf, img, &jpeg.Options{400})
		if err != nil {
			log.Println(err)
			continue
		}

		imbyte := buf.Bytes()

		err = client.SetImageFromBytes(imbyte)
		if err != nil {
			log.Println(err)
			continue
		}
		text, _ := client.Text()
		keywords = keywords + " " + text

	}

	v := stopwords.CleanString(keywords, "en", false)
	s := strings.Split(v, " ")

	err = doc.Close()
	err = client.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return s, nil

}

//function to check if an extry exists in the HLI, returns the position (or where to insert)
func checkHLI1(entry string, records [][]string) (bool, string, int) {

	i := sort.Search(len(records), func(i int) bool { return entry <= records[i][0] })
	if i < len(records) && records[i][0] == entry {
		return true, records[i][1], i

	} else {
		return false, "", i
	}
}

//retrieves latest HLI data to use in indexing
func setHLI() ([][]string, error) {
	latestHLI, err := Shell.Resolve(HLI)
	if err != nil {
		log.Println(err)
		return nil, &noipns{HLI}
	}
	cidHLI := strings.Split(latestHLI, "s/")[1]

	cat, err := Shell.Cat(cidHLI)
	if err != nil {
		log.Println(err)
		return nil, &cIDmissing{cidHLI}
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

	err = f.Close()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return records, nil
}

//indexing from the server
//checks if entry exist in HLI
//check if entry for specific document exist in KWI, otherwise add
//adds to IPNS
func CreateIndexEntryServer1(data []string, cid string) error {

	TopLevelIndex, err := setHLI()
	change := false
	if err != nil {
		log.Println(err)
		return err
	}

	//index all words in data
	for _, s := range data {

		//remove empty or 1 letter entries
		if s == "" || len(s) == 1 {
			continue
		}

		//check if there is an index file available
		exist, indexCID, HLIposition := checkHLI1(s, TopLevelIndex)

		// if index file is available add to it
		// otherwise create one
		if exist {

			cat, err := Shell.Cat(indexCID)
			if err != nil {
				log.Println(err)
				continue
			}
			result, err := ioutil.ReadAll(cat)
			if err != nil {
				log.Println(err)
				continue
			}

			err = cat.Close()
			if err != nil {
				log.Println(err)
				continue
			}

			err = ioutil.WriteFile("./test_index/"+s+".csv", result, 0644)
			if err != nil {
				log.Println(err)
				continue
			}

			f, err := os.Open("./test_index/" + s + ".csv")
			if err != nil {
				log.Println(err)
				continue
			}

			reader := csv.NewReader(f)
			records, _ := reader.ReadAll()

			err = f.Close()
			if err != nil {
				log.Println(err)
				continue
			}

			//check if the entry for the CID is already in the index
			//if so, continue, otherwise, add to the sorted index
			i := sort.Search(len(records), func(i int) bool { return cid <= records[i][0] })
			if i < len(records) && records[i][0] == cid {
			} else {

				var entry = []string{
					cid, "pdf",
				}

				records = append(records, []string{""})
				copy(records[i+1:], records[i:])
				records[i] = entry

				//write new records to the file
				_ = os.Truncate("./test_index/"+s+".csv", 0)

				file, err := os.OpenFile("./test_index/"+s+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Println(err)
					continue
				}

				writer := csv.NewWriter(file)
				_ = writer.WriteAll(records)
				writer.Flush()

				err = file.Close()
				if err != nil {
					log.Println(err)
					continue
				}

				//add new sub-index file to ipfs
				k, err := os.OpenFile("./test_index/"+s+".csv", os.O_RDONLY, 0644)
				if err != nil {
					log.Println(err)
					continue
				}

				id, err := Shell.Add(k)
				if err != nil {
					log.Println(err)
					continue
				}
				TopLevelIndex[HLIposition][1] = id
				change = true
			}

		} else {
			f, err := os.Create("./test_index/" + s + ".csv")
			if err != nil {
				log.Println(err)
				continue
			}

			//var entry = [][]string{
			//	{cid, "pdf;" + time.Now().Format("2006-01-02 15:04")},
			//}
			var entry = [][]string{
				{cid, "pdf"},
			}

			writer := csv.NewWriter(f)
			err = writer.WriteAll(entry)
			if err != nil {
				log.Println(err)
				continue
			}
			err = f.Close()
			if err != nil {
				log.Println(err)
			}

			//add file to ipfs
			k, err := os.OpenFile("./test_index/"+s+".csv", os.O_RDONLY, 0644)
			if err != nil {
				log.Println(err)
				continue
			}

			id, err := Shell.Add(k)
			if err != nil {
				log.Println(err)
				continue
			}

			e := []string{s, id}

			TopLevelIndex = append(TopLevelIndex, []string{""})
			copy(TopLevelIndex[HLIposition+1:], TopLevelIndex[HLIposition:])
			TopLevelIndex[HLIposition] = e
			change = true
		}

	}

	if change != false {
		log.Println("Updating the HLI...")

		//safe new HLI and add to ipfs
		_ = os.Truncate("./HLI/HLI.csv", 0)
		f, err := os.OpenFile("./HLI/HLI.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		writer := csv.NewWriter(f)
		err = writer.WriteAll(TopLevelIndex)
		writer.Flush()
		err = f.Close()
		if err != nil {
			log.Println(err)
			return err
		}

		z, err := os.OpenFile("./HLI/HLI.csv", os.O_RDONLY, 0644)
		contentid, err := Shell.Add(z)
		if err != nil {
			log.Println(err)
		}

		err = Shell.Publish("", "/ipfs/"+contentid)
		if err != nil {
			log.Println(err)
			return err
		}

	}
	return nil
}

//indexing from the client
func CreateIndexEntryClient1(data []string, cid string) error {

	TopLevelIndex, err := setHLI()
	change := false
	if err != nil {
		log.Println(err)
		return err
	}

	//index all words in data
	for _, s := range data {

		//remove empty or 1 letter entries
		if s == "" || len(s) == 1 {
			continue
		}

		//check if there is an index file available
		exist, indexCID, HLIposition := checkHLI1(s, TopLevelIndex)

		// if index file is available add to it
		// otherwise create one
		if exist {

			cat, err := Shell.Cat(indexCID)
			if err != nil {
				log.Println(err)
				continue
			}
			result, err := ioutil.ReadAll(cat)
			if err != nil {
				log.Println(err)
				continue
			}

			err = cat.Close()
			if err != nil {
				log.Println(err)
				continue
			}

			err = ioutil.WriteFile("./test_index/"+s+".csv", result, 0644)
			if err != nil {
				log.Println(err)
				continue
			}

			f, err := os.Open("./test_index/" + s + ".csv")
			if err != nil {
				log.Println(err)
				continue
			}

			reader := csv.NewReader(f)
			records, _ := reader.ReadAll()

			err = f.Close()
			if err != nil {
				log.Println(err)
				continue
			}

			//check if the entry for the CID is already in the index
			//if so, continue, otherwise, add to the sorted index
			i := sort.Search(len(records), func(i int) bool { return cid <= records[i][0] })
			if i < len(records) && records[i][0] == cid {
			} else {
				var entry = []string{
					cid, "pdf",
				}

				records = append(records, []string{""})
				copy(records[i+1:], records[i:])
				records[i] = entry

				//write new records to the file
				_ = os.Truncate("./test_index/"+s+".csv", 0)

				file, err := os.OpenFile("./test_index/"+s+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Println(err)
					continue
				}
				writer := csv.NewWriter(file)
				_ = writer.WriteAll(records)
				writer.Flush()

				err = file.Close()
				if err != nil {
					log.Println(err)
					continue
				}

				//add new sub-index file to ipfs
				k, err := os.OpenFile("./test_index/"+s+".csv", os.O_RDONLY, 0644)
				if err != nil {
					log.Println(err)
					continue
				}

				id, err := Shell.Add(k)
				if err != nil {
					log.Println(err)
					continue
				}

				TopLevelIndex[HLIposition][1] = id
				change = true

			}

		} else {

			f, err := os.Create("./test_index/" + s + ".csv")
			if err != nil {
				log.Println(err)
				continue
			}

			//var entry = [][]string{
			//	{cid, "pdf;" + time.Now().Format("2006-01-02 15:04")},
			//}
			var entry = [][]string{
				{cid, "pdf"},
			}

			writer := csv.NewWriter(f)
			err = writer.WriteAll(entry)
			if err != nil {
				log.Println(err)
				continue
			}
			err = f.Close()
			if err != nil {
				log.Println(err)
				continue
			}

			//add file to ipfs
			k, err := os.OpenFile("./test_index/"+s+".csv", os.O_RDONLY, 0644)
			if err != nil {
				log.Println(err)
				continue
			}

			id, err := Shell.Add(k)
			if err != nil {
				log.Println(err)
				continue
			}

			e := []string{s, id}

			TopLevelIndex = append(TopLevelIndex, []string{""})
			copy(TopLevelIndex[HLIposition+1:], TopLevelIndex[HLIposition:])
			TopLevelIndex[HLIposition] = e
			change = true
		}

	}

	if change != false {
		//safe new HLI and add to ipfs
		_ = os.Truncate("./HLI/HLI.csv", 0)
		f, err := os.OpenFile("./HLI/HLI.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		writer := csv.NewWriter(f)
		err = writer.WriteAll(TopLevelIndex)
		writer.Flush()
		err = f.Close()
		if err != nil {
			log.Println(err)
			return err
		}

		z, err := os.OpenFile("./HLI/HLI.csv", os.O_RDONLY, 0644)
		contentid, err := Shell.Add(z)
		if err != nil {
			log.Println(err)
			return err
		}

		//from here will run server side
		err = serverHLIUpdate(contentid)
		if err != nil {
			log.Println(err)
			return err
		}

	}
	return nil
}
