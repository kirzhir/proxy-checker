package checker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Checker struct {
	Http     *http.Client
	endpoint string
}

func NewChecker(endpoint string, timeout uint) *Checker {
	return &Checker{
		&http.Client{Timeout: time.Duration(timeout * uint(time.Second))},
		endpoint,
	}
}

func (c *Checker) Check(proxy string) (bool, error) {
	// start := time.Now()
	c.Http.Transport = &http.Transport{Proxy: http.ProxyURL(&url.URL{Host: proxy})}
	response, err := c.Http.Get(c.endpoint)

	if err != nil {
		return false, err
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	// time_diff := time.Now().UnixNano() - start.UnixNano()

	// fmt.Println(time_diff)
	fmt.Println(string(body))

	return true, nil
}
