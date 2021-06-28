package iterators

import (
	"bufio"
	"context"

	"github.com/Shopify/go-storage"
	"github.com/Shopify/goose/maintenance"
)

type fileIterator struct {
	filePath string
	storage  storage.FS

	index   int64
	file    *storage.File
	scanner *bufio.Scanner
}

func NewFileIterator(filePath string, storage storage.FS) maintenance.Iterator {
	return &fileIterator{filePath: filePath, storage: storage}
}

func (fi *fileIterator) Next(ctx context.Context, cursor int64) ([]interface{}, int64, error) {
	if fi.scanner == nil || fi.index > cursor {
		err := fi.resetScanner(ctx)
		if err != nil {
			return nil, 0, err
		}
	}

	for ; fi.index <= cursor; fi.index++ { // Advance file index up to cursor's location
		if !fi.scanner.Scan() {
			defer fi.file.Close() // Close the file if we're done reading

			if fi.scanner.Err() != nil {
				return nil, 0, fi.scanner.Err()
			}

			return nil, 0, nil
		}
	}

	return []interface{}{fi.scanner.Text()}, fi.index, nil
}

func (fi *fileIterator) resetScanner(ctx context.Context) error {
	if fi.file != nil { // Close the old file descriptor
		err := fi.file.Close()
		if err != nil {
			return err
		}
	}

	file, err := fi.storage.Open(ctx, fi.filePath, nil)
	if err != nil {
		return err
	}

	fi.index = 0
	fi.file = file
	fi.scanner = bufio.NewScanner(fi.file)

	return nil
}
