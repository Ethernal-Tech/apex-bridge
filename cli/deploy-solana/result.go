package clideploysolana

type deployProgramResult struct {
	Output string
}

func (r deployProgramResult) GetOutput() string {
	return r.Output
}
