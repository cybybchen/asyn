package blockword

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func ExampleBlockWord_IsValid() {
	bw := New()
	bw.Load()
	fmt.Println(bw.IsValid("毛泽东"))
	fmt.Println(bw.IsValid("毛泽|东"))
	fmt.Println(bw.IsValid("毛"))
	fmt.Println(bw.IsValid("我是毛毛毛毛毛毛毛毛"))
	fmt.Println(bw.ReplaceDefault("毛泽东"))
	fmt.Println(bw.ReplaceDefault("毛泽|东"))
	fmt.Println(bw.ReplaceDefault("毛 泽东"))
	fmt.Println(bw.ReplaceDefault("毛泽东一生反孔"))
	fmt.Println(bw.ReplaceDefault("毛&泽|东"))
	fmt.Println(bw.ReplaceDefault("5.5狗粮模具"))

	// Output:
	// false
	// false
	// true
	// true
	// ***
	// ****
	// ****
	// *******
	// *****
	// *******
}

func TestBlockWord_IsValid(t *testing.T) {
	bw := New()
	bw.Load()

	b, err := load("testdata/words_normal.txt.bk")
	if err != nil {
		t.Errorf("验证错误 [%v]", err)
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if bw.IsValid(line) {
			t.Errorf("验证错误 [%v]", line)
		}
	}
}

func TestBlockWord_ReplaceDefault(t *testing.T) {
	bw := New()
	bw.Load()

	lines, err := loadSlice("testdata/words_normal.txt.bk")
	if err != nil {
		t.Errorf("验证错误 [%v]", err)
	}
	for i, line := range lines {
		newStr := bw.ReplaceDefault(line)
		for _, ch := range newStr {
			if ch != '*' {
				t.Errorf("行数[%d] 验证错误 [%v][%s]", i, line, newStr)
			}
		}
	}
}

func BenchmarkBlockWord_IsValid(b *testing.B) {
	b.Run("10个字合法", func(b *testing.B) {
		b.ReportAllocs()
		bw := New()
		bw.Load()
		var str string
		for i := 0; i < 1; i++ {
			str += "我是毛毛毛毛毛毛毛毛"
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bw.IsValid(str)
		}
	})
	b.Run("10个字合法", func(b *testing.B) {
		b.ReportAllocs()
		bw := New()
		bw.Load()
		var str string
		for i := 0; i < 1; i++ {
			str += "我是毛毛毛毛毛毛毛毛"
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bw.IsValidOnlyNormal(str)
		}
	})

	b.Run("100个字合法", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		bw := New()
		bw.Load()
		var str string
		for i := 0; i < 10; i++ {
			str += "我是毛毛毛毛毛毛毛毛"
		}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			_ = bw.IsValid(str)
		}
	})
	b.Run("100个字合法", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		bw := New()
		bw.Load()
		var str string
		for i := 0; i < 10; i++ {
			str += "我是毛毛毛毛毛毛毛毛"
		}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			_ = bw.IsValidOnlyNormal(str)
		}
	})
	//
	//b.Run("1000个字合法", func(b *testing.B) {
	//	b.StopTimer()
	//	bw, _ := New("testdata/cn.bw")
	//	var str string
	//	for i := 0; i < 100; i++ {
	//		str += "我是毛毛毛毛毛毛毛毛"
	//	}
	//	b.StartTimer()
	//	for i := 0; i < b.N; i++ {
	//		_ = bw.IsValid(str)
	//	}
	//})
	//
	//b.Run("1000个字不合法", func(b *testing.B) {
	//	b.StopTimer()
	//	bw, _ := New("testdata/cn.bw")
	//	var str string
	//	for i := 0; i < 100; i++ {
	//		str += "我是毛泽东"
	//	}
	//	b.StartTimer()
	//	for i := 0; i < b.N; i++ {
	//		_ = bw.IsValid(str)
	//	}
	//})

}

func TestAllSubStrings2(t *testing.T) {
	str1 := "((.*?)东(.*?)条(.*?)英(.*?)机(.*?))|((.*?)丽(.*?)张(.*?)高(.*?))"
	reg1 := regexp.MustCompile(str1)
	fmt.Println(reg1.MatchString("丽张高"))
}

func BenchmarkRegexp(b *testing.B) {
	b.Run("合并", func(b *testing.B) {
		str1 := "((.*?)东(.*?)条(.*?)英(.*?)机(.*?))|((.*?)丽(.*?)张(.*?)高(.*?))|((.*?)丽(.*?)張(.*?)高(.*?))|((.*?)丽(.*?)高(.*?)张(.*?))"
		reg1 := regexp.MustCompile(str1)
		for i := 0; i < b.N; i++ {
			reg1.MatchString("丽高2张")
		}
	})

	b.Run("合并", func(b *testing.B) {
		str1 := "(.*?)东(.*?)条(.*?)英(.*?)机(.*?)"
		reg1 := regexp.MustCompile(str1)
		str2 := "(.*?)丽(.*?)张(.*?)高(.*?)"
		reg2 := regexp.MustCompile(str2)
		str3 := "(.*?)丽(.*?)張(.*?)高(.*?)"
		reg3 := regexp.MustCompile(str3)
		str4 := "(.*?)丽(.*?)高(.*?)张(.*?)"
		reg4 := regexp.MustCompile(str4)
		for i := 0; i < b.N; i++ {
			reg1.MatchString("丽高2张")
			reg2.MatchString("丽高2张")
			reg3.MatchString("丽高2张")
			reg4.MatchString("丽高2张")
		}
	})
}
