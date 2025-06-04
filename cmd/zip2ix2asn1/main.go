package main

import (
	"fmt"
	"log"
	"os"

	z2i "github.com/takanoriyanagitani/go-zip2ix"
)

func envValByKey(key string) (string, error) {
	val, found := os.LookupEnv(key)
	switch found {
	case true:
		return val, nil
	default:
		return "", fmt.Errorf("env val %s missing", key)
	}
}

func sub() error {
	zfilename, e := envValByKey("ENV_INPUT_ZIP_FILENAME")
	if nil != e {
		return e
	}

	e = z2i.ZipfilenameToStdout(zfilename)
	if nil != e {
		return fmt.Errorf("unable to process the zip file %s: %v", zfilename, e)
	}

	return nil
}

func main() {
	e := sub()
	if nil != e {
		log.Printf("%v\n", e)
	}
}
