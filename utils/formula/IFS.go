package formula

import (
	"errors"
)

func IFSCheck(func_list []string) error {
	if func_list[0] != "IFS" {
		return errors.New("NOT IFS , FORMULA ANA ERROR")
	}
	fn_list := func_list[2 : len(func_list)-1]
	if len(fn_list) == 0 {
		return errors.New("IFS函数参数错误")
	}
	tups, _ := SplitArraysByArray(fn_list, []string{"&", "|"})
	for _, each := range tups {
		temps, _ := SplitArraysByArray(each, []string{"<", ">", "=", "≤", "≥"})
		if len(temps) != 2 {
			return errors.New("IFS函数参数不合法")
		}
		for _, each_t := range temps {
			err := CheckArithmetic(each_t, "IFS")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func IFSCount(func_list []string) (error, float64) {
	if func_list[0] != "IFS" {
		return errors.New("NOT IFS , FORMULA ANA ERROR"), 0
	}
	fn_list := func_list[2 : len(func_list)-1]
	cmps := []string{}
	for _, each := range fn_list {
		if each == "&" || each == "|" {
			cmps = append(cmps, each)
		}
	}
	tups, _ := SplitArraysByArray(fn_list, []string{"&", "|"})
	res := []bool{}
	for _, each := range tups {
		err, cre := iFSCount(each)
		if err != nil {
			return err, 0
		}
		res = append(res, cre)
	}
	if len(res) == 1 {
		return nil, boolToFloat(res[0])
	}
	recm := false
	for index, each := range res {
		if index == 0 {
			continue
		}
		cmp := cmps[index-1]
		if index == 1 {
			recm = iFCountCm(res[0], each, cmp)
		} else {
			recm = iFCountCm(recm, each, cmp)
		}
	}
	return nil, boolToFloat(recm)
}

func boolToFloat(is bool) float64 {
	if is == true {
		return 1
	}
	return 0
}

func iFSCount(e []string) (error, bool) {
	temps, op := SplitArraysByArray(e, []string{"<", ">", "=", "≤", "≥"})
	a, err := countArithmetic(temps[0], 1, 2)

	if err != nil {
		return err, false
	}
	b, err := countArithmetic(temps[1], 1, 2)
	if err != nil {
		return err, false
	}
	switch op {
	case ">":
		return nil, a > b
	case "<":
		return nil, a < b
	case "=":
		return nil, a == b
	case "≤":
		return nil, a <= b
	case "≥":
		return nil, a >= b
	}
	return nil, false
}

func iFCountCm(a, b bool, op string) bool {
	switch op {
	case "&":
		return a && b
	case "||":
		return a || b
	}
	return false
}
