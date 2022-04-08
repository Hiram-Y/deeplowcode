package formula

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func countArithmetic(func_list []string, outcome_state, decimal_places int) (float64, error) {
F:
	index := -1
	for f_index, each := range func_list {
		if each == "(" {
			index = f_index
		}
	}
	if index != -1 {
		temp := []string{}
		start_index := index + 1
		end_index := -1
		for i := index + 1; i < len(func_list); i++ {
			if func_list[i] != ")" {
				temp = append(temp, func_list[i])
				continue
			}
			end_index = i
			break
		}
		tmpResult, err := count(strings.Join(temp, ""))
		if err != nil {
			return 0, err
		}
		tempResutStr := fmt.Sprintf("%2f", tmpResult)
		func_list = append(append(func_list[:start_index-1], tempResutStr), func_list[end_index+1:]...)
		goto F
	}
	tmpResult, err := count(strings.Join(func_list, ""))

	if err == nil {
		switch outcome_state {
		case 1:
			tmpResult = round(tmpResult, decimal_places)
		case 2:
			if float64(int(tmpResult)) != tmpResult {
				tmpResult = float64(int(tmpResult + 1))
			}
		case 3:
			if float64(int(tmpResult)) != tmpResult {
				tmpResult = float64(int(tmpResult))
			}
		}
	}
	return tmpResult, err
}

func CheckArithmetic(func_list []string, formula_name string) error {

F:
	index := -1
	for f_index, each := range func_list {
		if each == "(" {
			index = f_index
		}
	}
	if index != -1 {
		temp := []string{}
		start_index := index + 1
		end_index := -1
		for i := index + 1; i < len(func_list); i++ {
			if func_list[i] != ")" {
				temp = append(temp, func_list[i])
				continue
			}
			end_index = i
			break
		}
		if end_index == -1 {
			return errors.New(fmt.Sprintf("函数%s中有括号不完整", formula_name))
		}
		err := checkSimpleCount(func_list, formula_name)
		if err != nil {
			//fmt.Println(err.Error())
			return err
		}
		func_list = append(append(func_list[:start_index-1], "1"), func_list[end_index+1:]...)
		goto F
	}
	err := checkSimpleCount(func_list, formula_name)
	if err != nil {
		//fmt.Println(err.Error())
		return err
	}
	return nil
}

func round(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}

func checkSimpleCount(func_list []string, formula_name string) error {
	for index, each_t := range func_list {
		if strings.Contains(each_t, "string_check") {
			return errors.New(fmt.Sprintf("%s函数不可参与计算", strings.Split(each_t, "_")[0]))
		}
		if !isOperatorwithoutequ(each_t) && !isNum(each_t) && !isLetter(each_t) {
			return errors.New(fmt.Sprintf("函数%s中存在不合法的运算符%s", formula_name, each_t))
		}
		if index == 0 {
			continue
		}
		err := CheckLetter(each_t, func_list[index-1], formula_name)
		if err != nil {
			return err
		}
	}
	last := func_list[len(func_list)-1]
	if isOperatorwithoutk(last) {
		return errors.New(fmt.Sprintf("函数%s中运算符%s未连接参数", formula_name, last))
	}
	return nil
}
