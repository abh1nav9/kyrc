package identity

import "os"

func statMode(path string) (os.FileMode, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fi.Mode(), nil
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}
