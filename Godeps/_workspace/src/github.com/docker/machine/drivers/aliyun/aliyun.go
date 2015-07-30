package aliyun

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"io/ioutil"

	"strings"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
	"github.com/jiangshengwu/aliyun-sdk-for-go/ecs"
	"github.com/jiangshengwu/aliyun-sdk-for-go/util"
)

type Driver struct {
	AccessKeyId      string
	AccessKeySecret  string
	RegionId         string
	SecurityGroupId  string
	ImageId          string
	InstanceTypeId   string
	InstanceId       string
	PrivateIPAddress string
	SSHPass          string
	PublicKey        []byte
	DockerArgs       string
	InsecureRegistry string
	BandwidthOut     string
	MachineName      string
	IPAddress        string
	SSHKey           string
	SSHUser          string
	SSHPort          int
	CaCertPath       string
	PrivateKeyPath   string
	DriverKeyPath    string
	SwarmMaster      bool
	SwarmHost        string
	SwarmDiscovery   string
	storePath        string
}

const (
	defaultTimeout      = 1 * time.Second
	maxTry              = 10
	groupPrefix         = "docker-machine-"
	pageSize            = "50"
	groupNotFoundErr    = "InvalidSecurityGroupId.NotFound"
	defaultImage        = "ubuntu1404_64_20G_aliaegis_20150325.vhd"
	defaultInstanceType = "ecs.t1.small"
)

func init() {
	drivers.Register("aliyun", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "ALIYUN_ACCESS_KEY_ID",
			Name:   "aliyun-access-key-id",
			Usage:  "Aliyun Access Key Id",
		},
		cli.StringFlag{
			EnvVar: "ALIYUN_ACCESS_KEY_SECRET",
			Name:   "aliyun-access-key-secret",
			Usage:  "Aliyun Access Key Secret",
		},
		cli.StringFlag{
			Name:  "aliyun-region-id",
			Usage: "Aliyun Region Id",
		},
		cli.StringFlag{
			Name:  "aliyun-security-group-id",
			Usage: "Aliyun Security Group Id",
		},
		cli.StringFlag{
			Name:  "aliyun-image-id",
			Usage: "Aliyun Image Id",
		},
		cli.StringFlag{
			Name:  "aliyun-instance-type-id",
			Usage: "Aliyun Instance Type Id",
		},
		cli.StringFlag{
			//TODO 以后改为随机生成密码
			Name:  "aliyun-ssh-pass",
			Usage: "Aliyun Instance SSH Password",
			Value: "ASDqwe123",
		},
		cli.StringFlag{
			Name:  "aliyun-bandwidth-out",
			Usage: "Aliyun Internet Bandwidth Out",
			Value: "1",
		},
		cli.StringFlag{
			//TODO CentOS6支持
			Name:  "aliyun-docker-args",
			Usage: "Aliyun Docker Daemon Arguments",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{
		MachineName:    machineName,
		storePath:      storePath,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
	}, nil
}

func (d *Driver) DriverName() string {
	return "aliyun"
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "root"
	}

	return d.SSHUser
}

