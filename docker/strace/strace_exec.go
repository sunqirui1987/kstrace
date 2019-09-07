package strace

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/fntlnz/mountinfo"
	"github.com/spf13/cobra"
)

type StraceExec struct {
	podUID           string
	containerName    string
	programPath      string
	bpftraceExecPath string
}

func NewStraceExec() *StraceExec {
	return &StraceExec{}
}

func NewStraceExecCommand() *cobra.Command {
	strace := NewStraceExec()
	cmd := &cobra.Command{
		RunE: func(c *cobra.Command, args []string) error {
			if err := strace.Validate(c, args); err != nil {
				return err
			}

			if err := strace.Complete(c, args); err != nil {
				return err
			}

			if err := strace.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&strace.containerName, "container", "c", strace.containerName, "Specify  container")
	cmd.Flags().StringVarP(&strace.podUID, "poduid", "p", strace.podUID, "Specify  pod ")
	cmd.Flags().StringVarP(&strace.programPath, "file", "f", "/program/root.bt", "Specify  bpftrace path")
	cmd.Flags().StringVarP(&strace.bpftraceExecPath, "bpftracebinary", "b", "/bin/bpftrace", "Specify  bpftrace exec path")
	return cmd
}

func (o *StraceExec) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

// Complete completes the setup of the command.
func (o *StraceExec) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *StraceExec) Run() error {
	programPath := o.programPath
	pid, err := findPidByPodContainer(o.podUID, o.containerName)
	if err != nil {
		return err
	}
	if pid == nil {
		return fmt.Errorf("pid not found")
	}
	if len(*pid) == 0 {
		return fmt.Errorf("invalid pid found")
	}

	f, err := ioutil.ReadFile(programPath)
	if err != nil {
		return err
	}
	programPath = path.Join(os.TempDir(), "program-container.bt")
	r := strings.Replace(string(f), "$container_pid", *pid, -1)
	if err := ioutil.WriteFile(programPath, []byte(r), 0755); err != nil {
		return err
	}

	fmt.Println("if your program has maps to print, send a SIGINT using Ctrl-C, if you want to interrupt the execution send SIGINT two times")
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Signal(syscall.SIGINT))

	go func() {
		killable := false
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				if !killable {
					killable = true
					fmt.Println("\nfirst SIGINT received, now if your program had maps and did not free them it should print them out")
					continue
				}
				return
			}
		}
	}()

	c := exec.CommandContext(ctx, o.bpftraceExecPath, programPath)
	c.Stdout = os.Stdout
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	return c.Run()
}

func findPidByPodContainer(podUID, containerName string) (*string, error) {
	d, err := os.Open("/proc")

	if err != nil {
		return nil, err
	}

	defer d.Close()

	for {
		dirs, err := d.Readdir(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, di := range dirs {
			if !di.IsDir() {
				continue
			}
			dname := di.Name()
			if dname[0] < '0' || dname[0] > '9' {
				continue
			}

			mi, err := mountinfo.GetMountInfo(path.Join("/proc", dname, "mountinfo"))
			if err != nil {
				continue
			}

			for _, m := range mi {
				root := m.Root
				if strings.Contains(root, podUID) && strings.Contains(root, containerName) {
					return &dname, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no process found for specified pod and container")
}
