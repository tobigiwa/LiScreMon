//go:build ignore

package main

import (
	"agent"
	"context"
	"log"
	"os"
	"path/filepath"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err) // exit
	}

	file, err := os.Create(filepath.Join(cwd, "frontend", "index.html"))
	if err != nil {
		log.Fatalln("your index.html file might not be up-to-date, ensure compilation is in the project root dir:", err) // exit
	}
	defer file.Close()

	if err = agent.IndexPage().Render(context.TODO(), file); err != nil {
		log.Fatalln("your index.html file might not be up-to-date, could not render into index.html: ", err) // exit
	}

	log.Println("index.html file generated and placed in the frontend directory")
}
