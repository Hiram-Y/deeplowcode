package utils

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type stackNode struct {
	Data interface{}
	next *stackNode
}

type linkStack struct {
	top   *stackNode
	Count int
}

func (link *linkStack) init() {
	link.top = nil
	link.Count = 0
}

func (link *linkStack) push(data interface{}) {
	node := new(stackNode)
	node.Data = data
	node.next = link.top
	link.top = node
	link.Count++
}

func (link *linkStack) pop() interface{} {
	if link.top == nil {
		return nil
	}
	returnData := link.top.Data
	link.top = link.top.next
	link.Count--
	return returnData
}

//lookTop func Look up the top element in the stack, but not pop.
func (link *linkStack) lookTop() interface{} {
	if link.top == nil {
		return nil
	}
	return link.top.Data
}

func CountFunc1(expr string, outcome_state, decimal_places int) (float64, error) {
	var step int = 0
F:
	exprSlice := strings.Split(expr, "")
	index := strings.LastIndex(expr, "[")
	lenExpr := len(exprSlice)

	if lenExpr == step {

		return 0, errors.New("请检查公式是否存在中文字符")
	}
	if index != -1 {
		var strBuff bytes.Buffer
		for i := index + 1; i < lenExpr; i++ {
			if exprSlice[i] != "]" {
				strBuff.WriteString(exprSlice[i])
				continue
			}
			break
		}
		r := strBuff.String()
		cerr, tmp := checkFunction(r)
		if cerr != nil {
			return 0, cerr
		}
		if tmp == nil {
			return 0, errors.New("公式名称不存在")
		}
		strBuff.Reset()
		strBuff.WriteString("[")
		strBuff.WriteString(r)
		strBuff.WriteString("]")
		expr = strings.Replace(expr, strBuff.String(), fmt.Sprint(tmp), -1)
		goto F
	}

	return Calculate(expr, outcome_state, decimal_places)
}

func checkFunction(r string) (error, interface{}) {
	function_name := strings.Split(r, "(")[0]
	switch function_name {
	case "IFS":
		re, err := iFS(r)

		return err, re
	}
	return nil, nil
}

func iFS(data string) (int, error) {
	real_data := data[4 : len(data)-1]
	and := []string{}
	diff := []string{}
	var step int = 0
L1:
	exprSlice := strings.Split(real_data, "")
	index := strings.LastIndex(real_data, "(")
	lenExpr := len(exprSlice)
	if lenExpr == step {

		return 0, errors.New("计算死循环,请检查公式是否存在中文字符")
	}
	step = lenExpr
	s_count := strings.Count(real_data, "(")
	if index != -1 && s_count > 1 {
		var strBuff bytes.Buffer
		for i := index + 1; i < lenExpr; i++ {
			if exprSlice[i] != ")" {
				strBuff.WriteString(exprSlice[i])
				continue
			}
			break
		}
		r := strBuff.String()

		tmpResult, err := count(r)
		if err != nil {
			return 0, err
		}
		strBuff.Reset()
		strBuff.WriteString("(")
		strBuff.WriteString(r)
		strBuff.WriteString(")")

		tempResutStr := strconv.FormatFloat(tmpResult, 'f', 3, 64)
		if len(strBuff.String()) != len(tempResutStr) {
			real_data = strings.Replace(real_data, strBuff.String(), tempResutStr, -1)
		} else {
			real_data = strings.Replace(real_data, strBuff.String(), strconv.FormatFloat(tmpResult, 'f', 4, 64), -1)
		}
		goto L1
	}
	for _, each := range real_data {
		each_s := string(each)
		if each_s == "&" || each_s == "|" {
			and = append(and, each_s)
		}
	}
	diff = strings.FieldsFunc(real_data, func(r rune) bool {
		return r == '&' || r == '|'
	})
	diff_r := []bool{}
	for _, each := range diff {
		diff_r = append(diff_r, iFSCheck(each))
	}

	re_ := false
	if len(diff_r) == 1 {
		re_ = diff_r[0]
		if re_ == true {
			return 1, nil
		}
		return 0, nil
	}
	for index, each := range diff_r {
		if index == 0 {
			continue
		}
		op := and[index-1]
		if index == 1 {
			re_ = iFCheCKAnd(diff_r[0], each, op)
		} else {
			re_ = iFCheCKAnd(re_, each, op)
		}
	}
	if re_ == true {
		return 1, nil
	}
	return 0, nil
}

