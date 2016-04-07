package global

import "github.com/root-gg/wigo/wigo/executor"

type ProbeConfig struct {
	Path string
	Name string
	Delay int
}

type Probe struct {
	Config *ProbeConfig
	Result *ProbeResult
	Executor *executor.ProbeExecutor
}

func (p *Probe) NewProbe(path string, name string, delay string){
	p = new(Probe)

	p.Config = new(ProbeConfig)
	p.Config.Path = path
	p.Config.Name = name
	p.Config.Delay = delay

	//p.Result = ???

	p.Executor = executor.NewProbeExecutor(path,delay)
}

func (p *Probe) NewProbeFromJson(bytes []byte) (err error){
	return
}