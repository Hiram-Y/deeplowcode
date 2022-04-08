package formula

import (
	"DeepWorkload/utils"
	"errors"
	"fmt"
	"strconv"
)

func SplitArraysByArray(f_list []string, split []string) (tups [][]string, splitStr string) {
	temp := []string{}
	for _, each := range f_list {
		if !utils.IsExistStr(split, each) {
			temp = append(temp, each)
		} else {
			splitStr = each
			tups = append(tups, temp)
			temp = []string{}
		}
	}
	tups = append(tups, temp)
	return tups, splitStr
}

func CheckLetter(f, n, formula_name string) error {
	Letters := []string{}
	for i := 65; i < 91; i++ {
		Letters = append(Letters, string(i))
	}
	if utils.StrInArray(f, Letters) && utils.StrInArray(n, Letters) {
		fmt.Println(f, n)
		return errors.New(fmt.Sprintf("函数%s存在参数没有运算符连接", formula_name))
	}
	if isNum(f) && utils.StrInArray(n, Letters) {
		fmt.Println(f, n)
		return errors.New(fmt.Sprintf("函数%s存在参数没有运算符连接", formula_name))
	}
	if isNum(n) && utils.StrInArray(f, Letters) {
		fmt.Println(f, n)
		return errors.New(fmt.Sprintf("函数%s存在参数没有运算符连接", formula_name))
	}

	if isOperatorwithoutk(n) && isOperatorwithoutk(f) {
		fmt.Println(f, n)
		return errors.New(fmt.Sprintf("函数%s中运算符%s%s不可连接", formula_name, n, f))
	}
	return nil
}

func isNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isOperator(r string) bool {
	return r == "+" || r == "-" || r == "*" || r == "/" || r == ">" || r == "<" || r == "≤" || r == "≥" || r == "=" || r == ")" || r == "("
}

func isOperatorwithoutequ(r string) bool {
	return r == "+" || r == "-" || r == "*" || r == "/" || r == ">" || r == "<" || r == "≤" || r == "≥" || r == ")" || r == "("
}

func isOperatorwithoutk(r string) bool {
	return r == "+" || r == "-" || r == "*" || r == "/" || r == ">" || r == "<" || r == "≤" || r == "≥" || r == "="
}

func isLetter(s string) bool {
	Letters := []string{}
	for i := 65; i < 91; i++ {
		Letters = append(Letters, string(i))
	}
	if utils.StrInArray(s, Letters) {
		return true
	}
	return false
}
