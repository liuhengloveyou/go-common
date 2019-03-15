package common

import (
	"io"
	"io/ioutil"
	"os"
)

func CopyFile(src, dst string, append bool) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	if append {
		file, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			return err
		}

		n, err := file.Write(input)
		if err == nil && n < len(input) {
			err = io.ErrShortWrite
		}

		if err != nil {
			return err
		}
	} else {
		if err = ioutil.WriteFile(dst, input, 755); err != nil {
			return err
		}
	}

	return nil
}
