package context

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	dcluster "github.com/weibocom/dockerf/cluster"
	dmachine "github.com/weibocom/dockerf/machine"
	"strings"
	"sync"
)

func (ctx *ClusterContext) createUnmanagedSlaves(md dcluster.MachineDescription, alreadyExistedMachines []dmachine.MachineInfo) error {

	var wg sync.WaitGroup
	log.Debugf("Already started machine: %+v", alreadyExistedMachines)

	errs := []string{}
	for _, address := range md.UnmanagedIps {
		if ctx.isUnmanagedMachineAlreadyStarted(alreadyExistedMachines, address) {
			log.Debugf("Unmanaged machine %s has been already started... ", address)
			continue
		}
		wg.Add(1)
		go func(address string) {
			defer wg.Done()
			name, err := ctx.mProxy.CreateUnmanagedSlave(md, address)
			if err != nil {
				log.Errorf("Failed to Create machine of address '%s'. Error:%s\n", address, err.Error())
				errs = append(errs, err.Error())
			} else {
				log.Infof("Machine(%s) created and started, begin to init slave.\n", name)
				if err := ctx.initSlave(name, md); err != nil {
					log.Errorf("Failed to init slave '%s'\n", name)
					errs = append(errs, err.Error())
				}
			}
		}(address)
	}
	wg.Wait()

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "---"))
	}
	return nil
}

// func (ctx *ClusterContext) destroyUnmanagedSlaves(md dcluster.MachineDescription, alreadyExistedMachines []dmachine.MachineInfo) error {

// }

func (ctx *ClusterContext) isUnmanagedMachineAlreadyStarted(alreadyStartedMachines []dmachine.MachineInfo, ip string) bool {
	for _, machine := range alreadyStartedMachines {
		if strings.Contains(ip, machine.IP) {
			return true
		}
	}
	return false
}
