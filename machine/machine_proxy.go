package machine

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/codegangsta/cli"
	dutils "github.com/weibocom/dockerf/utils"
)

type MachineProxy struct {
	App *cli.App
}

func NewMachineProxy(name string) *MachineProxy {
	app := NewMachineApp(name)
	proxy := &MachineProxy{
		App: app,
	}
	return proxy
}

func (mp *MachineProxy) Run(args ...string) error {
	rArgs := []string{mp.App.Name}
	rArgs = append(rArgs, args...)
	return mp.App.Run(rArgs)
}

func (mp *MachineProxy) Destroy(names ...string) error {
	if len(names) == 0 {
		return nil
	}
	args := []string{
		"rm",
	}
	args = append(args, names...)
	return mp.Run(args...)
}

func (mp *MachineProxy) Create(name string, options ...string) error {
	args := []string{"create"}
	args = append(args, options...)
	args = append(args, name)
	fmt.Printf("Create machine. name:%s, args:%+v\n", name, args)
	return mp.Run(args...)
}

func (mp *MachineProxy) Start(names ...string) ([]string, error) {
	l := len(names)
	if l == 0 {
		return []string{}, nil
	}
	successMachineNames := []string{}
	errs := []string{}
	var wg sync.WaitGroup
	wg.Add(l)
	for _, name := range names {
		go func(nm string) {
			defer wg.Done()
			args := []string{"start", nm}
			err := mp.Run(args...)
			if err != nil {
				errs = append(errs, err.Error())
				fmt.Printf("Machine(%s) start failed:%s\n", name, err.Error())
			} else {
				successMachineNames = append(successMachineNames, nm)
				fmt.Printf("Machine(%s) Started.\n", nm)
			}
		}(name)
	}
	wg.Wait()
	var err error = nil
	if len(errs) > 0 {
		err = errors.New(strings.Join(errs, "---"))
	}
	return successMachineNames, err
}

func (mp *MachineProxy) ExecCmd(machine, command string) error {
	args := []string{
		"ssh",
		machine,
		command,
	}
	return mp.Run(args...)
}

func (mp *MachineProxy) IP(machine string) (string, error) {
	cmd := os.Args[0]
	args := []string{"machine", "ip", machine}
	data, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Printf("Failed to load ip for '%s'\n", machine)
		return "", err
	}
	ip := string(data)
	ip = strings.Replace(ip, "\n", "", -1)
	ip = strings.Replace(ip, "\r", "", -1)
	return ip, nil
}

func (mp *MachineProxy) IPs(nodes []string) ([]string, error) {
	cmd := os.Args[0]
	args := []string{"machine", "ip"}
	args = append(args, nodes...)
	data, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Printf("Failed to load ips for '%s'\n", nodes)
		return []string{}, err
	}
	ips := []string{}
	ipLines := strings.NewReader(string(data))
	br := bufio.NewReader(ipLines)
	for {
		if ip, _, err := br.ReadLine(); err == nil {
			ips = append(ips, string(ip))
		} else {
			break
		}
	}
	if len(ips) != len(nodes) {
		return []string{}, errors.New(fmt.Sprintf("Failed to load ips. nodes:%+v, ips:%+v", nodes, ips))
	}
	return ips, nil
}

func (mp *MachineProxy) Config(node string, cluster string) (string, error) {
	if cluster != "" && cluster != "swarm" {
		panic("Cluster is not supported by '" + cluster + "'")
	}
	cmd := os.Args[0]
	args := []string{"machine", "config"}
	if cluster != "" {
		args = append(args, "--swarm")
	}
	args = append(args, node)

	data, err := exec.Command(cmd, args...).Output()
	if err != nil {
		dutils.Errorf("Fail to get machine config, name: %s, error: %s", node, err.Error())
		return "", err
	}
	sr := strings.NewReader(string(data))
	br := bufio.NewReader(sr)
	for {
		if bline, _, err := br.ReadLine(); err == nil {
			line := string(bline)
			// TODO tell this line is a configuration info, other than logs
			if strings.HasPrefix(line, "--tlsverify") && strings.Contains(line, "--tlscacert") {
				return line, nil
			}
		} else {
			return "", errors.New("fail to get configuration info. try to exec shell '" + cmd + " " + strings.Join(args, " ") + "' to check: " + err.Error())
		}
	}
	return "", errors.New("Can not get configuration info. try to exec shell '" + cmd + " " + strings.Join(args, " ") + "' to check.")
}

func (mp *MachineProxy) List(filter func(mi *MachineInfo) bool) ([]MachineInfo, error) {
	mis := []MachineInfo{}
	cmd := os.Args[0]
	args := []string{"machine", "ls"}
	bytes, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Printf("Cannot list the machine info.\n")
		return mis, err
	}
	br := bufio.NewReaderSize(strings.NewReader(string(bytes)), 64*1024)
	validate := false
	for {
		bLine, _, err := br.ReadLine()
		if err == nil {
			line := string(bLine)
			if !validate {
				if strings.HasPrefix(line, "NAME") {
					validate = true
				}
				continue
			}
			fields := strings.Fields(line)
			fl := len(fields)
			idx := 0
			// fields to machine info
			// name
			mi := MachineInfo{}
			if fl > idx {
				mi.Name = fields[idx]
				idx++
			}
			// active
			if fl > idx {
				if fields[idx] == "*" {
					mi.Active = "*"
					idx++
				}
			}
			// driver
			if fl > idx {
				mi.Driver = fields[idx]
				idx++
			}
			// state
			if fl > idx {
				mi.State = fields[idx]
				idx++
			}
			// last fields
			idx = fl - 1
			if idx >= 0 {
				lastField := fields[idx]
				if lastField == "(master)" {
					idx--
				}
				mi.Master = fields[idx]
				idx--
			}
			if idx >= 0 { // idx may be -1 where there are some error machines in the cluster.
				// url
				if mi.State != fields[idx] {
					mi.URL = fields[idx]
					ip, _ := parseMachineIpFromUrl(mi.URL)
					mi.IP = ip
				}
			}
			if filter(&mi) {
				mis = append(mis, mi)
			}
		} else if err.Error() == "EOF" {
			break
		} else {
			fmt.Printf("Error to Read from machine 'ls' output: ", err.Error())
			return []MachineInfo{}, err
		}
	}
	return mis, nil
}

//解析machine ip，machine url格式：tcp://192.168.99.100:2376
func parseMachineIpFromUrl(machineUrl string) (string, error) {
	u, err := url.Parse(machineUrl)
	if err != nil {
		dutils.Error(fmt.Sprintf("Invalid machine url, url: %s, error: %s", machineUrl, err.Error()))
		return "", err
	}
	return strings.Split(u.Host, ":")[0], nil
}