func iFSGetOperator(a string) string {
	for _, each := range a {
		each_s := string(each)
		if each_s == ">" || each_s == "<" || each_s == "=" {
			return each_s
		}
	}
	return ""
}

func iFCheCKAnd(a, b bool, op string) bool {
	switch op {
	case "&":
		return a && b
	case "||":
		return a || b
	}
	return false
}

func iFSCheck(e string) bool {

	op := iFSGetOperator(e)

	ab := strings.FieldsFunc(e, func(r rune) bool {
		return r == '>' || r == '<' || r == '=' || r == '≤' || r == '≥'
	})
	a, _ := strconv.ParseFloat(ab[0], 64)
	b, _ := strconv.ParseFloat(ab[1], 64)
	switch op {
	case ">":
		return a > b
	case "<":
		return a < b
	case "=":
		return a == b
	case "≤":
		return a <= b
	case "≥":
		return a >= b
	}
	return false
}

//Calculate func 计算方法 -1*(-9--8/-4)/(1-9)*8--8
//1 四舍五入 2 向上取整 3 向下取整
func Calculate(expr string, outcome_state, decimal_places int) (float64, error) {

	var step int = 0

L:
	exprSlice := strings.Split(expr, "")
	index := strings.LastIndex(expr, "(")
	lenExpr := len(exprSlice)
	if lenExpr == step {
		return 0, errors.New("计算死循环,请检查公式是否存在中文字符")
	}
	step = lenExpr
	if index != -1 {
		var strBuff bytes.Buffer
		for i := index + 1; i < lenExpr; i++ {
			if exprSlice[i] != ")" {
				strBuff.WriteString(exprSlice[i])
				continue
			}
			break
		}
		r := strBuff.String()

		tmpResult, err := count(r)
		if err != nil {
			return 0, err
		}
		strBuff.Reset()
		strBuff.WriteString("(")
		strBuff.WriteString(r)
		strBuff.WriteString(")")
		tempResutStr := strconv.FormatFloat(tmpResult, 'f', 3, 64)
		if len(strBuff.String()) != len(tempResutStr) {
			expr = strings.Replace(expr, strBuff.String(), tempResutStr, -1)
		} else {
			expr = strings.Replace(expr, strBuff.String(), strconv.FormatFloat(tmpResult, 'f', 4, 64), -1)
		}
		goto L
	}
	tmpResult, err := count(expr)
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

func round(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}

func count(data string) (float64, error) {
	if 1 == len(data) {
		return strconv.ParseFloat(data, 64)
	}
	arr := generateRPN(data)
	return calculateRPN(arr)
}

func calculateRPN(datas []string) (float64, error) {
	var stack linkStack
	var flag bool = false
	var symbolChar string
	stack.init()
	for i := 0; i < len(datas); i++ {
		if isNumberString(datas[i]) {
			if f, err := strconv.ParseFloat(datas[i], 64); err != nil {
				return 0, errors.New(fmt.Sprintf("计算解析错误,无法将%s,转化成一个float64类型,请检查公式", datas[i]))
			} else {
				if flag { //处理同时两个符号
					switch symbolChar {
					case "-":
						stack.push(0 - f)
					case "+":
						stack.push(f)
					}
					flag = false
					continue
				}
				stack.push(f)
			}
		} else {
			p1 := stack.pop()
			p2 := stack.pop()
			if p2 == nil && !isNumberString(datas[i]) { //如果p2为空 同时两个符号
				stack.push(p1)
				symbolChar = datas[i]
				flag = true
				continue
			}
			f1 := p1.(float64)

			f2 := p2.(float64)

			p3, err := normalCalculate(f2, f1, datas[i])

			if err != nil {
				return 0, err
			}
			stack.push(p3)
		}
	}
	res := stack.pop().(float64)
	//zlog.Debugf("gfc 计算结果:%f", nil, res)
	return res, nil
}

func booltofloat(is bool) float64 {
	if is == true {
		return 1
	}
	return 0
}

func normalCalculate(a, b float64, operation string) (float64, error) {
	switch operation {
	case "*":
		return a * b, nil
	case "-":
		return a - b, nil
	case "+":
		return a + b, nil
	case "/":
		if 0 == b {
			return 0, errors.New("计算遇到除数为0，默认返回0")
		} else {
			return a / b, nil
		}
	case "&":
		return booltofloat(a == b), nil
	case ">":
		return booltofloat(a > b), nil
	case "≥":
		return booltofloat(a >= b), nil
	case "<":
		return booltofloat(a < b), nil
	case "≤":
		return booltofloat(a <= b), nil
	default:
		return 0, errors.New(fmt.Sprintf("不支持的运算符%s", operation))
	}
}

func getExprSlice(exp string) []string {
	var symbolSlic []string

	var preBytes bytes.Buffer

	expBytes := []byte(exp)
	lenExp := len(expBytes)
	for i := 0; i < lenExp; i++ {
		if !isNumber(expBytes[i]) && "" != preBytes.String() { //符号
			symbolSlic = append(symbolSlic, preBytes.String())
			symbolSlic = append(symbolSlic, string(expBytes[i]))
			preBytes.Reset()
			continue
		}
		preBytes.WriteByte(expBytes[i])
	}
	if preBytes.Len() > 0 {
		symbolSlic = append(symbolSlic, preBytes.String())
	}
	return symbolSlic
}

func generateRPN(exp string) []string {

	var stack linkStack
	stack.init()

	var spiltedStr = getExprSlice(exp)
	var datas []string

	for i := 0; i < len(spiltedStr); i++ { // 遍历每一个字符
		tmp := spiltedStr[i] //当前字符

		if !isNumberString(tmp) { //是否是数字
			// 四种情况入栈
			// 1 左括号直接入栈
			// 2 栈内为空直接入栈
			// 3 栈顶为左括号，直接入栈
			// 4 当前元素不为右括号时，在比较栈顶元素与当前元素，如果当前元素大，直接入栈。
			if tmp == "(" ||
				stack.lookTop() == nil || stack.lookTop().(string) == "(" ||
				(compareOperator(tmp, stack.lookTop().(string)) == 1 && tmp != ")") {
				stack.push(tmp)
			} else { // ) priority
				if tmp == ")" { //当前元素为右括号时，提取操作符，直到碰见左括号
					for {
						popi := stack.pop()
						if popi != nil {
							if pop := popi.(string); pop == "(" {
								break
							} else {
								datas = append(datas, pop)
							}
						}
						break
					}
				} else { //当前元素为操作符时，不断地与栈顶元素比较直到遇到比自己小的（或者栈空了），然后入栈。
					for {
						pop := stack.lookTop()
						if pop != nil && compareOperator(tmp, pop.(string)) != 1 {
							datas = append(datas, stack.pop().(string))
						} else {
							stack.push(tmp)
							break
						}
					}
				}
			}

		} else {
			datas = append(datas, tmp)
		}
	}

	//将栈内剩余的操作符全部弹出。
	for {
		if pop := stack.pop(); pop != nil {
			datas = append(datas, pop.(string))
		} else {
			break
		}
	}
	return datas
}

// if return 1, o1 > o2.
// if return 0, o1 = 02
// if return -1, o1 < o2
func compareOperator(o1, o2 string) int {
	// + - * /
	var o1Priority int
	if o1 == "+" || o1 == "-" {
		o1Priority = 1
	} else {
		o1Priority = 2
	}
	var o2Priority int
	if o2 == "+" || o2 == "-" {
		o2Priority = 1
	} else {
		o2Priority = 2
	}
	if o1Priority > o2Priority {
		return 1
	} else if o1Priority == o2Priority {
		return 0
	} else {
		return -1
	}
}

func isNumberString(o1 string) bool {
	if o1 == "+" || o1 == "-" || o1 == "*" || o1 == "/" || o1 == "(" || o1 == ")" {
		return false
	} else {
		return true
	}
}

func convertToStrings(s string) []string {
	var strs []string
	bys := []byte(s)
	var tmp string
	for i := 0; i < len(bys); i++ {
		if !isNumber(bys[i]) {
			if tmp != "" {
				strs = append(strs, tmp)
				tmp = ""
			}
			strs = append(strs, string(bys[i]))
		} else {
			tmp = tmp + string(bys[i])
		}
	}
	strs = append(strs, tmp)
	return strs
}

func isNumber(o1 byte) bool {
	if o1 == '+' || o1 == '-' || o1 == '*' || o1 == '/' || o1 == '(' || o1 == ')' {
		return false
	} else {
		return true
	}
}
