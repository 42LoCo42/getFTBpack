package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/go-faster/errors"
)

const manifestURL = "https://api.modpacks.ch/public/modpack/%s/%s"
const outputDir = "out"

type File struct {
	Path string `json:"path"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Sha1 string `json:"sha1"`
}

type Manifest struct {
	Files   []File `json:"files"`
	Targets []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"targets"`
}

func (m *Manifest) Run() {
	fmt.Println("runner")
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <pack ID> <release ID>", os.Args[0])
	}

	packID := os.Args[1]
	releaseID := os.Args[2]

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(errors.Wrap(err, "could not make outdir"))
	}
	if err := os.Chdir(outputDir); err != nil {
		log.Fatal(errors.Wrap(err, "could not enter outdir"))
	}

	log.Print("[1;34mDownloading manifest...[m")
	resp, err := http.Get(fmt.Sprintf(manifestURL, packID, releaseID))
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not GET manifest"))
	}
	defer resp.Body.Close()

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		log.Fatal(errors.Wrap(err, "could not decode manifest JSON"))
	}

	ch := make(chan File)
	wg := sync.WaitGroup{}
	failed := []string{}

	for i := 0; i < 16; i++ {
		go func(i int) {
			for {
				if err, outPath := func() (error, string) {
					file := <-ch
					defer wg.Done()

					outPath := path.Join(file.Path, file.Name)

					if err := os.MkdirAll(file.Path, 0755); err != nil {
						return errors.Wrap(err, "could not make outdir"), outPath
					}

					out, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, 0644)
					if err != nil {
						return errors.Wrap(err, "could not open outfile"), outPath
					}
					defer out.Close()

					resp, err := http.Get(file.URL)
					if err != nil {
						return errors.Wrap(err, "could not request outfile"), outPath
					}
					defer resp.Body.Close()

					if _, err := io.Copy(out, resp.Body); err != nil {
						return errors.Wrap(err, "could not write outfile"), outPath
					}

					if _, err := out.Seek(0, 0); err != nil {
						return errors.Wrap(err, "could not rewind outfile"), outPath
					}

					written, err := io.ReadAll(out)
					if err != nil {
						return errors.Wrap(err, "could not read outfile"), outPath
					}

					sum := sha1.Sum(written)
					enc := hex.EncodeToString(sum[:])
					if enc != file.Sha1 {
						return errors.New("sha1 mismatch!"), outPath
					}

					log.Printf("%v -> %v", file.URL, outPath)
					return nil, ""
				}(); err != nil {
					log.Printf("[1;31m%v[m", errors.Wrap(err, outPath))
					failed = append(failed, outPath)
				}
			}
		}(i)
	}

	log.Printf("[1;34mDispatching %d jobs...[m", len(manifest.Files))
	for _, file := range manifest.Files {
		wg.Add(1)
		ch <- file
	}

	log.Print("[1;34mWaiting for last jobs to complete...[m")
	wg.Wait()

	if len(failed) == 0 {
		log.Print("[1;32mDownload successful![m")
		for _, target := range manifest.Targets {
			log.Printf("%s: %s", target.Name, target.Version)
		}
	} else {
		log.Print("[1;31mDownload failed for these files:[m")
		for _, name := range failed {
			log.Print(name)
		}
		log.Print("[1;31m↑  Check above for errors! ↑  [m")
	}
}
