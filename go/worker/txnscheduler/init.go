package txnscheduler

import (
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	workerCommon "github.com/oasislabs/oasis-core/go/worker/common"
	"github.com/oasislabs/oasis-core/go/worker/compute"
	"github.com/oasislabs/oasis-core/go/worker/registration"
	txnSchedulerAlgorithm "github.com/oasislabs/oasis-core/go/worker/txnscheduler/algorithm"
)

const (
	// CfgWorkerEnabled enables the tx scheduler worker.
	CfgWorkerEnabled = "worker.txn_scheduler.enabled"
	// CfgCheckTxEnabled enables checking each transaction before scheduling it.
	CfgCheckTxEnabled = "worker.txn_scheduler.check_tx.enabled"
)

// Flags has the configuration flags.
var Flags = flag.NewFlagSet("", flag.ContinueOnError)

// Enabled reads our enabled flag from viper.
func Enabled() bool {
	return viper.GetBool(CfgWorkerEnabled)
}

// CheckTxEnabled reads our CheckTx enabled flag from viper.
func CheckTxEnabled() bool {
	return viper.GetBool(CfgCheckTxEnabled)
}

// New creates a new worker.
func New(
	commonWorker *workerCommon.Worker,
	compute *compute.Worker,
	registration *registration.Worker,
) (*Worker, error) {
	return newWorker(Enabled(), commonWorker, compute, registration, CheckTxEnabled())
}

func init() {
	Flags.Bool(CfgWorkerEnabled, false, "Enable transaction scheduler process")
	Flags.Bool(CfgCheckTxEnabled, false, "Enable checking transactions before scheduling them")

	_ = viper.BindPFlags(Flags)

	Flags.AddFlagSet(txnSchedulerAlgorithm.Flags)
}
