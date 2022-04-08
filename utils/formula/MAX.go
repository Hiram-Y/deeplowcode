package formula

import (
	"errors"
	"math"
)

func MAXCheck(func_list []string) error {
	if func_list[0] != "MAX" {
		return errors.New("NOT MAX , FORMULA ANA ERROR")
	}
	fn_list := func_list[2 : len(func_list)-1]
	if len(fn_list) == 0 {
		return errors.New(" MAX函数参数错误")
	}
	n_list, _ := SplitArraysByArray(fn_list, []string{";"})
	for _, each := range n_list {
		err := CheckArithmetic(each, "MAX")
		if err != nil {
			return err
		}
	}
	return nil
}

func MAXCount(func_list []string) (error, float64) {
	if func_list[0] != "MAX" {
		return errors.New("NOT MAX , FORMULA ANA ERROR"), 0
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
	return nil, max(av)

}

func max(xs []float64) (max float64) {
	switch len(xs) {
	case 0:
		max = 0
	default:
		max = xs[0]
		for _, v := range xs {
			max = math.Max(max, v)
		}
	}
	return
}
