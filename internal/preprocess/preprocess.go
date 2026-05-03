package preprocess

import (
	"encoding/json"
	"os"
	"path/filepath"
)

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
