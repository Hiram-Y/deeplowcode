package formula

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

func ABSCheck(func_list []string) error {
	if func_list[0] != "ABS" {
		return errors.New("NOT ABS , FORMULA ANA ERROR")
	}
	fn_list := func_list[2 : len(func_list)-1]

	if len(fn_list) == 0 {
		return errors.New("ABS函数参数错误")
	}

	if len(fn_list) == 1 {
		if !(isLetter(fn_list[0]) || isNum(fn_list[0])) {
			return errors.New("ABS函数参数错误")
		}
	} else {
		err := CheckArithmetic(fn_list, "ABS")
		if err != nil {
			return err
		}
	}
	return nil
}

func ABSCount(func_list []string) (error, float64) {
	if func_list[0] != "ABS" {
		return errors.New("NOT ABS , FORMULA ANA ERROR"), 0
	}
	fn_list := func_list[2 : len(func_list)-1]
	if len(fn_list) == 1 {
		if f, err := strconv.ParseFloat(fn_list[0], 64); err != nil {
			return errors.New(fmt.Sprintf("计算解析错误,无法将%s,转化成一个float64类型,请检查公式", fn_list[0])), 0
		} else {
			return nil, math.Abs(f)
		}
	} else {
		f, err := countArithmetic(fn_list, 1, 2)
		if err != nil {
			return err, 0
		}
		return nil, math.Abs(f)
	}
}
