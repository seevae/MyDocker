package container

import (
	log "github.com/Sirupsen/logrus"
	"syscall"
	"os/exec"
	"os"
)

func NewParentProcess(tty bool)(*exec.Cmd, *os.File){
	readPipe,writePipe,err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v",err)
		return nil,nil
	}
	cmd := exec.Command("/proc/self/exe","init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	mntURL := "/root/mnt/"
	rootURL := "/root/"
	NewWorkSpace(rootURL, mntURL)
	cmd.Dir = mntURL
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil,nil,err
	}
	return read, write, nil
}

//Create a AUFS filesystem as container root workspace
func NewWorkSpace(rootURL string,mntURL string){
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL,mntURL)
}

//将busybox.tar解压到busybox目录下,作为容器的只读层 
func CreateReadOnlyLayer(rootURL string){
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil{
		log.Info("Fail to judge whether dir %s exists. %v",busyboxURL,err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL,0777);err != nil {
			log.Errorf("Mkdir dir %s error. %v",busyboxURL,err)
		}
		if _, err := exec.Command("tar","-xvf",busyboxTarURL,"-C",busyboxURL).CombinedOutput();err != nil{
			log.Errorf("unTar dir %s error %v",busyboxTarURL,err)
		}
	}
}

//创建一个名为writeLayer的文件夹作为容器唯一的可写层 
func CreateWriteLayer(rootURL string){
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL,0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v",writeURL, err)
	}
}

//创建mnt文件夹作为挂载点,然后将writeLayer目录和busybox目录mount到mnt目录下
func CreateMountPoint(rootURL string, mntURL string){
	//创建mnt文件夹作为挂载点
	if err := os.Mkdir(mntURL,0777); err != nil {
		log.Errorf("Mkdir dir %s error . %v",mntURL, err)
	}
	//把writeLayer目录和busybox目录mount到mnt目录下
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount","-t","aufs","-o",dirs,"none",mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run();err!=nil{
		log.Errorf("%v",err)
	}
}

//判断路径是否存在的函数
func PathExists(path string)(bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootURL string, mntURL string) {
	DeleteMountPoint(rootURL, mntURL)
	DeleteWriteLayer(rootURL)
}

func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount",mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v",err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove dir %s error %v", writeURL, err)
	}
}
