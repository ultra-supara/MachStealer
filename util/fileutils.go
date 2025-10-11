package util

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/exp/constraints"
)

func FileCopy(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("FileCopy failed: %w", err)
	}

	err = os.WriteFile(dst, b, 0644)
	if err != nil {
		return fmt.Errorf("FileCopy failed: %w", err)
	}
	return nil
}

func TimeStamp(stamp int64) time.Time {
	s := time.Unix(stamp, 0)
	if s.Local().Year() > 9999 {
		return time.Date(9999, 12, 13, 23, 59, 59, 0, time.Local)
	}
	return s
}

func TimeEpoch(epoch int64) time.Time {
	maxTime := int64(99633311740000000)
	if epoch > maxTime {
		return time.Date(2049, 1, 1, 1, 1, 1, 1, time.Local)
	}
	t := time.Date(1601, 1, 1, 0, 0, 0, 0, time.Local)
	d := time.Duration(epoch)
	for i := 0; i < 1000; i++ {
		t = t.Add(d)
	}
	return t
}

func IntToBool[T constraints.Signed](a T) bool {
	switch a {
	case 0, -1:
		return false
	}
	return true
}
