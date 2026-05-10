package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"rinha2026/internal/model"
	"rinha2026/internal/preprocess"
)

func main() {
	workingPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(workingPath, "resources", "references.json.gz")

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(bufio.NewReaderSize(f, 1<<20))
	if err != nil {
		log.Fatal(err)
	}
	defer gz.Close()

	os.Chdir("..")
	os.Chdir("..")
	out, err := os.Create("vectors.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	bw := bufio.NewWriterSize(out, 1<<20)

	bw.WriteString(preprocess.Magic)
	binary.Write(bw, binary.LittleEndian, preprocess.Version)
	binary.Write(bw, binary.LittleEndian, uint32(0))
	binary.Write(bw, binary.LittleEndian, uint32(preprocess.Dims))

	dec := json.NewDecoder(gz)
	dec.Token()
	buf := make([]byte, preprocess.Stride)
	var idx int32
	var r model.RawVector

	for dec.More() {
		if err := dec.Decode(&r); err != nil {
			log.Fatalf("erro ao decodificar vetor idx: %v - %v", idx, err)
		}
		if len(r.Vector) != preprocess.Dims {
			log.Fatalf("erro no número de dimensões idx: %v - %v", idx, len(r.Vector))
		}

		for i, v := range r.Vector {
			buf[i] = byte(preprocess.Quantize(v))
		}

		if r.Label == "fraud" {
			buf[preprocess.Dims] = 1
		} else {
			buf[preprocess.Dims] = 0
		}
		buf[preprocess.Dims+1] = 0 //Padding de cache line

		bw.Write(buf)
		idx++
	}
	if err := bw.Flush(); err != nil {
		log.Fatal(err)
	}
	if _, err := out.Seek(8, io.SeekStart); err != nil {
		log.Fatal(err)
	}
	binary.Write(out, binary.LittleEndian, uint32(idx))
}
