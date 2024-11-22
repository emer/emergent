package patgen

import (
	"fmt"
	"slices"
	"testing"

	"cogentcore.org/core/tensor/table"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

func TestVocab(t *testing.T) {
	NewRand(10)
	m := make(Vocab)
	AddVocabEmpty(m, "empty", 6, 3, 3)
	AddVocabPermutedBinary(m, "A", 6, 3, 3, 0.3, 0.5)
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

	exempty := `Tensor: [row: 6, Y: 3, X: 3]
[0 0]:       0       0       0 
[0 1]:       0       0       0 
[0 2]:       0       0       0 
[1 0]:       0       0       0 
[1 1]:       0       0       0 
[1 2]:       0       0       0 
[2 0]:       0       0       0 
[2 1]:       0       0       0 
[2 2]:       0       0       0 
[3 0]:       0       0       0 
[3 1]:       0       0       0 
[3 2]:       0       0       0 
[4 0]:       0       0       0 
[4 1]:       0       0       0 
[4 2]:       0       0       0 
[5 0]:       0       0       0 
[5 1]:       0       0       0 
[5 2]:       0       0       0 
`
	// fmt.Println("empty")
	// fmt.Println(m["empty"].String())
	assert.Equal(t, exempty, m["empty"].String())

	exa := `Tensor: [row: 6, Y: 3, X: 3]
[0 0]:       0       1       1 
[0 1]:       0       0       0 
[0 2]:       1       0       0 
[1 0]:       0       0       0 
[1 1]:       1       0       1 
[1 2]:       0       1       0 
[2 0]:       1       0       1 
[2 1]:       0       0       0 
[2 2]:       0       1       0 
[3 0]:       0       1       0 
[3 1]:       0       1       0 
[3 2]:       0       0       1 
[4 0]:       0       0       0 
[4 1]:       0       1       0 
[4 2]:       1       1       0 
[5 0]:       0       0       0 
[5 1]:       1       0       0 
[5 2]:       1       0       1 
`

	// fmt.Println("A")
	// fmt.Println(m["A"].String())
	assert.Equal(t, exa, m["A"].String())

	exb := `Tensor: [row: 6, Y: 3, X: 3]
[0 0]:       0       1       1 
[0 1]:       0       0       0 
[0 2]:       1       0       0 
[1 0]:       0       0       1 
[1 1]:       0       1       0 
[1 2]:       1       0       0 
[2 0]:       0       0       1 
[2 1]:       0       1       0 
[2 2]:       1       0       0 
[3 0]:       0       1       0 
[3 1]:       0       1       0 
[3 2]:       1       0       0 
[4 0]:       0       1       0 
[4 1]:       0       1       0 
[4 2]:       1       0       0 
[5 0]:       0       1       1 
[5 1]:       0       0       0 
[5 2]:       1       0       0 
`

	// fmt.Println("B")
	// fmt.Println(m["B"].String())
	assert.Equal(t, exb, m["B"].String())

	exctxt := `Tensor: [row: 6, Y: 3, X: 3]
[0 0]:       0       1       1 
[0 1]:       0       0       0 
[0 2]:       1       0       0 
[1 0]:       0       1       1 
[1 1]:       0       0       0 
[1 2]:       1       0       0 
[2 0]:       0       1       1 
[2 1]:       0       0       0 
[2 2]:       1       0       0 
[3 0]:       0       1       1 
[3 1]:       0       0       0 
[3 2]:       1       0       0 
[4 0]:       0       1       1 
[4 1]:       0       0       0 
[4 2]:       1       0       0 
[5 0]:       0       1       1 
[5 1]:       0       0       0 
[5 2]:       1       0       0 
`

	// fmt.Println("ctxt1")
	// fmt.Println(m["ctxt1"].String())
	assert.Equal(t, exctxt, m["ctxt1"].String())

	exabc := `Tensor: [row: 12, Y: 3, X: 3]
[0 0]:       0       1       1 
[0 1]:       0       0       0 
[0 2]:       1       0       0 
[1 0]:       0       0       0 
[1 1]:       1       0       1 
[1 2]:       0       1       0 
[2 0]:       1       0       1 
[2 1]:       0       0       0 
[2 2]:       0       1       0 
[3 0]:       0       1       0 
[3 1]:       0       1       0 
[3 2]:       0       0       1 
[4 0]:       0       0       0 
[4 1]:       0       1       0 
[4 2]:       1       1       0 
[5 0]:       0       0       0 
[5 1]:       1       0       0 
[5 2]:       1       0       1 
[6 0]:       0       1       1 
[6 1]:       0       0       0 
[6 2]:       1       0       0 
[7 0]:       0       0       1 
[7 1]:       0       1       0 
[7 2]:       1       0       0 
[8 0]:       0       0       1 
[8 1]:       0       1       0 
[8 2]:       1       0       0 
[9 0]:       0       1       0 
[9 1]:       0       1       0 
[9 2]:       1       0       0 
[10 0]:       0       1       0 
[10 1]:       0       1       0 
[10 2]:       1       0       0 
[11 0]:       0       1       1 
[11 1]:       0       0       0 
[11 2]:       1       0       0 
`
	// fmt.Println("AB-C")
	// fmt.Println(m["AB-C"].String())
	assert.Equal(t, exabc, m["AB-C"].String())

	exap := `Tensor: [row: 6, Y: 3, X: 3]
[0 0]:       0       1       1 
[0 1]:       0       0       0 
[0 2]:       1       0       0 
[1 0]:       0       0       0 
[1 1]:       1       0       1 
[1 2]:       0       1       0 
[2 0]:       1       0       1 
[2 1]:       0       0       0 
[2 2]:       0       1       0 
[3 0]:       0       1       0 
[3 1]:       0       1       0 
[3 2]:       0       0       1 
[4 0]:       0       0       0 
[4 1]:       0       1       0 
[4 2]:       1       1       0 
[5 0]:       0       0       0 
[5 1]:       1       0       0 
[5 2]:       1       0       1 
`

	// fmt.Println("A'")
	// fmt.Println(m["A'"].String())
	assert.Equal(t, exap, m["A'"].String())

	exbp := `Tensor: [row: 6, Y: 3, X: 3]
[0 0]:       0       1       1 
[0 1]:       0       0       0 
[0 2]:       1       0       0 
[1 0]:       0       1       0 
[1 1]:       0       1       0 
[1 2]:       1       0       0 
[2 0]:       0       0       1 
[2 1]:       0       1       0 
[2 2]:       1       0       0 
[3 0]:       0       0       1 
[3 1]:       0       1       0 
[3 2]:       1       0       0 
[4 0]:       0       1       1 
[4 1]:       0       0       0 
[4 2]:       1       0       0 
[5 0]:       0       1       0 
[5 1]:       0       1       0 
[5 2]:       1       0       0 
`

	// fmt.Println("B'")
	// fmt.Println(m["B'"].String())
	assert.Equal(t, exbp, m["B'"].String())

	exbpp := `Tensor: [row: 6, Y: 3, X: 3]
[0 0]:       0       1       1 
[0 1]:       0       0       0 
[0 2]:       1       0       0 
[1 0]:       0       1       0 
[1 1]:       0       1       0 
[1 2]:       1       0       0 
[2 0]:       0       0       1 
[2 1]:       0       1       0 
[2 2]:       1       0       0 
[3 0]:       0       0       1 
[3 1]:       0       1       0 
[3 2]:       1       0       0 
[4 0]:       0       1       1 
[4 1]:       0       0       0 
[4 2]:       1       0       0 
[5 0]:       0       1       0 
[5 1]:       0       1       0 
[5 2]:       1       0       0 
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

	exip := `Tensor: [Row: 6, ySize: 3, xSize: 2, poolY: 3, poolX: 3]
[0 0 0]:       0       0       0       0       1       0 
[0 0 1]:       0       1       0       0       1       0 
[0 0 2]:       1       1       0       1       0       0 
[0 1 0]:       0       1       1       0       1       1 
[0 1 1]:       0       0       0       0       0       0 
[0 1 2]:       1       0       0       1       0       0 
[0 2 0]:       0       0       0       0       1       1 
[0 2 1]:       0       0       0       0       0       0 
[0 2 2]:       0       0       0       1       0       0 
[1 0 0]:       0       1       0       0       1       0 
[1 0 1]:       0       1       0       0       1       0 
[1 0 2]:       0       0       1       1       0       0 
[1 1 0]:       0       1       1       0       1       1 
[1 1 1]:       0       0       0       0       0       0 
[1 1 2]:       1       0       0       1       0       0 
[1 2 0]:       0       0       0       0       0       1 
[1 2 1]:       0       0       0       0       1       0 
[1 2 2]:       0       0       0       1       0       0 
[2 0 0]:       1       0       1       0       0       1 
[2 0 1]:       0       0       0       0       1       0 
[2 0 2]:       0       1       0       1       0       0 
[2 1 0]:       0       1       1       0       1       1 
[2 1 1]:       0       0       0       0       0       0 
[2 1 2]:       1       0       0       1       0       0 
[2 2 0]:       0       0       0       0       0       1 
[2 2 1]:       0       0       0       0       1       0 
[2 2 2]:       0       0       0       1       0       0 
[3 0 0]:       0       0       0       0       0       1 
[3 0 1]:       1       0       1       0       1       0 
[3 0 2]:       0       1       0       1       0       0 
[3 1 0]:       0       1       1       0       1       1 
[3 1 1]:       0       0       0       0       0       0 
[3 1 2]:       1       0       0       1       0       0 
[3 2 0]:       0       0       0       0       1       0 
[3 2 1]:       0       0       0       0       1       0 
[3 2 2]:       0       0       0       1       0       0 
[4 0 0]:       0       0       0       0       1       1 
[4 0 1]:       1       0       0       0       0       0 
[4 0 2]:       1       0       1       1       0       0 
[4 1 0]:       0       1       1       0       1       1 
[4 1 1]:       0       0       0       0       0       0 
[4 1 2]:       1       0       0       1       0       0 
[4 2 0]:       0       0       0       0       1       0 
[4 2 1]:       0       0       0       0       1       0 
[4 2 2]:       0       0       0       1       0       0 
[5 0 0]:       0       1       1       0       1       1 
[5 0 1]:       0       0       0       0       0       0 
[5 0 2]:       1       0       0       1       0       0 
[5 1 0]:       0       1       1       0       1       1 
[5 1 1]:       0       0       0       0       0       0 
[5 1 2]:       1       0       0       1       0       0 
[5 2 0]:       0       0       0       0       1       1 
[5 2 1]:       0       0       0       0       0       0 
[5 2 2]:       0       0       0       1       0       0 
`
	// fmt.Println("Input Pats")
	// fmt.Println(dt.ColumnByName("Input").Shape.Sizes)
	// fmt.Println(dt.ColumnByName("Input").String())
	assert.Equal(t, []int{6, 3, 2, 3, 3}, dt.Column("Input").Shape().Sizes)
	assert.Equal(t, exip, dt.Column("Input").String())

	exop := `Tensor: [Row: 6, ySize: 3, xSize: 2, poolY: 3, poolX: 3]
[0 0 0]:       0       0       0       0       1       0 
[0 0 1]:       0       1       0       0       1       0 
[0 0 2]:       1       1       0       1       0       0 
[0 1 0]:       0       1       1       0       1       1 
[0 1 1]:       0       0       0       0       0       0 
[0 1 2]:       1       0       0       1       0       0 
[0 2 0]:       0       0       0       0       1       1 
[0 2 1]:       0       0       0       0       0       0 
[0 2 2]:       0       0       0       1       0       0 
[1 0 0]:       0       1       0       0       1       0 
[1 0 1]:       0       1       0       0       1       0 
[1 0 2]:       0       0       1       1       0       0 
[1 1 0]:       0       1       1       0       1       1 
[1 1 1]:       0       0       0       0       0       0 
[1 1 2]:       1       0       0       1       0       0 
[1 2 0]:       0       0       0       0       0       1 
[1 2 1]:       0       0       0       0       1       0 
[1 2 2]:       0       0       0       1       0       0 
[2 0 0]:       1       0       1       0       0       1 
[2 0 1]:       0       0       0       0       1       0 
[2 0 2]:       0       1       0       1       0       0 
[2 1 0]:       0       1       1       0       1       1 
[2 1 1]:       0       0       0       0       0       0 
[2 1 2]:       1       0       0       1       0       0 
[2 2 0]:       0       0       0       0       0       1 
[2 2 1]:       0       0       0       0       1       0 
[2 2 2]:       0       0       0       1       0       0 
[3 0 0]:       0       0       0       0       0       1 
[3 0 1]:       1       0       1       0       1       0 
[3 0 2]:       0       1       0       1       0       0 
[3 1 0]:       0       1       1       0       1       1 
[3 1 1]:       0       0       0       0       0       0 
[3 1 2]:       1       0       0       1       0       0 
[3 2 0]:       0       0       0       0       1       0 
[3 2 1]:       0       0       0       0       1       0 
[3 2 2]:       0       0       0       1       0       0 
[4 0 0]:       0       0       0       0       1       1 
[4 0 1]:       1       0       0       0       0       0 
[4 0 2]:       1       0       1       1       0       0 
[4 1 0]:       0       1       1       0       1       1 
[4 1 1]:       0       0       0       0       0       0 
[4 1 2]:       1       0       0       1       0       0 
[4 2 0]:       0       0       0       0       1       0 
[4 2 1]:       0       0       0       0       1       0 
[4 2 2]:       0       0       0       1       0       0 
[5 0 0]:       0       1       1       0       1       1 
[5 0 1]:       0       0       0       0       0       0 
[5 0 2]:       1       0       0       1       0       0 
[5 1 0]:       0       1       1       0       1       1 
[5 1 1]:       0       0       0       0       0       0 
[5 1 2]:       1       0       0       1       0       0 
[5 2 0]:       0       0       0       0       1       1 
[5 2 1]:       0       0       0       0       0       0 
[5 2 2]:       0       0       0       1       0       0 
`

	// fmt.Println("ECout Pats")
	// fmt.Println(dt.ColumnByName("ECout").Shape.Sizes)
	// fmt.Println(dt.ColumnByName("ECout").String())

	assert.Equal(t, []int{6, 3, 2, 3, 3}, dt.Column("ECout").Shape().Sizes)
	assert.Equal(t, exop, dt.Column("ECout").String())
}
