package pinyin

import (
	"fmt"
	"testing"
)

func TestConvert(t *testing.T) {
	str, err := New("hi111").Split("").Convert()
	fmt.Println(str)
	fmt.Println("---")

	if err != nil {
		t.Error(err)
	} else {
		t.Log(str)
	}
}
