package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kusumoto/bx-line-notice/config"

	"github.com/spf13/viper"
)

var myClient = &http.Client{Timeout: 10 * time.Second}
var cacheAPI = new([]BxJSONStructure)
var setDelay time.Duration = 5
var replaceLastData = false

var bxAPI = ""
var lineAccessToken = ""

// DiffModel is contain difference compare object
type DiffModel struct {
	OldValue BxJSONStructure
	NewValue BxJSONStructure
}

// BxJSONStructure is BX.in.th json return structure
type BxJSONStructure struct {
	Change            float64                `json:"change"`
	LastPrice         float64                `json:"last_price"`
	PairingID         int                    `json:"pairing_id"`
	PrimaryCurrency   string                 `json:"primary_currency"`
	SecondaryCurrency string                 `json:"secondary_currency"`
	Volume24hours     float64                `json:"volume_24hours"`
	Orderbook         OrderbookJSONStructure `json:"orderbook"`
}

// OrderbookJSONStructure is BX.in.th json return orderbook struture
type OrderbookJSONStructure struct {
	Asks BidsJSONStructure `json:"asks"`
	Bids BidsJSONStructure `json:"bids"`
}

// BidsJSONStructure is BX.in.th json return bid structure
type BidsJSONStructure struct {
	Highbid float64 `json:"highbid"`
	Total   float64 `json:"total"`
	Volume  float64 `json:"volume"`
}

// BxJSONObject is BX.in.th json mapper
type BxJSONObject map[string]BxJSONStructure

func getJSON(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func compareNewOrderAndCacheOrder(correctItem []BxJSONStructure) {
	if len(*cacheAPI) == 0 {
		splitTextBeforeSendToLine(firstReporter(correctItem))
	} else {
		splitTextBeforeSendToLine(compareDataFromAPIAndCache(correctItem))
	}
}

func firstReporter(correctItem []BxJSONStructure) []string {
	var stringCollections []string
	for _, bxItem := range correctItem {
		var stringBuffer bytes.Buffer
		stringBuffer.WriteString("\n")
		stringBuffer.WriteString(bxItem.PrimaryCurrency +
			" 👉🏻 " + bxItem.SecondaryCurrency +
			" \n🚂 Change : " + strconv.FormatFloat(bxItem.Change, 'f', -1, 64) +
			" \n💎 Last Price : " + strconv.FormatFloat(bxItem.LastPrice, 'f', -1, 64))
		stringCollections = append(stringCollections, stringBuffer.String())
	}
	return stringCollections
}

func compareDataFromAPIAndCache(correctItem []BxJSONStructure) []string {
	var compareResult = cacheDataAndAPIMapper(correctItem)
	var stringCollections []string
	for _, result := range compareResult {
		if result.OldValue.PairingID == result.NewValue.PairingID {
			if result.OldValue.LastPrice > result.NewValue.LastPrice {
				var stringBuffer bytes.Buffer
				stringBuffer.WriteString("\n\n")
				stringBuffer.WriteString(result.NewValue.PrimaryCurrency +
					" 👉🏻 " + result.NewValue.SecondaryCurrency +
					" \n🚂 Change : New ➡️ " + strconv.FormatFloat(result.NewValue.Change, 'f', -1, 64) +
					" Old ➡️ " + strconv.FormatFloat(result.OldValue.Change, 'f', -1, 64) +
					" \n✈️ Last Price : New ➡️ " + strconv.FormatFloat(result.NewValue.LastPrice, 'f', -1, 64) +
					" Old ➡️ " + strconv.FormatFloat(result.OldValue.LastPrice, 'f', -1, 64))
				stringCollections = append(stringCollections, stringBuffer.String())
			}
		}
	}
	return stringCollections
}

func cacheDataAndAPIMapper(correctItem []BxJSONStructure) []DiffModel {
	var compareResult = []DiffModel{}
	for _, bxItem := range correctItem {
		for _, bxCacheItem := range *cacheAPI {
			if bxItem.PairingID == bxCacheItem.PairingID {
				var resultItem = DiffModel{OldValue: bxCacheItem, NewValue: bxItem}
				compareResult = append(compareResult, resultItem)
			}
		}
	}
	return compareResult
}

func splitTextBeforeSendToLine(collections []string) {
	var collectionLen = len(collections)
	var stringBuffer bytes.Buffer
	for index := 0; index < collectionLen; index++ {
		stringBuffer.WriteString(collections[index])
		if index%15 == 0 && index != 1 && index != 0 || index+1 == collectionLen {
			sendToLineNotifiy(stringBuffer.String())
			stringBuffer.Reset()
		}
	}
}

func sendToLineNotifiy(message string) {
	data := url.Values{"message": {message}}
	r, _ := http.NewRequest("POST", "https://notify-api.line.me/api/notify", strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Authorization", "Bearer "+lineAccessToken)
	_, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err.Error())
	}
}

func readConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("json")
	viper.SetDefault("http_timeout", 10)
	viper.SetDefault("delay", 5)

	var config config.GeneralConfig

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	myClient = &http.Client{Timeout: config.HTTPTimeout * time.Second}
	setDelay = config.Delay
	bxAPI = config.BXAPIUrl
	lineAccessToken = config.LineAccessToken
	replaceLastData = config.ReplaceLastData
}

func replaceLowerData(correctItem []BxJSONStructure) []BxJSONStructure {
	var tempCache = []BxJSONStructure{}
	for _, bxItem := range correctItem {
		for _, bxCacheItem := range *cacheAPI {
			if bxItem.PairingID == bxCacheItem.PairingID && bxCacheItem.LastPrice > bxItem.LastPrice {
				tempCache = append(tempCache, bxItem)
			} else if bxItem.PairingID == bxCacheItem.PairingID {
				tempCache = append(tempCache, bxCacheItem)
			}
		}
	}
	return tempCache
}

func bxObjectConverter(correctItem *BxJSONObject) []BxJSONStructure {
	var bxResult = []BxJSONStructure{}
	for _, bxItem := range *correctItem {
		bxResult = append(bxResult, bxItem)
	}
	return bxResult
}

func main() {
	readConfig()
	for {
		var correctObj = new(BxJSONObject)
		err := getJSON(bxAPI, correctObj)
		if err != nil {
			log.Fatal(err)
			fmt.Println(err.Error())
		}
		var bxFinalObject = bxObjectConverter(correctObj)
		compareNewOrderAndCacheOrder(bxFinalObject)
		if replaceLastData {
			*cacheAPI = bxFinalObject
		} else {
			if len(*cacheAPI) == 0 {
				*cacheAPI = bxFinalObject
			} else {
				*cacheAPI = replaceLowerData(bxFinalObject)
			}
		}
		time.Sleep(setDelay * 1000 * time.Millisecond)
	}
}
