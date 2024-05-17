package blockword

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func load(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func loadSlice(filename string) ([]string, error) {
	str, err := load(filename)
	if err != nil {
		return nil, err
	}
	str = strings.Replace(str, "\r", "", -1)
	rows := strings.Split(str, "\n")
	return rows, nil
}

func loadRegexp() ([]*regexp.Regexp, *regexp.Regexp, error) {
	//newLines := make([]string, 0, len(cache_tables.ChatMaskRegexTemplateSTM.GetContents()))
	//result := make([]*regexp.Regexp, 0, len(cache_tables.ChatMaskRegexTemplateSTM.GetContents()))
	//for _, cfg := range cache_tables.ChatMaskRegexTemplateSTM.GetContents() {
	//	line := strings.TrimSpace(cfg.WordRegex())
	//	if line == "" {
	//		continue
	//	}
	//	result = append(result, reg)
	//	line = "(" + line + ")"
	//	newLines = append(newLines, line)
	//}
	//newStr := strings.Join(newLines, "|")
	//reg, err := regexp.Compile(newStr)
	//if err != nil {
	//	return result, nil, err
	//}
	//return result, reg, err
	return []*regexp.Regexp{}, nil, nil
}

func loadNormal() (words map[rune]*Word, err error) {
	words = make(map[rune]*Word)
	//for _, cfg := range cache_tables.ChatMaskTemplateSTM.GetContents() {
	//	if len(cfg.Word()) == 0 {
	//		continue
	//	}
	//	rowr := []rune(cfg.Word())
	//	loadWithIndex(words, rowr, 0)
	//}

	return words, nil
}

func loadWithIndex(li map[rune]*Word, rowr []rune, index int) bool {
	if len(rowr) <= index {
		return true
	}
	fmd, ok := li[rowr[index]]
	if !ok {
		fmd = new(Word)
		fmd.Content = rowr[index]
		fmd.SubWord = make(map[rune]*Word)
		li[rowr[index]] = fmd
	}
	index++
	end := loadWithIndex(fmd.SubWord, rowr, index)
	if !fmd.IsEnd {
		fmd.IsEnd = end
	}
	return false
}
