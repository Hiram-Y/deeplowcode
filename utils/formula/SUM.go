package formula

import (
	"errors"
)

func SUMCheck(func_list []string) error {
	if func_list[0] != "SUM" {
		return errors.New("NOT SUM , FORMULA ANA ERROR")
	}
	fn_list := func_list[2 : len(func_list)-1]
	if len(fn_list) == 0 {
		return errors.New(" SUM函数参数错误")
	}
	n_list, _ := SplitArraysByArray(fn_list, []string{";"})
	for _, each := range n_list {
		err := CheckArithmetic(each, "SUM")
		if err != nil {
			return err
		}
	}
	return nil
}

func SUMCount(func_list []string) (error, float64) {
	if func_list[0] != "SUM" {
		return errors.New("NOT SUM , FORMULA ANA ERROR"), 0
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
	return nil, sum(av)
}

func sum(xs []float64) (sum float64) {
	switch len(xs) {
	case 0:
		sum = 0
	default:
		sum = 0
		for _, v := range xs {
			sum += v
		}
	}
	return
}
