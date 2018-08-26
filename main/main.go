package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	modeSelected := false
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please select mode. (1 for server, 2 for client)")
	for {
		option, err := reader.ReadString('\n')
		option = strings.Replace(option, "\r\n", "", -1)
		if err != nil {
			fmt.Println(err)
		}
		switch option {
		case "1":
			NewServer()
			modeSelected = true
		case "2":
			NewClient()
			modeSelected = true
		default:
			fmt.Println("Wrong option, please try again.")
		}
		if modeSelected {
			break
		}
	}

}
