package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	appFile, err := os.Create("./sql/03_app.csv")
	if err != nil {
		panic(err)
	}
	defer appFile.Close()

	apiFile, err := os.Create("./sql/04_api.csv")
	if err != nil {
		panic(err)
	}
	defer apiFile.Close()

	evFile, err := os.Create("./sql/05_ev.csv")
	if err != nil {
		panic(err)
	}
	defer evFile.Close()

	docFile, err := os.Create("./sql/06_doc.csv")
	if err != nil {
		panic(err)
	}
	defer docFile.Close()

	appBuff := bufio.NewWriter(appFile)
	apiBuff := bufio.NewWriter(apiFile)
	evBuff := bufio.NewWriter(evFile)
	docBuff := bufio.NewWriter(docFile)

	apps := 1000
	apis := 10
	events := 10
	docs := 10

	for i := 0; i < apps; i++ {
		panicOnWritingErrors(fmt.Fprintf(appBuff, "%d,app-%d,tenant-%d\n", i, i, i))
		for k := 0; k < apis; k++ {
			panicOnWritingErrors(fmt.Fprintf(apiBuff, "%d,api-%d,%d\n", k+i*apis, k+i*apis, i))
		}

		for k := 0; k < events; k++ {
			panicOnWritingErrors(fmt.Fprintf(evBuff, "%d,ev-%d,%d\n", k+i*apis, k+i*apis, i))
		}
		for k := 0; k < docs; k++ {
			panicOnWritingErrors(fmt.Fprintf(docBuff, "%d,doc-%d,%d\n", k+i*apis, k+i*apis, i))

		}
	}

	if err := appBuff.Flush(); err != nil {
		panic(err)
	}
	if err := apiBuff.Flush(); err != nil {
		panic(err)
	}
	if err := evBuff.Flush(); err != nil {
		panic(err)
	}
	if err := docBuff.Flush(); err != nil {
		panic(err)
	}

}

func panicOnWritingErrors(_ int, err error) {
	if err != nil {
		panic(err)
	}
}
