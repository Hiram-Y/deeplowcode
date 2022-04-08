package formula

import (
	"fmt"
	"testing"
)

func TestCheckFormula(t *testing.T) {
	temp := []string{"+", "IFS", "(", "A", ">", "B", ")"}
	err := CheckFormula(temp)
	fmt.Println(err)
}

func TestABSCount(t *testing.T) {
	f, err := Count([]string{"ABS", "(", "1", "-", "3", ")"}, 1, 1)
	fmt.Println(f, err)
}

func TestCount(t *testing.T) {
	f, err := Count([]string{"1"}, 1, 1)
	fmt.Println(f, err)
}
