package main

import (
    "fmt"
    "os"
    //"encoding/json"
    //"io/ioutil"
    //"strconv"
)

func main() {
	fmt.Println("It works! =)")
	fmt.Println("JSON file name:", os.Args[1])

	jsonFileName := os.Args[1]
	jsonFile, err := os.Open(jsonFileName)

    	if err != nil {
		// Handle error.
      		fmt.Println(err)
    	}
    	fmt.Println("Successfully opened file.")
    	defer jsonFile.Close()
}
