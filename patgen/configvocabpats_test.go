package patgen

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/emer/etable/etable"
)

func TestVocab(t *testing.T) {
	fmt.Println("Testing starts")
	m := make(Vocab)
	AddVocabEmpty(m, "empty", 6, 3, 3)
	AddVocabPermutedBinary(m, "A", 6, 3, 3, 0.3, 0.5)
	AddVocabDrift(m, "B", 6, 0.2, "A", 0) // nOn=4*(3*3*0.3); nDrift=nOn*0.5
	AddVocabRepeat(m, "ctxt1", 6, "A", 0)
	VocabConcat(m, "AB-C", []string{"A", "B"})
	VocabSlice(m, "AB-C", []string{"A'", "B'"}, []int{0, 6, 12}) // 3 cutoffs for 2 vocabs
	VocabShuffle(m, []string{"B'"})
	AddVocabClone(m, "B''", "B'")

	fmt.Println("map")
	fmt.Println(reflect.ValueOf(m).MapKeys())

	fmt.Println("empty")
	fmt.Println(m["empty"].String())

	fmt.Println("A")
	fmt.Println(m["A"].String())

	fmt.Println("B")
	fmt.Println(m["B"].String())

	fmt.Println("ctxt1")
	fmt.Println(m["ctxt1"].String())

	fmt.Println("AB-C")
	fmt.Println(m["AB-C"].String())

	fmt.Println("A'")
	fmt.Println(m["A'"].String())

	fmt.Println("B'")
	fmt.Println(m["B'"].String())

	fmt.Println("B''")
	fmt.Println(m["B''"].String())

	// config pats
	dt := etable.NewTable("TrainAB")
	InitPats(dt, "TrainAB", "describe", "Input", "ECout", 6, 3, 2, 3, 3)
	MixPats(dt, m, "Input", []string{"A", "B", "ctxt1", "ctxt1", "empty", "B'"})
	MixPats(dt, m, "ECout", []string{"A", "B", "ctxt1", "ctxt1", "empty", "B'"})

	// try shuffle
	Shuffle(dt, []int{0, 1, 2, 3, 4, 5}, []string{"Input", "ECout"}, false)

	fmt.Println("Input Pats")
	fmt.Println(dt.ColByName("Input").Shapes())
	fmt.Println(dt.ColByName("Input").T())

	fmt.Println("ECout Pats")
	fmt.Println(dt.ColByName("ECout").Shapes())
	fmt.Println(dt.ColByName("ECout").T())
}
