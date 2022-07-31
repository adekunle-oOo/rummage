# Rummage

Rummage is an open, collaborative, and decentralised search mechanism for IPFS. A rummage node is able to crawl content on IPFS and add this to the index, which itself is stored in a decentralised manner on IPFS.

Rummage currenly supports parsing PDF's and any HTML files stored on IPFS.

## How Rummage Works

Rummage divides the crawled data into two levels of indices.

- High Level Index (HLI)
- Key Word Index (KWI)

High Level Index stores the CID for the KWI of each keyword, then making this a highly indexable mesh. This data is constantly updated when a new craw is submitted to the network

We allow for two specific actions on the client

- Search
- Crawl

Search will be used to perform sophisticated searches, while Crawl will be used to submit crawl data by clients

## Installation

```bash
git clone https://github.com/adekunle-oOo/rummage.git
```

Install dependencies

```bash
sudo apt-get install g++
sudo apt-get install autoconf automake libtool
sudo apt-get install autoconf-archive
sudo apt-get install pkg-config
sudo apt-get install libpng-dev
sudo apt-get install libjpeg8-dev
sudo apt-get install libtiff5-dev
sudo apt-get install zlib1g-dev
wget http://www.leptonica.org/source/leptonica-1.81.1.tar.gz
sudo tar xf leptonica-1.81.1.tar.gz
cd leptonica-1.81.1 &&\
sudo ./configure &&\
sudo apt install make
sudo make &&\
sudo make install
sudo apt-get install tesseract-ocr # or sudo apt install tesseract-ocr
sudo apt install libtesseract-dev
```

Install Packages

```bash
go get -t github.com/otiai10/gosseract
go get github.com/ipfs/go-ipfs-api
go get github.com/wealdtech/go-ens/v3
go get github.com/otiai10/gosseract/v2
```

Build

```bash
sudo go build Rummage/CLI/.
```

Run (BETA)

```bash
./CLI
```
