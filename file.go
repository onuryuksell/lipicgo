package lipicgo

import (
	_ "image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
)

func GetFromPath(filePath string) (*os.File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	return f, err
}
