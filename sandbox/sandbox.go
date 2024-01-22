package sandbox

import (
	"context"
	"crypto/tls"
//	"encoding/json"
	"fmt"
	"io"
//	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaSub-go/account"
	"github.com/LalatinaHub/LatinaSub-go/helper"
	"github.com/LalatinaHub/LatinaSub-go/ipapi"
	B "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
)

var (
	populateType []string = []string{"cdn", "sni"}
)

type SandBox struct {
	Link        string
	ConnectMode []string
	IpapiObj    ipapi.Ipapi
}

func worker(link, connectMode string) (string, ipapi.Ipapi) {
	var (
		acc        = account.New(link)
		options    option.Options
		listenPort uint
	)

	// Guard
	if acc.Outbound.Type == "" {
		return "", ipapi.Ipapi{}
	}

	if connectMode == "cdn" {
		options, listenPort = generateConfig(acc.PopulateCDN())
	} else {
		options, listenPort = generateConfig(acc.PopulateSNI())
	}

	box, err := B.New(B.Options{
		Context: context.Background(),
		Options: options,
	})
	if err != nil {
		panic(err)
	}
	defer box.Close()

	// Start sing-box client
	box.Start()

	proxyClient, _ := url.Parse(fmt.Sprintf("socks5://0.0.0.0:%d", listenPort))
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyClient),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	buf := new(strings.Builder)

	res, _ := httpClient.Get("http://ipapi.co/json")
	defer res.Body.Close()
	resData, _ := ioutil.ReadAll(res.Body)
	var result map[string]interface{}
	json.Unmarshal(resData, &result)
	ip := result["ip"].(string)
	resp, err := httpClient.Get("http://ipinfo.io/"+ip+"/json")

//	resp, err := httpClient.Get("http://ipapi.co/json")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	io.Copy(buf, resp.Body)
	if resp.StatusCode == 200 {
		return connectMode, ipapi.Parse(buf.String())
	}

	return "", ipapi.Ipapi{}
}

func Test(link string) *SandBox {
	var sb SandBox = SandBox{}

	// Constructor
	sb.Link = link

	for _, t := range populateType {
		switch strings.Split(link, "://")[0] {
		case "ssr", "ss":
			if t == "cdn" {
				continue
			}
		}

		mode, ipapi := func(link, t string) (string, ipapi.Ipapi) {
			defer helper.CatchError(false)
			return worker(link, t)
		}(link, t)

		if mode != "" {
			sb.ConnectMode = append(sb.ConnectMode, mode)
			sb.IpapiObj = ipapi
		}
	}

	return &sb
}
