package patgen

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

func TestVocab(t *testing.T) {
	fmt.Println("Testing starts")
	m := make(map[string]*etensor.Float32)
	AddVocabEmpty(m, "empty", 3, 3, 3)
	AddVocabPermutedBinary(m, "A", 3, 3, 3, 0.3, 0.5)
	AddVocabDrift(m, "B", 3, 0.2, "A", 0) // nOn=4*(3*3*0.3); nDrift=nOn*0.5
	AddVocabRepeat(m, "ctxt1", 3, "A", 0)
	VocabConcat(m, "AB-C", []string{"A", "B"})
	VocabSlice(m, "AB-C", []string{"A'", "B'"}, []int{0, 3, 6}) // 3 cutoffs for 2 vocabs
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
	InitPats(dt, "TrainAB", "describe", "Input", "ECout", 3, 3, 2, 3, 3)
	MixPats(dt, m, "Input", []string{"A", "B", "ctxt1", "ctxt1", "empty", "B'"})

	fmt.Println("Pats'")
	fmt.Println(dt.ColByName("Input").Shapes())
	fmt.Println(dt.ColByName("Input").T())
}
