package content

import "encoding/json"

var mockContents = map[string]Content{
	"tr": {
		SMS:          "Merhaba {{first_name}}, sepetin seni bekliyor! Bugün tamamla, %30 indirimi kaçırma.",
		EmailSubject: "{{first_name}}, sepetinde %30 indirim seni bekliyor",
		EmailBody: "Merhaba {{first_name}},\n\n" +
			"Sepetine eklediğin ürünler hâlâ seni bekliyor.\n\n" +
			"Siparişini bu hafta tamamlarsan %30 indirimden yararlanabilirsin.\n\n" +
			"Görüşmek üzere!",
	},
	"en": {
		SMS:          "Hi {{first_name}}, your cart is waiting! Finish today and get 30% off.",
		EmailSubject: "{{first_name}}, 30% off is waiting in your cart",
		EmailBody: "Hi {{first_name}},\n\n" +
			"The items you added to your cart are still waiting for you.\n\n" +
			"Complete your order this week and enjoy 30% off.\n\n" +
			"See you soon!",
	},
}

// MockContent returns the default tool input JSON for "tr" or "en". Other
// codes get "en".
func MockContent(lang string) []byte {
	c, ok := mockContents[lang]
	if !ok {
		c = mockContents["en"]
	}
	raw, _ := json.Marshal(c)
	return raw
}
