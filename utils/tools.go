package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func RemoveRepeatInt(s []int) []int {
	result := []int{}
	m := make(map[int]bool)
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}

func StringArrayToINArray(temp []string) string {
	for index, each := range temp {
		if each == "" {
			continue
		}
		temp[index] = fmt.Sprintf(" '%s'", each)
	}
	fmt.Println(strings.Join(temp, ","))

	return strings.Join(temp, ",")
}

func IntArrayToStr(temp []int) string {
	s := make([]string, len(temp))
	for i, v := range temp {
		s[i] = strconv.Itoa(int(v))
	}
	return "{" + strings.Join(s, ",") + "}"
}

func StringArrayToStr(temp []string) string {
	return "{" + strings.Join(temp, ",") + "}"
}

func StringValueToIntArray(temp string) []int {
	if len(temp) == 2 {
		return []int{}
	}
	temp = temp[1 : len(temp)-1]
	temps := strings.Split(temp, ",")
	re_temps := []int{}
	for _, each := range temps {
		each_i, _ := strconv.Atoi(each)
		re_temps = append(re_temps, each_i)
	}
	return re_temps
}
func StringValueToStrArray(temp string) []string {
	if len(temp) <= 2 {
		return []string{}
	}
	temp = temp[1 : len(temp)-1]
	temps := strings.Split(temp, ",")
	re_temps := []string{}
	for _, each := range temps {
		re_temps = append(re_temps, each)
	}
	return re_temps
}

func GetLetterByIdx(idx int) string {
	if idx < 27 {
		return string(64 + idx)
	} else {
		index := idx / 26
		index1 := idx % 26
		return fmt.Sprintf("%s%s", string(64+index), string(index1))
	}
}

func GetNextLetters(this_letter string) string {
	codeAscaii := []rune(this_letter)
	if len(codeAscaii) == 1 {
		if codeAscaii[0] < 90 {
			return string(codeAscaii[0] + 1)
		} else {
			return "AA"
		}
	} else {
		if codeAscaii[1] == 90 {
			return fmt.Sprintf("%s%s", string(codeAscaii[0]+1), "A")
		} else {
			return fmt.Sprintf("%s%s", string(codeAscaii[0]), string(codeAscaii[1]+1))
		}
	}
}

func SqlArrayValueStr(temp []string) string {
	return "{" + strings.Join(temp, ",") + "}"
}

func SqlArrayValue(temp []int) string {
	s := make([]string, len(temp))
	for i, v := range temp {
		s[i] = strconv.Itoa(int(v))
	}
	return "{" + strings.Join(s, ",") + "}"
}
func SqlStringValue(temp string) []int {
	if temp == "" {
		return []int{}
	}

	temp = temp[1 : len(temp)-1]
	temps := strings.Split(temp, ",")
	re_temps := []int{}
	for _, each := range temps {
		each_i, _ := strconv.Atoi(each)
		re_temps = append(re_temps, each_i)
	}
	return re_temps
}

func RemoveRepByMap(slc []string) []string {
	result := []string{}
	tempMap := map[string]byte{}
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l {
			result = append(result, e)
		}
	}
	return result
}

func IsExistInt(temps []int, temp int) bool {
	for _, each := range temps {
		if each == temp {
			return true
		}
	}
	return false
}

func IsExistStr(temps []string, temp string) bool {
	for _, each := range temps {
		if each == temp {
			return true
		}
	}
	return false
}

func DeleteRepeat(list []string) []string {
	mapdata := make(map[string]interface{})
	if len(list) <= 0 {
		return nil
	}
	for _, v := range list {
		mapdata[v] = "true"
	}
	var datas []string
	for k, _ := range mapdata {
		if k == "" {
			continue
		}
		datas = append(datas, k)
	}
	return datas
}

func StrInArray(temp string, temps []string) bool {
	for _, each := range temps {
		if each == temp {
			return true
		}
	}
	return false
}

func MD5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

func GenValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < width; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}
