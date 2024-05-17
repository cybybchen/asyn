package blockword

var gBlockword = &BlockWord{}

func Init() {
	gBlockword = New()
	gBlockword.Load()
}

func Global() *BlockWord {
	return gBlockword
}
