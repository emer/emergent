package patgen

import (
	"fmt"
	"testing"

	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

func TestVocab(t *testing.T) {
	fmt.Println("Say hi")
	m := make(map[string]*etensor.Float32)
	AddVocabVoid(m, 3, 3, 3, "void")
	AddVocab(m, 3, 3, 3, 0.3, "A")
	AddVocabDrift(m, 3, 3, 3, 0.3, 0.2, "B") // nOn=4*(3*3*0.3); nDrift=nOn*0.5
	AddVocabRepeat(m, 3, 3, 3, 0.3, "ctxt1")
	VocabConcat(m, "AB-C", []string{"A", "B"})
	VocabSlice(m, "AB-C", []string{"A'", "B'"}, []int{0, 3, 6}) // 3 cutoffs for 2 vocabs
	VocabShuffle(m, []string{"B'"})
	VocabClone(m, "B'", "B''")

	fmt.Println("\n\n\nvoid")
	fmt.Println(m["void"].T())

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
	ConfigPats(dt, m, "Input", []string{"A", "B", "ctxt1", "ctxt1", "void", "B'"})

	fmt.Println("\n\n\nPats'")
	fmt.Println(dt.ColByName("Input").Shapes())
	fmt.Println(dt.ColByName("Input").T())
}

// // example code
// patgen.AddVocabVoid(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, "void")
// patgen.AddVocab(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "A")
// patgen.AddVocab(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "B")
// patgen.AddVocab(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "lA")
// patgen.AddVocab(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "lB")
// patgen.VocabClone(ss.PoolVocab, "B", "C")
// patgen.VocabDrift(ss.PoolVocab, 1, "C") // B --drift--> C
// patgen.AddVocabRepeat(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "ctxt1")
// patgen.AddVocabRepeat(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "ctxt2")
// patgen.AddVocabRepeat(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "ctxt3")
// patgen.AddVocabRepeat(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "ctxt4")
// patgen.AddVocabRepeat(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "ctxt5")
// patgen.AddVocabRepeat(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "ctxt6")
// patgen.AddVocabRepeat(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "ctxt7")
// patgen.AddVocabRepeat(ss.PoolVocab, 10, ss.ECPool.Y, ss.ECPool.X, ss.ECPctAct, "ctxt8")

// patgen.InitPats(ss.TrainAB, "TrainAB", "TrainAB Pats", "Input", "ECout", 10, ss.YSize, ss.XSize, ss.ECPool.Y, ss.ECPool.X)
// patgen.ConfigPats(ss.TrainAB, ss.PoolVocab, "Input", []string{"A", "B", "ctxt1", "ctxt2", "ctxt3", "ctxt4"})
// patgen.ConfigPats(ss.TrainAB, ss.PoolVocab, "ECout", []string{"A", "B", "ctxt1", "ctxt2", "ctxt3", "ctxt4"})

// patgen.InitPats(ss.TestAB, "TestAB", "TestAB Pats", "Input", "ECout", 10, ss.YSize, ss.XSize, ss.ECPool.Y, ss.ECPool.X)
// patgen.ConfigPats(ss.TestAB, ss.PoolVocab, "Input", []string{"A", "void", "ctxt1", "ctxt2", "ctxt3", "ctxt4"})
// patgen.ConfigPats(ss.TestAB, ss.PoolVocab, "ECout", []string{"A", "B", "ctxt1", "ctxt2", "ctxt3", "ctxt4"})

// patgen.InitPats(ss.TrainAC, "TrainAC", "TrainAC Pats", "Input", "ECout", 10, ss.YSize, ss.XSize, ss.ECPool.Y, ss.ECPool.X)
// patgen.ConfigPats(ss.TrainAC, ss.PoolVocab, "Input", []string{"A", "C", "ctxt5", "ctxt6", "ctxt7", "ctxt8"})
// patgen.ConfigPats(ss.TrainAC, ss.PoolVocab, "ECout", []string{"A", "C", "ctxt5", "ctxt6", "ctxt7", "ctxt8"})

// patgen.InitPats(ss.TestAC, "TestAC", "TestAC Pats", "Input", "ECout", 10, ss.YSize, ss.XSize, ss.ECPool.Y, ss.ECPool.X)
// patgen.ConfigPats(ss.TestAC, ss.PoolVocab, "Input", []string{"A", "void", "ctxt5", "ctxt6", "ctxt7", "ctxt8"})
// patgen.ConfigPats(ss.TestAC, ss.PoolVocab, "ECout", []string{"A", "C", "ctxt5", "ctxt6", "ctxt7", "ctxt8"})

// patgen.InitPats(ss.TestLure, "TestLure", "TestLure Pats", "Input", "ECout", 10, ss.YSize, ss.XSize, ss.ECPool.Y, ss.ECPool.X)
// patgen.ConfigPats(ss.TestLure, ss.PoolVocab, "Input", []string{"lA", "void", "ctxt1", "ctxt2", "ctxt3", "ctxt4"}) // arbitrary ctxt here
// patgen.ConfigPats(ss.TestLure, ss.PoolVocab, "ECout", []string{"lA", "lB", "ctxt1", "ctxt2", "ctxt3", "ctxt4"})   // arbitrary ctxt here
