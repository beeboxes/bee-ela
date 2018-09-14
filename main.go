package main

import (
	"encoding/xml"
	"bytes"
	"context"
	"flag"
	//"io"
	"io/ioutil"

	"github.com/olivere/elastic"
	"golang.org/x/net/html/charset"
	//""
	"fmt"

	"encoding/json"
	"encoding/base64"
)

var (
	xmlFilename = flag.String("xml", "1.xml", "XML路径")
	//imgFilename = flag.String("img", "1.jpg", "图片路径")
	elasticServer = flag.String("ela", "http://localhost:9200", "Elastic服务器地址")
)

type Result struct {
	XMLName   xml.Name `xml:"SCAN_INFO" json:"-"`
	Name      string   `xml:"NAME"`
	Sex       string   `xml:"SEX"`
	Nation    string   `xml:"NATION"`
	Birthday  string   `xml:"BIRTHDAY"`
	Address   string   `xml:"ADDRESS"`
	Number    string   `xml:"NUMBER"`
	ScanTime  string   `xml:"SCANTIME"`
	HeadImage string   `xml:"HEADIMAGE"`
	CardImage string   `xml:"CARDIMAGE"`
	CardType  string   `xml:"CARDTYPE"`
}

const mapping = `
{
	"mappings":{
		"renyuan":{
			"properties":{
				"Name":{
					"type":"text"
				},
				"Sex":{
					"type":"text"
				},
				"Nation":{
					"type":"text"
				},
				"Birthday":{
					"type":   "date",
					"format": "yyyy-MM-dd"
				},
				"Address":{
					"type":"text"
				},
				"Number":{
					"type":"text"
				},
				"ScanTime":{
					"type":   "date",
					"format": "yyyy-MM-dd HH:mm:ss"
				},
				"HeadImage":{
					"type":"binary"
				},
				"CardImage":{
					"type":"binary"
				},
				"CardType":{
					"type":"text"
				}
			}
		}
	}
}`

func main() {
	flag.Parse()
	fmt.Println(*xmlFilename)
	//fmt.Println(*imgFilename)
	fmt.Println(*elasticServer)
	xmlBuffer, err := ioutil.ReadFile(*xmlFilename)

	//fmt.Println(xmlBuffer)
	if err != nil {
		fmt.Println(err.Error())
	}

	var ret Result
	decoder := xml.NewDecoder(bytes.NewReader(xmlBuffer))
	decoder.CharsetReader = charset.NewReaderLabel

	err = decoder.Decode(&ret)
	cardImage, err := ioutil.ReadFile(ret.CardImage)
	if err == nil {
		ret.CardImage = base64.StdEncoding.EncodeToString(cardImage)
	} else {
		panic(err)
	}

	headImage, err := ioutil.ReadFile(ret.HeadImage)
	if err == nil {
		ret.HeadImage = base64.StdEncoding.EncodeToString(headImage)
	} else {
		panic(err)
	}

	resultJSON, err := json.Marshal(ret)
	if err != nil {
		fmt.Println(err.Error())
	}

	//fmt.Println(string(resultJSON))

	client, err := elastic.NewClient(elastic.SetURL(*elasticServer))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	info, code, err := client.Ping(*elasticServer).Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Elasticsearch 返回码： %d 版本： %s\n", code, info.Version.Number)

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists("jiudian").Do(ctx)
	if err != nil {
		panic(err)
	}
	if !exists {
		createIndex, err := client.CreateIndex("jiudian").BodyString(mapping).Do(ctx)
		if err != nil {
			panic(err)
		}
		if !createIndex.Acknowledged {
			fmt.Println("创建索引失败")
		}
	} else {
		fmt.Println("索引已存在，跳过新建索引")
	}


	put2, err := client.Index().
		Index("jiudian").
		Type("renyuan").
		BodyString(string(resultJSON)).
		Do(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("人员索引 ID: %s, 索引名称: %s, 索引类型: %s\n", put2.Id, put2.Index, put2.Type)
}
