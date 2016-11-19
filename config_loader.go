package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type ConfigLoader struct {
	// Table of extension to unmarshal func
	Unmarshals map[string]func([]byte, interface{}) error
}

func (c *ConfigLoader) Load(basePath string, obj interface{}) error {
	var fun func([]byte, interface{}) error
	var path string
	ok := false
	for ext, f := range c.Unmarshals {
		path = fmt.Sprintf("%s.%s", basePath, ext)
		if Exists(path) {
			fun = f
			ok = true
			break
		}
	}
	if !ok {
		return errors.Errorf("Cofnig file for %s does not exist", basePath)
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "ReadFile is failed")
	}

	return fun(b, obj)
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
