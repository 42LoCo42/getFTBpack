package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
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

	log.Print("Downloading manifest...")
	resp, err := http.Get(fmt.Sprintf(manifestURL, packID, releaseID))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}
	if err := os.Chdir(outputDir); err != nil {
		panic(err)
	}

	ch := make(chan File)
	wg := sync.WaitGroup{}

	for i := 0; i < 16; i++ {
		go func(i int){
			for {
				func() {
					file := <- ch
					defer wg.Done()

					outPath := path.Join(file.Path, file.Name)
					fmt.Printf("\x1B[%d;1H\x1B[K%s", i + 3, outPath)

					if err := os.MkdirAll(file.Path, 0755); err != nil {
						log.Print(err)
						return
					}

					out, err := os.Create(outPath)
					if err != nil {
						log.Print(err)
						return
					}
					defer out.Close()

					resp, err := http.Get(file.URL)
					if err != nil {
						log.Print(err)
						return
					}
					defer resp.Body.Close()

					if _, err := io.Copy(out, resp.Body); err != nil {
						log.Print(err)
						return
					}
				}()
			}
		}(i)
	}

	fmt.Print("\x1B[H\x1B[J")
	log.Printf("Dispatching %d jobs...", len(manifest.Files))
	for _, file := range manifest.Files {
		wg.Add(1)
		ch <- file
	}

	log.Print("\x1B[20;1HFinalizing...")
	wg.Wait()

	for _, target := range manifest.Targets {
		log.Printf("%s: %s", target.Name, target.Version)
	}
}
