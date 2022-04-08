package formula

import "errors"

func AVERAGECheck(func_list []string) error {
	if func_list[0] != "AVERAGE" {
		return errors.New("NOT AVERAGE , FORMULA ANA ERROR")
	}
	fn_list := func_list[2 : len(func_list)-1]
	if len(fn_list) == 0 {
		return errors.New(" AVERAGE函数参数错误")
	}
	n_list, _ := SplitArraysByArray(fn_list, []string{";"})
	for _, each := range n_list {
		err := CheckArithmetic(each, "AVERAGE")
		if err != nil {
			return err
		}
	}
	return nil
}

func AVERAGECount(func_list []string) (error, float64) {
	if func_list[0] != "AVERAGE" {
		return errors.New("NOT AVERAGE , FORMULA ANA ERROR"), 0
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
	return nil, average(av)

}

func average(xs []float64) (avg float64) {
	sum := 0.00
	switch len(xs) {
	case 0:
		avg = 0
	default:
		for _, v := range xs {
			sum += v
		}
		avg = sum / float64(len(xs))
	}
	return
}
