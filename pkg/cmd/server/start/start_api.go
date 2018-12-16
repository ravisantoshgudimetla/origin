package start

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"github.com/openshift/library-go/pkg/serviceability"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	configapi "github.com/openshift/origin/pkg/cmd/server/apis/config"
	"github.com/openshift/origin/pkg/cmd/server/start/options"
)

var apiLong = templates.LongDesc(`
	Start the master API

	This command starts the master API.  Running

	    %[1]s start master %[2]s

	will start the server listening for incoming API requests. The server
	will run in the foreground until you terminate the process.`)

// NewCommandStartMasterAPI starts only the APIserver
func NewCommandStartMasterAPI(name, basename string, out, errout io.Writer, stopCh <-chan struct{}) (*cobra.Command, *MasterOptions) {
	opts := &MasterOptions{Output: out}
	opts.DefaultsFromName(basename)

	cmd := &cobra.Command{
		Use:   name,
		Short: "Launch master API",
		Long:  fmt.Sprintf(apiLong, basename, name),
		Run: func(c *cobra.Command, args []string) {
			if err := opts.Complete(); err != nil {
				fmt.Fprintln(errout, kcmdutil.UsageErrorf(c, err.Error()))
				return
			}

			if len(opts.ConfigFile) == 0 {
				fmt.Fprintln(errout, kcmdutil.UsageErrorf(c, "--config is required for this command"))
				return
			}

			if err := opts.Validate(args); err != nil {
				fmt.Fprintln(errout, kcmdutil.UsageErrorf(c, err.Error()))
				return
			}

			serviceability.StartProfiler()

			if err := opts.StartMaster(stopCh); err != nil {
				if kerrors.IsInvalid(err) {
					if details := err.(*kerrors.StatusError).ErrStatus.Details; details != nil {
						fmt.Fprintf(errout, "Invalid %s %s\n", details.Kind, details.Name)
						for _, cause := range details.Causes {
							fmt.Fprintf(errout, "  %s: %s\n", cause.Field, cause.Message)
						}
						os.Exit(255)
					}
				}
				glog.Fatal(err)
			}
		},
	}

	// allow the master IP address to be overridden on a per process basis
	masterAddr := flagtypes.Addr{
		Value:         "127.0.0.1:8443",
		DefaultScheme: "https",
		DefaultPort:   8443,
		AllowPrefix:   true,
	}.Default()

	// allow the listen address to be overridden on a per process basis
	listenArg := &options.ListenArg{
		ListenAddr: flagtypes.Addr{
			Value:         "127.0.0.1:8444",
			DefaultScheme: "https",
			DefaultPort:   8444,
			AllowPrefix:   true,
		}.Default(),
	}

	opts.MasterArgs = NewDefaultMasterArgs()
	opts.MasterArgs.StartAPI = true
	opts.MasterArgs.OverrideConfig = func(config *configapi.MasterConfig) error {
		// we do not currently enable multi host etcd for the cluster
		config.EtcdConfig = nil
		if masterAddr.Provided {
			if ip := net.ParseIP(masterAddr.Host); ip != nil {
				glog.V(2).Infof("Using a masterIP override %q", ip)
				config.KubernetesMasterConfig.MasterIP = ip.String()
			}
		}
		if listenArg.ListenAddr.Provided {
			addr := listenArg.ListenAddr.URL.Host
			glog.V(2).Infof("Using a listen address override %q", addr)
			applyBindAddressOverride(addr, config)
		}
		return nil
	}

	flags := cmd.Flags()
	// This command only supports reading from config and the override master address
	flags.StringVar(&opts.ConfigFile, "config", "", "Location of the master configuration file to run from. Required")
	cmd.MarkFlagFilename("config", "yaml", "yml")
	flags.Var(&masterAddr, "master", "The address the master should register for itself. Defaults to the master address from the config.")
	options.BindListenArg(listenArg, flags, "")

	return cmd, opts
}

// applyBindAddressOverride takes a given address and overrides the relevant sections of a MasterConfig
// TODO: move into helpers
func applyBindAddressOverride(addr string, config *configapi.MasterConfig) {
	defaultHost, defaultPort, err := net.SplitHostPort(addr)
	if err != nil {
		// is just a host
		defaultHost = addr
	}
	config.ServingInfo.BindAddress = overrideAddress(config.ServingInfo.BindAddress, defaultHost, defaultPort)
	if config.EtcdConfig != nil {
		config.EtcdConfig.ServingInfo.BindAddress = overrideAddress(config.EtcdConfig.ServingInfo.BindAddress, defaultHost, "")
		config.EtcdConfig.PeerServingInfo.BindAddress = overrideAddress(config.EtcdConfig.PeerServingInfo.BindAddress, defaultHost, "")
	}
	if config.DNSConfig != nil {
		config.DNSConfig.BindAddress = overrideAddress(config.DNSConfig.BindAddress, defaultHost, "")
	}
}

// overrideAddress applies an optional host or port override to a incoming addr. If host or port are empty they will
// not override the existing addr values.
func overrideAddress(addr, host, port string) string {
	existingHost, existingPort, err := net.SplitHostPort(addr)
	if err != nil {
		if len(host) > 0 {
			return host
		}
		return addr
	}
	if len(host) > 0 {
		existingHost = host
	}
	if len(port) > 0 {
		existingPort = port
	}
	if len(existingPort) == 0 {
		return existingHost
	}
	return net.JoinHostPort(existingHost, existingPort)
}
