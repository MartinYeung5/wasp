package sc

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	waspapi "github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/spf13/viper"
)

type Config struct {
	ShortName   string
	Name        string
	ProgramHash string

	bootupData *registry.BootupData
}

func (c *Config) Alias() string {
	if config.SCAlias != "" {
		return config.SCAlias
	}
	if c.ShortName != "" {
		return c.ShortName
	}
	panic("Which smart contract? (--sc=<alias> is required)")
}

func (c *Config) Href() string {
	return "/" + c.ShortName
}

var DefaultCommittee = []int{0, 1, 2, 3}

func (c *Config) SetCommittee(indexes []int) {
	config.Set("sc."+c.Alias()+".committee", indexes)
}

func (c *Config) Committee() []int {
	r := viper.GetIntSlice("sc." + c.Alias() + ".committee")
	if len(r) > 0 {
		return r
	}
	return DefaultCommittee
}

func (c *Config) SetQuorum(n uint16) {
	config.Set("sc."+c.Alias()+".quorum", int(n))
}

func (c *Config) Quorum() uint16 {
	q := viper.GetInt("sc." + c.Alias() + ".quorum")
	if q != 0 {
		return uint16(q)
	}
	return uint16(3)
}

func (c *Config) PrintUsage(s string) {
	fmt.Printf("Usage: %s %s %s\n", os.Args[0], c.ShortName, s)
}

func (c *Config) HandleSetCmd(args []string) {
	if len(args) != 2 {
		c.PrintUsage("set <key> <value>")
		os.Exit(1)
	}
	config.Set("sc."+c.Alias()+"."+args[0], args[1])
}

func (c *Config) usage(commands map[string]func([]string)) {
	cmdNames := make([]string, 0)
	for k := range commands {
		cmdNames = append(cmdNames, k)
	}

	c.PrintUsage(fmt.Sprintf("[options] [%s]", strings.Join(cmdNames, "|")))
	os.Exit(1)
}

func (c *Config) HandleCmd(args []string, commands map[string]func([]string)) {
	if len(args) < 1 {
		c.usage(commands)
	}
	cmd, ok := commands[args[0]]
	if !ok {
		c.usage(commands)
	}
	cmd(args[1:])
}

func (c *Config) SetAddress(address string) {
	config.SetSCAddress(c.Alias(), address)
}

func (c *Config) Address() *address.Address {
	return config.GetSCAddress(c.Alias())
}

func (c *Config) IsAvailable() bool {
	return config.TrySCAddress(c.Alias()) != nil
}

func (c *Config) Deploy(sigScheme signaturescheme.SignatureScheme) error {
	scAddress, err := Deploy(&DeployParams{
		Quorum:      c.Quorum(),
		Committee:   c.Committee(),
		Description: c.Name,
		ProgramHash: c.ProgramHash,
		SigScheme:   sigScheme,
	})
	if err == nil {
		c.SetAddress(scAddress.String())
		c.SetCommittee(c.Committee())
		c.SetQuorum(c.Quorum())
	}
	return err
}

type DeployParams struct {
	Quorum      uint16
	Committee   []int
	Description string
	ProgramHash string
	SigScheme   signaturescheme.SignatureScheme
}

func Deploy(params *DeployParams) (*address.Address, error) {
	scAddress, _, err := waspapi.CreateSC(waspapi.CreateSCParams{
		Node:                  config.GoshimmerClient(),
		CommitteeApiHosts:     config.CommitteeApi(params.Committee),
		CommitteePeeringHosts: config.CommitteePeering(params.Committee),
		AccessNodes:           []string{},
		N:                     uint16(len(params.Committee)),
		T:                     uint16(params.Quorum),
		OwnerSigScheme:        params.SigScheme,
		ProgramHash:           params.progHash(),
		Description:           params.Description,
		Textout:               os.Stdout,
		Prefix:                "[deploy] ",
	})
	if err != nil {
		return nil, err
	}
	err = waspapi.ActivateSCMulti(waspapi.ActivateSCParams{
		Addresses:         []*address.Address{scAddress},
		ApiHosts:          config.CommitteeApi(params.Committee),
		PublisherHosts:    config.CommitteeNanomsg(params.Committee),
		WaitForCompletion: config.WaitForConfirmation,
		Timeout:           30 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("Initialized %s smart contract\n", params.Description)
	fmt.Printf("SC Address: %s\n", scAddress)

	if config.SCAlias != "" {
		c := Config{
			ProgramHash: params.ProgramHash,
		}
		c.SetAddress(scAddress.String())
		c.SetCommittee(params.Committee)
		c.SetQuorum(params.Quorum)
	}

	return scAddress, nil
}

func (p *DeployParams) progHash() hashing.HashValue {
	hash, err := hashing.HashValueFromBase58(p.ProgramHash)
	if err != nil {
		panic(err)
	}
	return hash
}

func (c *Config) BootupData() *registry.BootupData {
	if c.bootupData != nil {
		return c.bootupData
	}
	d, exists, err := waspapi.GetSCData(config.WaspApi(), c.Address())
	if err != nil || !exists {
		panic(fmt.Sprintf("GetSCData host = %s, addr = %s exists = %v err = %v\n",
			config.WaspApi(), c.Address(), exists, err))
	}
	c.bootupData = d
	return c.bootupData
}
