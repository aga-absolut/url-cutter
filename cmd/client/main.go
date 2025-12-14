package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	endPoint := "http://localhost:8080/google.com/"

	data := url.Values{}

	fmt.Println("Введите длинный URL")

	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	long = strings.TrimSuffix(long, "\n")

	data.Set("url", long)

	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPost, endPoint, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println("Статус код", response.StatusCode)

	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))

}
