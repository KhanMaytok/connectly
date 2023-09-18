## Connectly example SDK

To install and test the SDK locally:

DOnload the repository and create a directory next to the downloaded repository, named testing:

    mkdir testing
    cd testing

Create a package:

    go mod init testing

Create a file main.go and add the following code:

```golang
package main

import (
	"encoding/json"
	"fmt"

	"github.com/khanmaytok/connectly"
)

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func main() {
	request := &connectly.BatchSendCampaignRequest{
		TemplateName: "template1",
		Language:     "en",
		BusinessID:   "f1980bf7-c7d6-40ec-b665-dbe13620bffa",
		APIKey:       "hehe",
		CsvFile:      "https://cdn.connectly.ai/custom/sample_connectly_campaign.csv",
	}

  // Call the SDK function
  response := connectly.BatchSendCampaign(request)
  fmt.Println(PrettyPrint(response))
}
```

At the generated go.mod file, add this code:

    require github.com/khanmaytok/connectly v0.0.0
    replace github.com/khanmaytok/connectly => ../connectly-go

Run the script:

    go run main.go

You will see the api mockup response id or error messages if present

