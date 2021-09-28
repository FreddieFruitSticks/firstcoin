package utils

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
