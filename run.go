package main

import(
	"github.com/seevae/mydocker/container"
	"github.com/seevae/mydocker/cgroups/subsystems"
	"github.com/seevae/mydocker/cgroups"
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"
)

func Run(tty bool,comArray []string,res *subsystems.ResourceConfig){
	parent,writePipe:=container.NewParentProcess(tty)
	if parent == nil{
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start();err != nil{
		log.Error(err)
	}

	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	sendInitCommand(comArray,writePipe)
	parent.Wait()
}

func sendInitCommand(comArray []string, writePipe *os.File){
	command := strings.Join(comArray," ")
	log.Infof("command all is %s",command)
	writePipe.WriteString(command)
	writePipe.Close()
}
