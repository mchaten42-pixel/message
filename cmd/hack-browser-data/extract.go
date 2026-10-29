package main

import (
	"fmt"
	"path/filepath"

	"github.com/mchaten42-pixel/message/browser"
	"github.com/mchaten42-pixel/message/log"
	"github.com/mchaten42-pixel/message/output"
	"github.com/mchaten42-pixel/message/types"
	"github.com/mchaten42-pixel/message/utils/fileutil"
)

func extractAndWrite(browsers []browser.Browser, categories []types.Category, outputDir, outputFormat string, compress bool) error {
	w, err := output.NewWriter(outputDir, outputFormat)
	if err != nil {
		return err
	}
	for _, b := range browsers {
		log.Infof("Extracting %s...", b.BrowserName())
		results, extractErr := b.Extract(categories)
		if extractErr != nil {
			log.Errorf("extract %s: %v", b.BrowserName(), extractErr)
		}
		for _, r := range results {
			w.Add(b.BrowserName(), r.Name, r.Data)
		}
	}
	if err := w.Write(); err != nil {
		return err
	}
	if compress {
		if err := fileutil.CompressDir(outputDir); err != nil {
			return fmt.Errorf("compress: %w", err)
		}
		log.Infof("Compressed: %s/%s.zip", outputDir, filepath.Base(outputDir))
	}
	return nil
}
