package main

import (
	"encoding/json"
	"fmt"
	"github.com/aiomonitors/godiscord"
	"io/ioutil"
	"net/http"
	"time"
)

type Product struct {
	Name          string `json:"name"`
	IsSaleProduct bool   `json:"isSaleProduct"`
	Image         string
	Images        [] Images `json:"images"`
	SellableUnits [] SellableUnits `json:"sellableUnits"`
}

type Images struct {
	Variations[] Variant `json:"variations"`
}

type SellableUnits struct {
	Price Price `json:"price"`
	IsRecaptchaOn bool  `json:"isRecaptchaOn"`
	Attributes[] Attribute `json:"attributes"`
	Status string `json:"stockLevelStatus"`
}

type Price struct {
	OriginalPrice float64 `json:"originalPrice"`
	CurrentPrice float64 `json:"value"`
}

type Attribute struct {
	Type string `json:"type"`
	Value string `json:"value"`
}

type Variant struct {
	Formant string `json:"format"`
	Url string `json:"url"`

}

func runMonitor(sku string, discordWebhook string) {
	client := &http.Client{}

	request, requestError := http.NewRequest("GET", 	fmt.Sprintf("https://www.footlocker.com/api/products/pdp/%s",sku), nil)
	if requestError != nil {
		fmt.Println(requestError)
		return
	}
	request.Close = true

	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36",
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"cache-control": "no-store,no-cache,must-revalidate,proxy-revalidate,max-age=0",
		"pragma": "no-cache",
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response, responseError := client.Do(request)
	if responseError != nil {
		fmt.Println(responseError)
		return
	}

	var product Product

	body, bodyErr := ioutil.ReadAll(response.Body)

	if bodyErr != nil {
		fmt.Println(bodyErr)
		return
	}
	response.Body.Close()

	err := json.Unmarshal(body, &product)

	if err != nil {
		fmt.Println(err)
		return
	}

	if product.IsSaleProduct {
		for _, images := range product.Images {
			for _, variant := range images.Variations {
				if variant.Formant == "large" {
					product.Image = variant.Url
				}
			}
		}

		sendWebhook(product, discordWebhook)
	}
}

func sendWebhook(product Product, discordWebhook string) {
	emb := godiscord.NewEmbed(product.Name, "", "")

	emb.SetThumbnail(product.Image)
	emb.SetColor("#ef5653")
	emb.SetFooter("FootLocker Monitor by Brennan", "")

	for _, size := range product.SellableUnits {
		if size.Status == "outOfStock" {
			continue
		}

		for _, attribute := range size.Attributes {
			if attribute.Type == "size" {
				emb.AddField(fmt.Sprintf("Size[%s]", attribute.Value), fmt.Sprintf("$%.2f", size.Price.CurrentPrice), false)
			}
		}
	}

	emb.Username = "FootLocker Monitor"

	emb.SendToWebhook(discordWebhook)
}

func main() {
	discordWebhook := "{DISCORD_WEBHOOK}"
	sku := "X6899010"

	for true {
		time.Sleep(3500 * time.Millisecond)

		runMonitor(sku, discordWebhook)
	}

}
