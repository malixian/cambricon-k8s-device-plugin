package main


//不同模型在不同的剩余资源下所使用的策略

type NpuCoreAllocat interface {
	SetAlgoName(string string)
	GetAlgoName() string
	GetUseCore() int
	GetModelParallel() int
	GetDataParallel() int
}


type Qlearning struct {
	AlgoName	string
	Mp          int
	Dp          int
}

func(ql *Qlearning) GetAlgoName() string {
	return ql.AlgoName
}

func(ql *Qlearning) SetAlgoName(algoName string) {
	ql.AlgoName = algoName
}

func(ql *Qlearning) GetModelParallel() int {
	return ql.Mp
}

func(ql *Qlearning) GetDataParallel() int {
	return ql.Dp
}

// todo 是否可以将使用的核数更新至etcd
func(ql *Qlearning) GetUseCore() int {
	return ql.Dp * ql.Mp
}