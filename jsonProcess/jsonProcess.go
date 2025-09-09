package jsonProcess

import (
	"encoding/json"
	"mcmod-crawler/crawler"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func SaveJSON(filename string, pack interface{}) error {
	bytes, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, bytes, 0666)
}

func parseViews(s string) int {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "，", "")
	re := regexp.MustCompile(`([\d.]+)(万)?`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0
	}
	num, _ := strconv.ParseFloat(matches[1], 32)
	if len(matches) >= 3 && matches[2] == "万" {
		num *= 10000
	}
	return int(num)
}

func SortJson(packs []crawler.ModPack, field string, descending bool) {
	sort.Slice(packs, func(i, j int) bool {
		var less bool
		switch field {
		case "title":
			less = packs[i].Title < packs[j].Title
		case "views":
			less = parseViews(packs[i].Views) < parseViews(packs[j].Views)
		case "points":
			less = packs[i].Points < packs[j].Points
		case "scores":
			less = packs[i].Scores < packs[j].Scores
		default:
			less = true
		}
		if descending {
			return !less
		}
		return less
	})
}

func LoadJson(filename string) ([]crawler.ModPack, error) {
	var packs []crawler.ModPack
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &packs)
	if err != nil {
		return nil, err
	}
	return packs, nil
}
