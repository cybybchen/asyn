package blockword

import (
	"strings"
)

const InvalidWords = " ,~,!,@,#,$,%,^,&,*,(,),_,-,+,=,?,<,>,.,—,，,。,/,\\,|,《,》,？,;,:,：,',‘,；,“,"

var invalidWord = strToMap(InvalidWords) //无效词汇，不参与敏感词汇判断直接忽略

func strToMap(str string) map[rune]struct{} {
	invalidWord := make(map[rune]struct{})
	words := strings.Split(str, ",")
	for _, v := range words {
		if len(v) == 0 {
			continue
		}
		r := []rune(v)
		invalidWord[r[0]] = struct{}{}
	}
	return invalidWord
}

// 设置无效词汇，不参与敏感词汇判断直接忽略
func (bw *BlockWord) InvalidWord(str string) {
	bw.invalid = strToMap(str)
}
