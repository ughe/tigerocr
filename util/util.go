package util

import (
	"bufio"
	"io/ioutil"
	"os"
)

func Abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}

func Read(fileName string) ([]byte, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func Write(buf []byte, fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	n := 0
	for n < len(buf) {
		m, err := f.Write(buf[n:])
		if err != nil {
			return err
		}
		n += m
	}
	return nil
}
