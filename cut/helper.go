package main

import (
	"fmt"
	"strconv"
	"strings"
)

func GetSeparator(s string) rune {
	var sep string
	s = `"` + s + `"`
	fmt.Sscanf(s, "%q", &sep)
	return ([]rune(sep))[0]
}

func ConvertStringsToInts(fields string) ([]int, error) {
	columns := []int{}
	for _, col := range strings.Split(fields, " ") {
		index, err := strconv.Atoi(string(col))
		if err != nil {
			return nil, err
		}
		columns = append(columns, index)
	}
	return columns, nil
}
