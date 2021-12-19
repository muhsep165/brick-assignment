package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

type description struct{
	Title 				string `json:"title"`
	SubTitle 			string `json:"subtitle"`
	Applink				string `json:"applink"`
    ShowAtFront 		bool `json:"showAtFront"`
    IsAnnotation 		bool `json:"isAnnotation"`
    TypeName 			string `json:"__typename"`
}

type urlProduct struct {
	Url 				string `json:"url"`
	Id 					int64 `json:"id"`
}

type images struct {
	UrlThumbnail			string `json:"URLThumbnail"`
}

type ProductList struct {
	Products   			[]Products `json:"products"`
}
type Products struct {
	Type 				string `json:"type"`
	Generated			bool `json:"generated"`
	Id					string `json:"id"`
}

type productIdentifier struct {
	ProductId			string `json:"productID"`
	Price 				int64 `json:"price"`
	ProductName			string `json:"productName"`
	ProductURL			string `json:"productURL"`
}

func main (){
	var urlProducts []string
	var count, page int
	col := colly.NewCollector(
		colly.AllowedDomains("tokopedia.com", "www.tokopedia.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	col.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("accept", "application/json, text/plain, */*")
		r.Headers.Set("accept-encoding", "gzip, deflate, br")
		r.Headers.Set("accept-language", "en-US,en;q=0.9")
		r.Headers.Set("Referer", "https://www.tokopedia.com/")
		if r.Ctx.Get("gis") != "" {
			gis := fmt.Sprintf("%s:%s", r.Ctx.Get("gis"), r.Ctx.Get("variables"))
			h := md5.New()
			h.Write([]byte(gis))
			gisHash := fmt.Sprintf("%x", h.Sum(nil))
			r.Headers.Set("X-Instagram-GIS", gisHash)
		}
		fmt.Println("Visiting : ", r.URL.String())
	})
	
	col.OnHTML("html", func(e *colly.HTMLElement) {
		d := col.Clone()
		
		requestIDURL := e.Request.AbsoluteURL(e.ChildAttr(`link[as="script"]`, "href"))
		d.Visit(requestIDURL)
		
		dat := e.ChildText("body")
		dat = strings.Replace(dat[strings.Index(dat, "window.__cache=") : strings.Index(dat, "}};")] + "}}", "window.__cache=", "", -1)
		dataProducts := dat[strings.Index(dat, "\"products\":[") : len(dat)-1]
		dataProducts = "{" + dataProducts[strings.Index(dataProducts, "\"products\":[") : strings.Index(dataProducts, "]")] + "]}"

		var c ProductList
		err := json.Unmarshal([]byte(dataProducts), &c)
		if err != nil {
			log.Fatalf("Error Procces :  : %q", err)
		}
		for _, j := range c.Products {
			if count == 100 {
				break
			}

			idProduct := strings.Replace(j.Id, "AceSearchProduct", "", -1)
			productUrl := dat[strings.Index(dat, "\"id\":" + idProduct) : len(dat)-1]
			productUrl = "{" + productUrl[strings.Index(productUrl, "\"id\":" + idProduct) : strings.Index(productUrl, "?whid")] + "\"}"
			var urlprod urlProduct
			err := json.Unmarshal([]byte(productUrl), &urlprod)
			if err != nil {
				continue
			}

			urlProducts = append(urlProducts, urlprod.Url)
			count++
		}
		if count == 100 {
			detailProduct(urlProducts)
		}
	})

	
	col.OnError(func(r *colly.Response, e error) {
		log.Println("error:", e, r.Request.URL, string(r.Body))
	})

	if page == 0{
		page = 1
	}

	col.Visit("https://www.tokopedia.com/p/handphone-tablet/handphone?page=" + strconv.Itoa(page))

	if count < 100 {
		page = page + 1
		col.Visit("https://www.tokopedia.com/p/handphone-tablet/handphone?page=" + strconv.Itoa(page))
	}

}

func detailProduct(products []string) {
	fileName := "result-data.csv"
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error Procces :  : %q", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	col := colly.NewCollector(
		colly.AllowedDomains("tokopedia.com", "www.tokopedia.com"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	col.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("accept", "application/json, text/plain, */*")
		r.Headers.Set("accept-encoding", "gzip, deflate, br")
		r.Headers.Set("accept-language", "en-US,en;q=0.9")
		r.Headers.Set("Referer", "https://www.tokopedia.com/")
		if r.Ctx.Get("gis") != "" {
			gis := fmt.Sprintf("%s:%s", r.Ctx.Get("gis"), r.Ctx.Get("variables"))
			h := md5.New()
			h.Write([]byte(gis))
			gisHash := fmt.Sprintf("%x", h.Sum(nil))
			r.Headers.Set("X-Instagram-GIS", gisHash)
		}
		fmt.Println("Visiting : ", r.URL.String())
	})
	
	col.OnHTML("html", func(e *colly.HTMLElement) {
		d := col.Clone()
		
		requestIDURL := e.Request.AbsoluteURL(e.ChildAttr(`link[as="script"]`, "href"))
		d.Visit(requestIDURL)
		
		dat := e.ChildText("body")
		
		jsonData := strings.Replace(dat[strings.Index(dat, "window.__cache=") : strings.Index(dat, "}};")] + "}}", "window.__cache=", "", -1)

		productId := strings.Replace(jsonData[strings.Index(jsonData, "&productID=") : strings.Index(jsonData, "&productURL=")], "&productID=", "", -1)

		productNameString := jsonData[strings.Index(jsonData, "\"productID\":\""+productId+"\",\"price\"") : len(jsonData)-1]
		
		productNameString = "{" + productNameString[0 : strings.Index(productNameString, "\"typename\":\"pdpProductVariantPicture\"")] + "\"typename\":\"pdpProductVariantPicture\"}}"
		var productName productIdentifier
		err := json.Unmarshal([]byte(productNameString), &productName)
		if err != nil {
			log.Fatalf("Error Procces :  : %q", err)
		}

		rating := strings.Replace(jsonData[strings.Index(jsonData, "\"rating\"") : strings.Index(jsonData, "\"__typename\":\"pdpStats\"")], "\"rating\":", "", 1)

		shopName := strings.Replace(jsonData[strings.Index(jsonData, "\"shopName\":") : strings.Index(jsonData, "\"minOrder\":")], "\"shopName\":", "", -1)

		image := "{" + strings.Replace(jsonData[strings.Index(jsonData,  "\"URLThumbnail\":") : strings.Index(jsonData,  "\"videoURLAndroid\"")], ",", "", -1) + "}"
		var img images
		err = json.Unmarshal([]byte(image), &img)
		if err != nil {
			log.Fatalf("Error Procces :  : %q", err)
		}

		jsonDesc := jsonData[strings.Index(jsonData, "\"title\":\"Deskripsi\"") : len(jsonData)-1]
		jsonDesc = "{" + jsonDesc[0 : strings.Index(jsonDesc, "}")] + "}"

		var desc description
		err = json.Unmarshal([]byte(jsonDesc), &desc)
		if err != nil {
			log.Fatalf("Error Procces :  : %q", err)
		}
		
		writer.Write([]string{
			productName.ProductName,
			desc.SubTitle,
			img.UrlThumbnail,
			strconv.Itoa(int(productName.Price)),
			rating,
			shopName,
		})
	})

	for i, j := range products {
		fmt.Println("Fetch data to : " + strconv.Itoa(i + 1))
		col.Visit(j)
	}
}
