package main

import (
	"pleco-api/internal/appsetup"
)

func main() {
	if err := appsetup.RunAPI(appsetup.RegisterDocsFromDisk); err != nil {
		panic(err)
	}
}
