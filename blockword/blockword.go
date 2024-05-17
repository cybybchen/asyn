package blockword

import (
	"regexp"
	"sync"

	"gitlab.sunborngame.com/base/log"
)

// BlockWord is the representation of a block-word manager.
// A BlockWord is safe for concurrent use by multiple goroutines.
type BlockWord struct {
	sync.RWMutex

	wordsNormal map[rune]*Word
	regWords    []*regexp.Regexp
	regWord     *regexp.Regexp

	invalid map[rune]struct{}
}

type Word struct {
	Content rune           //内容
	SubWord map[rune]*Word //屏蔽子集合
	IsEnd   bool           //是否为结束
}

// New
func New() *BlockWord {
	bw := &BlockWord{invalid: invalidWord}
	return bw
}

// Load 可以异步加载，加载完毕后，上锁替换指针
func (bw *BlockWord) Load() {
	// 1.加载普通文件
	originWords, err := loadNormal()
	if err != nil {
		log.Error("block word error : [%v]", err)
		return
	}
	bw.Lock()
	bw.wordsNormal = originWords
	bw.Unlock()
	// 2.加载正则文件
	regWords, regWord, err := loadRegexp()
	if err != nil {
		log.Error("block word error : [%v]", err)
		return
	}
	bw.Lock()
	bw.regWords = regWords
	bw.regWord = regWord
	bw.Unlock()
	log.Debug("block word load success, normal size = %d, regexps size = %d", len(bw.wordsNormal), len(bw.regWords))
}

// IsValid 判断字符串是否合法 true-合法 false-非法
func (bw *BlockWord) IsValid(str string) bool {
	if len(str) == 0 {
		return true
	}
	const char = '*'
	if _, ok := bw.isNormalValid(str, char); !ok {
		return false
	}

	return bw.isRegexpValid(str)
}

// IsValidOnlyNormal 不进行正则的，判断字符串是否合法 true-合法 false-非法
func (bw *BlockWord) IsValidOnlyNormal(str string) bool {
	if len(str) == 0 {
		return true
	}
	const char = '*'
	if _, ok := bw.isNormalValid(str, char); !ok {
		return false
	}

	return true
}

// ReplaceDefault 将非法字符替换成指定字符
func (bw *BlockWord) Replace(str string, char rune) string {
	if len(str) == 0 {
		return str
	}
	str, _ = bw.isNormalValid(str, char)
	return str
}

// ReplaceDefault 将非法字符替换成*
func (bw *BlockWord) ReplaceDefault(str string) string {
	const char = '*'
	return bw.Replace(str, char)
}

func (bw *BlockWord) isRegexpValid(str string) bool {
	bw.RLock()
	words := bw.regWords
	bw.RUnlock()
	for _, reg := range words {
		if reg.MatchString(str) {
			return false
		}
	}
	return true
}

func (bw *BlockWord) isNormalValid(str string, char rune) (newStr string, valid bool) {
	bw.RLock()
	wordsNormal := bw.wordsNormal
	bw.RUnlock()

	if len(wordsNormal) == 0 {
		return str, true
	}

	valid = true
	chars := []rune(str)

	for i := 0; i < len(chars); i++ {
		fmd, ok := wordsNormal[chars[i]]
		if !ok {
			continue
		}
		if len(chars) == 1 { //只有一个字符
			if fmd.IsEnd {
				valid = false
				chars[0] = char
			}
		} else if ok, index := filter(chars, i+1, fmd.SubWord, bw.invalid, char); ok {
			valid = false
			chars[i] = char
			i = index
		}
	}
	return string(chars), valid
}

// 递归调用检查屏蔽字
func filter(chars []rune, index int, filterWord map[rune]*Word, invalid map[rune]struct{}, char rune) (bool, int) {
	if len(chars) <= index {
		return false, index
	}
	currChar := chars[index]
	fw, ok := filterWord[currChar]

	if _, ok := invalid[currChar]; ok {
		subWord := filterWord
		if fw != nil {
			subWord = fw.SubWord
		}
		if ok, i := filter(chars, index+1, subWord, invalid, char); ok {
			chars[index] = char
			return true, i
		}
	}
	if !ok || fw == nil {
		return false, index
	}
	if fw.IsEnd {
		chars[index] = char
		ok, i := filter(chars, index+1, fw.SubWord, invalid, char)
		if ok {
			return true, i
		}
		return true, index
	} else if ok, i := filter(chars, index+1, fw.SubWord, invalid, char); ok {
		chars[index] = char
		return true, i
	}
	return false, index
}
