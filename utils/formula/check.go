package formula

import (
	"errors"
	"fmt"
)

func CheckFormula(formula_methods []string) error {

F:
	index := -1
	formulaName := ""
	for f_index, each := range formula_methods {
		if isInMapKes(each, Formaulas) {
			index = f_index
			formulaName = each
		}
	}
	if index != -1 {
		temp := []string{}
		start_index := index + 1
		end_index := -1
		for i := index; i < len(formula_methods); i++ {
			if formula_methods[i] != ")" {
				temp = append(temp, formula_methods[i])
				continue
			}
			temp = append(temp, formula_methods[i])
			end_index = i
			break
		}

		if end_index == -1 {
			return errors.New(fmt.Sprintf("函数%s括号不完整", formulaName))
		}
		err := checkFormulaFunction(temp)
		if err != nil {
			return err
		}
		if Formaulas[formulaName] == "Float" {
			formula_methods = append(append(formula_methods[:start_index-1], "1"), formula_methods[end_index+1:]...)
		} else {
			formula_methods = append(append(formula_methods[:start_index-1], formulaName+"_string_check"), formula_methods[end_index+1:]...)
		}
		goto F
	}
	err := CheckArithmetic(formula_methods, "") //check四则有问题，括号不完整未校验
	if err != nil {
		return err
	}
	return nil
}

func checkFormulaFunction(func_list []string) error {

	switch func_list[0] {
	case "IFS":
		return IFSCheck(func_list)
	case "ABS":
		return ABSCheck(func_list)
	case "AVERAGE":
		return AVERAGECheck(func_list)
	case "MIN":
		return MINCheck(func_list)
	case "MAX":
		return MAXCheck(func_list)
	case "SUM":
		return SUMCheck(func_list)
	}
	return errors.New(fmt.Sprintf("没有%s函数", func_list))
}

func isInMapKes(f string, fs map[string]string) bool {
	for k, _ := range fs {
		if k == f {
			return true
		}
	}
	return false
}
