package log

var exitFunctions []func() error

// AddExitFunctions 添加退出函数，将会在监测到退出信号时执行，执行顺序和添加顺序相同
func AddExitFunctions(fn ...func() error) {
	exitFunctions = append(exitFunctions, fn...)
}