func (d *Driver) GetSSHPassword() string {
	return d.SSHPass
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessKeyId = flags.String("aliyun-access-key-id")
	d.AccessKeySecret = flags.String("aliyun-access-key-secret")
	d.RegionId = flags.String("aliyun-region-id")
	d.SecurityGroupId = flags.String("aliyun-security-group-id")
	d.ImageId = flags.String("aliyun-image-id")
	d.InstanceTypeId = flags.String("aliyun-instance-type-id")
	d.SSHPass = flags.String("aliyun-ssh-pass")
	d.BandwidthOut = flags.String("aliyun-bandwidth-out")
	d.DockerArgs = flags.String("aliyun-docker-args")
	d.InsecureRegistry = flags.String("aliyun-docker-registry")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "root"
	d.SSHPort = 22

	if d.AccessKeyId == "" {
		return fmt.Errorf("aliyun driver requires the --aliyun-access-key-id option")
	}
	if d.AccessKeySecret == "" {
		return fmt.Errorf("aliyun driver requires the --aliyun-access-key-secret option")
	}
	if d.RegionId == "" {
		return fmt.Errorf("aliyun driver requires the --aliyun-region-id option")
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	regionValid := false
	imageValid := false
	typeValid := false
	cli := d.getClient()
	// Check if region valid
	regionResp, regionErr := cli.Region.DescribeRegions(nil)
	if regionErr != nil {
		return regionErr
	}
	regions := regionResp.AllRegions.AllRegion
	for _, v := range regions {
		if v.RegionId == d.RegionId {
			regionValid = true
			log.Debugf("region id is valid: %s", d.RegionId)
			break
		}
	}
	if !regionValid {
		return fmt.Errorf("region id is not valid: %s", d.RegionId)
	}

	// Check if image id valid
	if d.ImageId != "" {
		imageResp, imageErr := cli.Image.DescribeImages(map[string]string{
			"RegionId": d.RegionId,
			"PageSize": pageSize,
		})
		if imageErr != nil {
			return imageErr
		}
		images := imageResp.AllImages.AllImage
		for _, v := range images {
			if v.ImageId == d.ImageId {
				imageValid = true
				log.Debugf("image id is valid: %s", d.ImageId)
				break
			}
		}
		if !imageValid {
			return fmt.Errorf("image id is not valid: %s", d.ImageId)
		}
	} else {
		d.ImageId = defaultImage
	}

	// Check if instance type valid
	if d.InstanceTypeId != "" {
		typeResp, typeErr := cli.Other.DescribeInstanceTypes(nil)
		if typeErr != nil {
			return typeErr
		}
		types := typeResp.AllInstanceTypes.AllInstanceType
		for _, v := range types {
			if v.InstanceTypeId == d.InstanceTypeId {
				typeValid = true
				log.Debugf("instance type id is valid: %s", d.InstanceTypeId)
				break
			}
		}
		if !typeValid {
			return fmt.Errorf("instance type id is not valid: %s", d.InstanceTypeId)
		}
	} else {
		d.InstanceTypeId = defaultInstanceType
	}

	if d.SecurityGroupId == "" {
		// Create new security group
		groupResp, err := cli.SecurityGroup.CreateSecurityGroup(map[string]string{
			"RegionId":          d.RegionId,
			"SecurityGroupName": groupPrefix + getGuid(),
		})
		if err != nil {
			return err
		}
		d.SecurityGroupId = groupResp.SecurityGroupId

		// Make security group accessible from network
		_, err = cli.SecurityGroup.AuthorizeSecurityGroup(map[string]string{
			"RegionId":        d.RegionId,
			"SecurityGroupId": d.SecurityGroupId,
			"IpProtocol":      "all",
			"PortRange":       "-1/-1",
			"SourceCidrIp":    "0.0.0.0/0",
		})
		if err != nil {
			return err
		}
	} else {
		// Check if security group id valid
		_, groupErr := cli.SecurityGroup.DescribeSecurityGroupAttribute(map[string]string{
			"RegionId":        d.RegionId,
			"SecurityGroupId": d.SecurityGroupId,
		})
		if groupErr != nil {
			if e, ok := groupErr.(*util.SdkError); ok {
				if e.Resp.Code == groupNotFoundErr {
					return fmt.Errorf("security group id is not valid: %s", d.SecurityGroupId)
				}
			}
			return groupErr
		}
		log.Debugf("security group id is valid: %s", d.SecurityGroupId)
	}

	return nil
}

func (d *Driver) Create() error {
	// Create SSH key
	if err := d.createSSHKey(); err != nil {
		return err
	}

	// Create instance
	log.Info("Creating ECS instance...")
	createResp, err := d.getClient().Instance.CreateInstance(map[string]string{
		"RegionId":                d.RegionId,
		"SecurityGroupId":         d.SecurityGroupId,
		"ImageId":                 d.ImageId,
		"InstanceType":            d.InstanceTypeId,
		"InstanceName":            d.MachineName,
		"InternetMaxBandwidthOut": d.BandwidthOut,
		"Password":                d.SSHPass,
	})
	if err != nil {
		return err
	}
	d.InstanceId = createResp.InstanceId

	// Allocate public ip address
	log.Info("Allocating public IP address...")
	ipResp, err := d.getClient().Network.AllocatePublicIpAddress(map[string]string{
		"InstanceId": d.InstanceId,
	})
	if err != nil {
		return nil
	}
	d.IPAddress = ipResp.IpAddress

	// Start instance
	log.Info("Starting instance, this may take several minutes...")
	_, err = d.getClient().Instance.StartInstance(map[string]string{
		"InstanceId": d.InstanceId,
	})
	if err != nil {
		return err
	}

	// Get private IP address
	statusResp, err := d.getClient().Instance.DescribeInstanceAttribute(map[string]string{
		"InstanceId": d.InstanceId,
	})
	if err != nil {
		return err
	}
	if ips := statusResp.InnerIpAddress.AllIpAddress; len(ips) > 0 {
		d.PrivateIPAddress = ips[0]
	}

	// Wait for instance to start
	if err := utils.WaitFor(d.waitToStartInstance); err != nil {
		return err
	}

	//TODO 暂时处理：等待20秒，主机真正启动
	time.Sleep(20 * time.Second)

	// Wait for ssh available
	port, e := d.GetSSHPort()
	if e != nil {
		return e
	}
	ip, e := d.GetIP()
	if e != nil {
		return e
	}
	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", ip, port)); err != nil {
		return err
	}

	// Upload SSH key to host
	log.Info("Upload SSH key to machine...")
	if err := d.uploadKeyPair(); err != nil {
		return err
	}

	log.Infof("Created Instance ID %s, Public IP address %s, Private IP address %s",
		d.InstanceId,
		d.IPAddress,
		d.PrivateIPAddress,
	)

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" {
		if d.PrivateIPAddress == "" {
			return "", fmt.Errorf("IP address is not set")
		}
		return d.PrivateIPAddress, nil
	}
	return d.IPAddress, nil
}

