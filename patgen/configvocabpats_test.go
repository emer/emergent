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
	AddVocabPermutedBinary(m, "A", 3, 3, 3, 0.3)
	AddVocabDrift(m, "B", 3, 0.2, "A", 0) // nOn=4*(3*3*0.3); nDrift=nOn*0.5
	AddVocabRepeat(m, "ctxt1", 3, "A", 0)
	VocabConcat(m, "AB-C", []string{"A", "B"})
	VocabSlice(m, "AB-C", []string{"A'", "B'"}, []int{0, 3, 6}) // 3 cutoffs for 2 vocabs
	VocabShuffle(m, []string{"B'"})
	AddVocabClone(m, "B''", "B'")

	fmt.Println("\n\n\nmap")
	fmt.Println(reflect.ValueOf(m).MapKeys())

	fmt.Println("\n\n\nempty")
	fmt.Println(m["empty"].T())

	fmt.Println("\n\n\nA")
	fmt.Println(m["A"].T())

	fmt.Println("\n\n\nB")
	fmt.Println(m["B"].T())

	fmt.Println("\n\n\nctxt1")
	fmt.Println(m["ctxt1"].T())

	fmt.Println("\n\n\nAB-C")
	fmt.Println(m["AB-C"].T())

	fmt.Println("\n\n\nA'")
	fmt.Println(m["A'"].T())

	fmt.Println("\n\n\nB'")
	fmt.Println(m["B'"].T())

	fmt.Println("\n\n\nB''")
	fmt.Println(m["B''"].T())

	// config pats
	dt := etable.NewTable("TrainAB")
	InitPats(dt, "TrainAB", "describe", "Input", "ECout", 3, 3, 2, 3, 3)
	MixPats(dt, m, "Input", []string{"A", "B", "ctxt1", "ctxt1", "empty", "B'"})

	fmt.Println("\n\n\nPats'")
	fmt.Println(dt.ColByName("Input").Shapes())
	fmt.Println(dt.ColByName("Input").T())
}
