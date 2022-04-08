package formula

import (
	"errors"
	"math"
)

func MINCheck(func_list []string) error {
	if func_list[0] != "MIN" {
		return errors.New("NOT MIN , FORMULA ANA ERROR")
	}
	fn_list := func_list[2 : len(func_list)-1]
	if len(fn_list) == 0 {
		return errors.New(" MIN函数参数错误")
	}
	n_list, _ := SplitArraysByArray(fn_list, []string{";"})
	for _, each := range n_list {
		err := CheckArithmetic(each, "MIN")
		if err != nil {
			return err
		}
	}
	return nil
}

func MINCount(func_list []string) (error, float64) {
	if func_list[0] != "MIN" {
		return errors.New("NOT MIN , FORMULA ANA ERROR"), 0
	}
	fn_list := func_list[2 : len(func_list)-1]
	n_list, _ := SplitArraysByArray(fn_list, []string{";"})
	av := []float64{}
	for _, each := range n_list {
		a, err := countArithmetic(each, 1, 2)
		if err != nil {
			return err, 0
		}
		av = append(av, a)
	}
	return nil, min(av)

}

func min(xs []float64) (min float64) {
	switch len(xs) {
	case 0:
		min = 0
	default:
		min = xs[0]
		for _, v := range xs {
			min = math.Min(min, v)
		}
	}
	return
}
