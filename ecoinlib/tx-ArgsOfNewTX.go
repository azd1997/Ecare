package ecoin

/*********************************************************************************************************************
                                                    ArgsOfNewTX接口
*********************************************************************************************************************/

// ArgsOfNewTX 新建交易时传入的参数结构体的接口。这样子做可以省掉上一版本中ParseArgs的步骤
type ArgsOfNewTX interface {
	CheckArgsValue(gsm *GlobalStateMachine) (err error)
}

//// BaseArgs 基本的参数项，这里其实是放了gsm *GlobalStateMachine进去
//type BaseArgs struct {
//	Gsm *GlobalStateMachine
//}