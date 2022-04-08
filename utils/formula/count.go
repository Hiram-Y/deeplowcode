package formula

import (
	"errors"
	"fmt"
)

func Count(formula_methods []string, outcome_state, decimal_places int) (float64, error) {

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
		err, result := countFormulaFunction(temp)
		if err != nil {
			return 0, err
		}
		if Formaulas[formulaName] == "Float" {
			re_str := fmt.Sprintf("%2f", result.(float64))
			formula_methods = append(append(formula_methods[:start_index-1], re_str), formula_methods[end_index+1:]...)
		} else {
			formula_methods = append(append(formula_methods[:start_index-1], result.(string)), formula_methods[end_index+1:]...)
		}
		goto F
	}

	re, err := countArithmetic(formula_methods, outcome_state, decimal_places)
	if err != nil {
		return 0, err
	}
	return re, nil
}

func countFormulaFunction(func_list []string) (error, interface{}) {

	switch func_list[0] {
	case "IFS":
		return IFSCount(func_list)
	case "ABS":
		return ABSCount(func_list)
	case "AVERAGE":
		return AVERAGECount(func_list)
	case "MIN":
		return MINCount(func_list)
	case "MAX":
		return MAXCount(func_list)
	case "SUM":
		return SUMCount(func_list)
	}
	return errors.New(fmt.Sprintf("没有%s函数", func_list)), nil
}
