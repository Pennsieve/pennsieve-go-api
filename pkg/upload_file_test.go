package pkg

import (
	"fmt"
	"sort"
	"testing"
)

func TestHandler(t *testing.T) {

	testPaths := []string{
		"abc1",
		"abc2",
		"abc1/abc2",
		"def1/abc2",
		"def1",
	}

	sort.Strings(testPaths)

	for _, j := range testPaths {
		fmt.Println(j)
	}

}
