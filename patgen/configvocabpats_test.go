package patgen

import (
	"fmt"
	"slices"
	"testing"

	"cogentcore.org/lab/table"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

func TestVocab(t *testing.T) {
	NewRand(10)
	m := make(Vocab)
	AddVocabEmpty(m, "empty", 6, 3, 3)
	AddVocabPermutedBinary(m, "A", 6, 3, 3, 0.3, 0.4)
	AddVocabDrift(m, "B", 6, 0.2, "A", 0) // nOn=4*(3*3*0.3); nDrift=nOn*0.5
	AddVocabRepeat(m, "ctxt1", 6, "A", 0)
	VocabConcat(m, "AB-C", []string{"A", "B"})
	VocabSlice(m, "AB-C", []string{"A'", "B'"}, []int{0, 6, 12}) // 3 cutoffs for 2 vocabs
	VocabShuffle(m, []string{"B'"})
	AddVocabClone(m, "B''", "B'")

	keys := maps.Keys(m)
	slices.Sort(keys)
	exmap := `[A A' AB-C B B' B'' ctxt1 empty]`
	// fmt.Println("map")
	// fmt.Println(reflect.ValueOf(m).MapKeys())
	assert.Equal(t, exmap, fmt.Sprintf("%v", keys))

	exempty := `[6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   0   0 
[0 1]     0   0   0 
[0 2]     0   0   0 
[1 0]     0   0   0 
[1 1]     0   0   0 
[1 2]     0   0   0 
[2 0]     0   0   0 
[2 1]     0   0   0 
[2 2]     0   0   0 
[3 0]     0   0   0 
[3 1]     0   0   0 
[3 2]     0   0   0 
[4 0]     0   0   0 
[4 1]     0   0   0 
[4 2]     0   0   0 
[5 0]     0   0   0 
[5 1]     0   0   0 
[5 2]     0   0   0 
`
	// fmt.Println("empty")
	// fmt.Println(m["empty"].String())
	assert.Equal(t, exempty, m["empty"].String())

	exa := `[6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     1   0   1 
[1 1]     1   0   0 
[1 2]     0   0   0 
[2 0]     0   0   0 
[2 1]     1   0   0 
[2 2]     1   1   0 
[3 0]     0   1   0 
[3 1]     0   0   0 
[3 2]     1   0   1 
[4 0]     1   0   1 
[4 1]     0   1   0 
[4 2]     0   0   0 
[5 0]     1   0   0 
[5 1]     0   0   1 
[5 2]     1   0   0 
`

	// fmt.Println("A")
	// fmt.Println(m["A"].String())
	assert.Equal(t, exa, m["A"].String())

	exb := `[6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     0   1   0 
[1 1]     1   0   1 
[1 2]     0   0   0 
[2 0]     0   1   0 
[2 1]     1   0   1 
[2 2]     0   0   0 
[3 0]     0   1   1 
[3 1]     1   0   0 
[3 2]     0   0   0 
[4 0]     0   1   1 
[4 1]     1   0   0 
[4 2]     0   0   0 
[5 0]     0   1   1 
[5 1]     1   0   0 
[5 2]     0   0   0 
`

	// fmt.Println("B")
	// fmt.Println(m["B"].String())
	assert.Equal(t, exb, m["B"].String())

	exctxt := `[6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     0   1   0 
[1 1]     1   0   1 
[1 2]     0   0   0 
[2 0]     0   1   0 
[2 1]     1   0   1 
[2 2]     0   0   0 
[3 0]     0   1   0 
[3 1]     1   0   1 
[3 2]     0   0   0 
[4 0]     0   1   0 
[4 1]     1   0   1 
[4 2]     0   0   0 
[5 0]     0   1   0 
[5 1]     1   0   1 
[5 2]     0   0   0 
`

	// fmt.Println("ctxt1")
	// fmt.Println(m["ctxt1"].String())
	assert.Equal(t, exctxt, m["ctxt1"].String())

	exabc := `[12 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     1   0   1 
[1 1]     1   0   0 
[1 2]     0   0   0 
[2 0]     0   0   0 
[2 1]     1   0   0 
[2 2]     1   1   0 
[3 0]     0   1   0 
[3 1]     0   0   0 
[3 2]     1   0   1 
[4 0]     1   0   1 
[4 1]     0   1   0 
[4 2]     0   0   0 
[5 0]     1   0   0 
[5 1]     0   0   1 
[5 2]     1   0   0 
[6 0]     0   1   0 
[6 1]     1   0   1 
[6 2]     0   0   0 
[7 0]     0   1   0 
[7 1]     1   0   1 
[7 2]     0   0   0 
[8 0]     0   1   0 
[8 1]     1   0   1 
[8 2]     0   0   0 
[9 0]     0   1   1 
[9 1]     1   0   0 
[9 2]     0   0   0 
[10 0]    0   1   1 
[10 1]    1   0   0 
[10 2]    0   0   0 
[11 0]    0   1   1 
[11 1]    1   0   0 
[11 2]    0   0   0 
`
	// fmt.Println("AB-C")
	// fmt.Println(m["AB-C"].String())
	assert.Equal(t, exabc, m["AB-C"].String())

	exap := `[6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     1   0   1 
[1 1]     1   0   0 
[1 2]     0   0   0 
[2 0]     0   0   0 
[2 1]     1   0   0 
[2 2]     1   1   0 
[3 0]     0   1   0 
[3 1]     0   0   0 
[3 2]     1   0   1 
[4 0]     1   0   1 
[4 1]     0   1   0 
[4 2]     0   0   0 
[5 0]     1   0   0 
[5 1]     0   0   1 
[5 2]     1   0   0 
`

	// fmt.Println("A'")
	// fmt.Println(m["A'"].String())
	assert.Equal(t, exap, m["A'"].String())

	exbp := `[6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     0   1   1 
[1 1]     1   0   0 
[1 2]     0   0   0 
[2 0]     0   1   1 
[2 1]     1   0   0 
[2 2]     0   0   0 
[3 0]     0   1   0 
[3 1]     1   0   1 
[3 2]     0   0   0 
[4 0]     0   1   1 
[4 1]     1   0   0 
[4 2]     0   0   0 
[5 0]     0   1   0 
[5 1]     1   0   1 
[5 2]     0   0   0 
`

	// fmt.Println("B'")
	// fmt.Println(m["B'"].String())
	assert.Equal(t, exbp, m["B'"].String())

	exbpp := `[6 3 3]
[r r c] [0] [1] [2] 
[0 0]     0   1   0 
[0 1]     1   0   1 
[0 2]     0   0   0 
[1 0]     0   1   1 
[1 1]     1   0   0 
[1 2]     0   0   0 
[2 0]     0   1   1 
[2 1]     1   0   0 
[2 2]     0   0   0 
[3 0]     0   1   0 
[3 1]     1   0   1 
[3 2]     0   0   0 
[4 0]     0   1   1 
[4 1]     1   0   0 
[4 2]     0   0   0 
[5 0]     0   1   0 
[5 1]     1   0   1 
[5 2]     0   0   0 
`

	// fmt.Println("B''")
	// fmt.Println(m["B''"].String())
	assert.Equal(t, exbpp, m["B''"].String())

	// config pats
	dt := table.New("TrainAB")
	InitPats(dt, "TrainAB", "describe", "Input", "ECout", 6, 3, 2, 3, 3)
	MixPats(dt, m, "Input", []string{"A", "B", "ctxt1", "ctxt1", "empty", "B'"})
	MixPats(dt, m, "ECout", []string{"A", "B", "ctxt1", "ctxt1", "empty", "B'"})

	// try shuffle
	Shuffle(dt, []int{0, 1, 2, 3, 4, 5}, []string{"Input", "ECout"}, false)

	exip := `Input [6 3 2 3 3]
[r r c r c] [0 0] [0 1] [0 2] [1 0] [1 1] [1 2] 
[0 0 0]         0     0     0     0     1     0 
[0 0 1]         1     0     0     1     0     1 
[0 0 2]         1     1     0     0     0     0 
[0 1 0]         0     1     0     0     1     0 
[0 1 1]         1     0     1     1     0     1 
[0 1 2]         0     0     0     0     0     0 
[0 2 0]         0     0     0     0     1     1 
[0 2 1]         0     0     0     1     0     0 
[0 2 2]         0     0     0     0     0     0 
[1 0 0]         1     0     1     0     1     1 
[1 0 1]         0     1     0     1     0     0 
[1 0 2]         0     0     0     0     0     0 
[1 1 0]         0     1     0     0     1     0 
[1 1 1]         1     0     1     1     0     1 
[1 1 2]         0     0     0     0     0     0 
[1 2 0]         0     0     0     0     1     1 
[1 2 1]         0     0     0     1     0     0 
[1 2 2]         0     0     0     0     0     0 
[2 0 0]         1     0     1     0     1     1 
[2 0 1]         0     1     0     1     0     0 
[2 0 2]         0     0     0     0     0     0 
[2 1 0]         0     1     0     0     1     0 
[2 1 1]         1     0     1     1     0     1 
[2 1 2]         0     0     0     0     0     0 
[2 2 0]         0     0     0     0     1     1 
[2 2 1]         0     0     0     1     0     0 
[2 2 2]         0     0     0     0     0     0 
[3 0 0]         1     0     0     0     1     1 
[3 0 1]         0     0     1     1     0     0 
[3 0 2]         1     0     0     0     0     0 
[3 1 0]         0     1     0     0     1     0 
[3 1 1]         1     0     1     1     0     1 
[3 1 2]         0     0     0     0     0     0 
[3 2 0]         0     0     0     0     1     0 
[3 2 1]         0     0     0     1     0     1 
[3 2 2]         0     0     0     0     0     0 
[4 0 0]         0     0     0     0     1     0 
[4 0 1]         1     0     0     1     0     1 
[4 0 2]         1     1     0     0     0     0 
[4 1 0]         0     1     0     0     1     0 
[4 1 1]         1     0     1     1     0     1 
[4 1 2]         0     0     0     0     0     0 
[4 2 0]         0     0     0     0     1     1 
[4 2 1]         0     0     0     1     0     0 
[4 2 2]         0     0     0     0     0     0 
[5 0 0]         1     0     0     0     1     1 
[5 0 1]         0     0     1     1     0     0 
[5 0 2]         1     0     0     0     0     0 
[5 1 0]         0     1     0     0     1     0 
[5 1 1]         1     0     1     1     0     1 
[5 1 2]         0     0     0     0     0     0 
[5 2 0]         0     0     0     0     1     0 
[5 2 1]         0     0     0     1     0     1 
[5 2 2]         0     0     0     0     0     0 
`
	// fmt.Println("Input Pats")
	// fmt.Println(dt.Column("Input").Shape().Sizes)
	// fmt.Println(dt.Column("Input").String())
	assert.Equal(t, []int{6, 3, 2, 3, 3}, dt.Column("Input").Shape().Sizes)
	assert.Equal(t, exip, dt.Column("Input").String())

	exop := `ECout [6 3 2 3 3]
[r r c r c] [0 0] [0 1] [0 2] [1 0] [1 1] [1 2] 
[0 0 0]         0     0     0     0     1     0 
[0 0 1]         1     0     0     1     0     1 
[0 0 2]         1     1     0     0     0     0 
[0 1 0]         0     1     0     0     1     0 
[0 1 1]         1     0     1     1     0     1 
[0 1 2]         0     0     0     0     0     0 
[0 2 0]         0     0     0     0     1     1 
[0 2 1]         0     0     0     1     0     0 
[0 2 2]         0     0     0     0     0     0 
[1 0 0]         1     0     1     0     1     1 
[1 0 1]         0     1     0     1     0     0 
[1 0 2]         0     0     0     0     0     0 
[1 1 0]         0     1     0     0     1     0 
[1 1 1]         1     0     1     1     0     1 
[1 1 2]         0     0     0     0     0     0 
[1 2 0]         0     0     0     0     1     1 
[1 2 1]         0     0     0     1     0     0 
[1 2 2]         0     0     0     0     0     0 
[2 0 0]         1     0     1     0     1     1 
[2 0 1]         0     1     0     1     0     0 
[2 0 2]         0     0     0     0     0     0 
[2 1 0]         0     1     0     0     1     0 
[2 1 1]         1     0     1     1     0     1 
[2 1 2]         0     0     0     0     0     0 
[2 2 0]         0     0     0     0     1     1 
[2 2 1]         0     0     0     1     0     0 
[2 2 2]         0     0     0     0     0     0 
[3 0 0]         1     0     0     0     1     1 
[3 0 1]         0     0     1     1     0     0 
[3 0 2]         1     0     0     0     0     0 
[3 1 0]         0     1     0     0     1     0 
[3 1 1]         1     0     1     1     0     1 
[3 1 2]         0     0     0     0     0     0 
[3 2 0]         0     0     0     0     1     0 
[3 2 1]         0     0     0     1     0     1 
[3 2 2]         0     0     0     0     0     0 
[4 0 0]         0     0     0     0     1     0 
[4 0 1]         1     0     0     1     0     1 
[4 0 2]         1     1     0     0     0     0 
[4 1 0]         0     1     0     0     1     0 
[4 1 1]         1     0     1     1     0     1 
[4 1 2]         0     0     0     0     0     0 
[4 2 0]         0     0     0     0     1     1 
[4 2 1]         0     0     0     1     0     0 
[4 2 2]         0     0     0     0     0     0 
[5 0 0]         1     0     0     0     1     1 
[5 0 1]         0     0     1     1     0     0 
[5 0 2]         1     0     0     0     0     0 
[5 1 0]         0     1     0     0     1     0 
[5 1 1]         1     0     1     1     0     1 
[5 1 2]         0     0     0     0     0     0 
[5 2 0]         0     0     0     0     1     0 
[5 2 1]         0     0     0     1     0     1 
[5 2 2]         0     0     0     0     0     0 
`

	// fmt.Println("ECout Pats")
	// fmt.Println(dt.Column("ECout").Shape().Sizes)
	// fmt.Println(dt.Column("ECout").String())

	assert.Equal(t, []int{6, 3, 2, 3, 3}, dt.Column("ECout").Shape().Sizes)
	assert.Equal(t, exop, dt.Column("ECout").String())
}
