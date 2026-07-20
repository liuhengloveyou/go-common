// +build !windows

package common

import (
	"fmt"
	"io/ioutil"
	"syscall"
)

func SingleInstane(pidfile string) {
	if _, e := LockPidFile(pidfile); e != nil {
		pid, _ := ioutil.ReadFile(pidfile)
		panic(fmt.Errorf("Already running: [%v]", string(pid)))
	}
}

func LockPidFile(pidfile string) (fd int, e error) {
	fd, e = syscall.Open(pidfile, syscall.O_CREAT|syscall.O_RDWR, 0777)
	if e != nil {
		return
	}

	e = syscall.Flock(fd, syscall.LOCK_NB|syscall.LOCK_EX)
	if e != nil {
		return
	}

	e = syscall.Ftruncate(fd, 0)
	if e != nil {
		return
	}

	_, e = syscall.Write(fd, []byte(fmt.Sprintf("%d", syscall.Getpid())))
	if e != nil {
		return
	}

	return
}


func UnLockFile(fd int) (e error) {
	if e = syscall.Flock(fd, syscall.LOCK_UN); e != nil {
		return
	}

	if e = syscall.Close(fd); e != nil {
		return
	}

	return
}

/*
func main() {
	SingleInstane("/tmp/pid")
	time.Sleep(100 * time.Second)
}
*/