func (d *Driver) GetPrivateIP() (string, error) {
	if d.PrivateIPAddress == "" {
		return "", fmt.Errorf("Private IP address is not set")
	}
	return d.PrivateIPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	statusResp, err := d.getClient().Instance.DescribeInstanceAttribute(map[string]string{
		"InstanceId": d.InstanceId,
	})
	if err != nil {
		return state.Error, err
	}
	switch statusResp.Status {
	case "Running":
		return state.Running, nil
	case "Starting":
		return state.Starting, nil
	case "Stopped":
		return state.Stopped, nil
	case "Stopping":
		return state.Stopping, nil
	default:
		return state.None, nil
	}
}

func (d *Driver) Start() error {
	log.Info("Starting...")

	_, err := d.getClient().Instance.StartInstance(map[string]string{
		"InstanceId": d.InstanceId,
	})
	return err
}

func (d *Driver) Stop() error {
	log.Info("Stopping...")

	_, err := d.getClient().Instance.StopInstance(map[string]string{
		"InstanceId": d.InstanceId,
	})
	return err
}

func (d *Driver) Remove() error {
	log.Info("Deleting...")

	if d.InstanceId == "" {
		// Instance id is empty due to some errors while creating,
		log.Warn("InstanceId is empty, assuming it has already bean removed from aliyun.")
		return nil
	}

	// If instance is running, kill it
	if st, _ := d.GetState(); st == state.Running {
		err := d.Kill()
		if err != nil {
			return err
		}
	}
	// Wait for instance to stop
	utils.WaitForSpecific(d.waitToStopInstance, 10, 6*time.Second)

	_, err := d.getClient().Instance.DeleteInstance(map[string]string{
		"InstanceId": d.InstanceId,
	})
	return err
}

func (d *Driver) Restart() error {
	log.Info("Restarting...")

	_, err := d.getClient().Instance.RebootInstance(map[string]string{
		"InstanceId": d.InstanceId,
	})
	return err
}

func (d *Driver) Kill() error {
	log.Info("Killing...")

	_, err := d.getClient().Instance.StopInstance(map[string]string{
		"InstanceId": d.InstanceId,
		"ForceStop":  "true",
	})
	return err
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) getClient() *ecs.EcsClient {
	cli := ecs.NewClient(
		d.AccessKeyId,
		d.AccessKeySecret,
		"",
	)
	return cli
}

func (d *Driver) createSSHKey() error {
	log.Info("Creating SSH Key Pair...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}
	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}
	d.PublicKey = publicKey
	log.Debug(publicKey)

	return nil
}

// MD5 hash
func getMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// Generate random guid
func getGuid() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getMd5String(base64.URLEncoding.EncodeToString(b))
}

func (d *Driver) waitToStartInstance() bool {
	st, err := d.GetState()
	if err != nil {
		return false
	}
	return st == state.Running
}

func (d *Driver) waitToStopInstance() bool {
	st, err := d.GetState()
	if err != nil {
		return false
	}
	return st == state.Stopped
}

func (d *Driver) uploadKeyPair() error {
	auth := ssh.Auth{
		Passwords: []string{d.SSHPass},
	}
	port, err := d.GetSSHPort()
	if err != nil {
		return nil
	}

	ip, err := d.GetIP()
	if err != nil {
		return nil
	}
	// ssh.SetDefaultClient(ssh.Native)

	sshCli, err := ssh.NewClient(d.GetSSHUsername(), ip, port, &auth)
	if err != nil {
		return err
	}

	command := fmt.Sprintf("mkdir -p ~/.ssh; echo '%s' > ~/.ssh/authorized_keys", string(d.PublicKey))
	output, err := sshCli.Output(command)
	log.Debugf("upload command: %s", command)
	log.Debugf("upload public key with err, output: %v: %s", err, output)

	if err != nil {
		return err
	}

	if strings.Contains(d.ImageId, "centos6") {
		output, err = sshCli.Output("chmod 644 /etc/sudoers; sed -r -i 's/^(Defaults\\s+requiretty)$/#\\1/' /etc/sudoers; chmod 400 /etc/sudoers")
		log.Debugf("enable sudo without tty with err, output: %v: %s", err, output)

		output, err = sshCli.Output("sed -i '$a ulimit -c unlimited' /etc/profile; source /etc/profile")
		log.Debugf("enable core dump with err, output: %v: %s", err, output)

		output, err = sshCli.Output("mkdir -p /corefile; echo \"/corefile/core-%e-%p-%t\" > /proc/sys/kernel/core_pattern")
		log.Debugf("modify core dump pattern with err, output: %v: %s", err, output)
	} else {
		output, err = sshCli.Output("route -n| grep -e '^172\\..*$'|awk '{print \"route del -net \"$1\" netmask \"$3\" dev \"$8|\"/bin/bash\"}'")
		log.Debugf("delete route rule with err, output: %v: %s", err, output)

		output, err = sshCli.Output("sed -r -i 's/^(up route add \\-net 172\\..*)$/#\\1/' /etc/network/interfaces")
		log.Debugf("fix route in /etc/network/interfaces with err, output: %v: %s", err, output)
	}

	return nil
}
