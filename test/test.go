package main

import (
	"fmt"
	"net/url"

	"github.com/go-apibox/apisign"
)

func main() {
	params := make(url.Values)
	params.Set("api_action", "Status.Overview")
	params.Set("api_agent_app", "sysinfo")
	params.Set("api_format", "json")
	params.Set("api_lang", "zh_cn")
	params.Set("api_timestamp", "1515502060")
	params.Set("api_nonce", "8YyjYz9t6H3ZVraY")

	fmt.Println(apisign.EncodeValues(params))
	fmt.Println(apisign.MakeSignString(params, "95bAzsK4AuYbrEnFjfUGdku5CXz2yKJn"))
}
