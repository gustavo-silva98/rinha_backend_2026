package preprocess

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	Magic      = "RVEC"
	Dims       = 14
	Stride     = 16
	Version    = uint32(1)
	HeaderSize = 16
)

type StoredRefs struct {
	Raw   []byte
	Count uint32
}

type NormalizationConstant struct {
	MaxAmount            int `json:"max_amount"`
	MaxInstallments      int `json:"max_installments"`
	AmountVsAvgRatio     int `json:"amount_vs_avg_ratio"`
	MaxMinutes           int `json:"max_minutes"`
	MaxKM                int `json:"max_km"`
	MaxTxCount24         int `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount int `json:"max_merchant_avg_amount"`
}

func LoadNormalization() (NormalizationConstant, error) {
	var norm NormalizationConstant
	workingPath, err := os.Getwd()
	if err != nil {
		return NormalizationConstant{}, err
	}
	path := filepath.Join(workingPath, "resources", "normalization.json")
	file, err := os.ReadFile(path)
	if err != nil {
		return NormalizationConstant{}, err
	}
	err = json.Unmarshal(file, &norm)
	if err != nil {
		return NormalizationConstant{}, err
	}
	return norm, nil

}

func LoadMCC() (map[string]float64, error) {
	var mcc map[string]float64
	workingPath, err := os.Getwd()
	if err != nil {
		return map[string]float64{}, err
	}
	path := filepath.Join(workingPath, "resources", "mcc_risk.json")
	file, err := os.ReadFile(path)
	if err != nil {
		return map[string]float64{}, err
	}
	err = json.Unmarshal(file, &mcc)
	if err != nil {
		return map[string]float64{}, err
	}
	return mcc, nil

}
func Quantize(v float64) int8 {
	if v == -1.0 {
		return -128
	}
	q := math.Round(v * 127.0)
	if q > 127 {
		q = 127
	}
	if q < 0 {
		q = 0
	}
	return int8(q)
}

func LoadRefs(path string) (*StoredRefs, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	data, err := syscall.Mmap(
		int(f.Fd()), 0, int(fi.Size()),
		syscall.PROT_READ, syscall.MAP_SHARED,
	)

	if err != nil {
		log.Fatal(err)
	}
	if string(data[:4]) != "RVEC" {
		return nil, fmt.Errorf("Magic inválido")
	}
	count := binary.LittleEndian.Uint32(data[8:12])

	syscall.Madvise(data, syscall.MADV_SEQUENTIAL)
	syscall.Madvise(data, syscall.MADV_WILLNEED)

	return &StoredRefs{Raw: data, Count: count}, nil

}

func (s *StoredRefs) Vec(i uint32) []int8 {
	off := HeaderSize + int(i)*Stride
	return unsafe.Slice((*int8)(unsafe.Pointer(&s.Raw[off])), Dims)
}

func (s *StoredRefs) Label(i uint32) uint8 {
	return s.Raw[HeaderSize+int(i)*Stride+Dims]
}

func (s *StoredRefs) ConvertLabel(i uint32) bool {
	label := s.Label(i)
	if label == 1 {
		return true
	} else {
		return false
	}
}

func Bruteforce(a, b []int8) int32 {
	var acc int32
	for i := 0; i < Dims; i++ {
		d := int32(a[i]) - int32(b[i])
		acc += d * d
	}
	return acc
}
