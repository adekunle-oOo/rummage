package Rummage

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-cid"
	ipfsapi "github.com/ipfs/go-ipfs-api"
	"github.com/multiformats/go-multibase"
	"github.com/pkg/errors"
	"github.com/wealdtech/go-multicodec"
)

// defining the global variables used
var (
	Shell        *ipfsapi.Shell
	Client       *ethclient.Client
	HLI          string
	serverPort   int
	serverIP     string
	passWord     string
	latestHLICid string
	version      uint
)

// function to convert to base32 encoding
func b32Cid(bytes []byte) (string, error) {
	data, codec, err := multicodec.RemoveCodec(bytes)
	if err != nil {
		return "", err
	}
	codecName, err := multicodec.Name(codec)
	if err != nil {
		return "", err
	}

	if codecName == "ipfs-ns" {
		thisCID, err := cid.Parse(data)
		if err != nil {
			return "", errors.Wrap(err, "failed to parse CID")
		}
		str, err := thisCID.StringOfBase(multibase.Base32)
		if err != nil {
			return "", errors.Wrap(err, "failed to obtain base36 representation")
		}
		return fmt.Sprintf("ipfs://%s", str), nil
	}

	return "", fmt.Errorf("unknown codec name %s", codecName)
}

// simple way of ensuring we have pdf's
// this will become more sophisticated in future releases
func isValidPdf(stream []byte) bool {
	// fmt.Println(stream)

	l := len(stream)
	isHeaderValid := stream[0] == 0x25 && stream[1] == 0x50 && stream[2] == 0x44 && stream[3] == 0x46 //%PDF

	//(.%%EOF)
	isTrailerValid1 := stream[l-6] == 0xa && stream[l-5] == 0x25 && stream[l-4] == 0x25 &&
		stream[l-3] == 0x45 && stream[l-2] == 0x4f && stream[l-1] == 0x46
	if isHeaderValid && isTrailerValid1 {
		return true
	}

	//(.%%EOF.)
	isTrailerValid2 := stream[l-7] == 0xa && stream[l-6] == 0x25 && stream[l-5] == 0x25 && stream[l-4] == 0x45 &&
		stream[l-3] == 0x4f && stream[l-2] == 0x46 && stream[l-1] == 0xa
	if isHeaderValid && isTrailerValid2 {
		return true
	}

	//(.%%EOF.)
	isTrailerValid4 := stream[l-7] == 0xd && stream[l-6] == 0x25 && stream[l-5] == 0x25 && stream[l-4] == 0x45 &&
		stream[l-3] == 0x4f && stream[l-2] == 0x46 && stream[l-1] == 0xd
	if isHeaderValid && isTrailerValid4 {
		return true
	}

	//(..%%EOF..)
	isTrailerValid3 := stream[l-8] == 0xd && stream[l-7] == 0x25 && stream[l-6] == 0x25 && stream[l-5] == 0x45 && stream[l-4] == 0x4f &&
		stream[l-3] == 0x46 && stream[l-2] == 0xd && stream[l-1] == 0xa
	if isHeaderValid && isTrailerValid3 {
		return true
	}

	return false
}

// function to get the latest ipns record from the server for the HLI
func getHLI() (string, error) {
	addr := strings.Join([]string{serverIP, strconv.Itoa(serverPort)}, ":")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = conn.Write([]byte(getMessage))
	_, err = conn.Write([]byte(StopCharacter))
	if err != nil {
		return "", err
	}
	buff := make([]byte, 1024)
	n, _ := conn.Read(buff)
	data := string(buff[:n])
	HLI := strings.Trim(data, "\r\n\r\n")
	// log.Printf("Update HLI: %s",HLI)
	return HLI, nil
}

// checks if required directories are present, if not create them
func setDirectories() error {
	if _, err := os.Stat("./HLI"); os.IsNotExist(err) {
		err := os.Mkdir("./HLI", 0700)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat("./retrieved_files"); os.IsNotExist(err) {
		err := os.Mkdir("./retrieved_files", 0700)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat("./test_index"); os.IsNotExist(err) {
		err := os.Mkdir("./test_index", 0700)
		if err != nil {
			return err
		}
	} else {
		err := os.RemoveAll("./test_index/")
		err = os.Mkdir("./test_index", 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

// starts pinning new HLI cid and removes old entry
func updatePin(newHLICid string) {
	err := Shell.Pin(newHLICid)
	err = Shell.Unpin(latestHLICid)
	if err != nil {
		log.Println(err)
	}
	latestHLICid = newHLICid
}
