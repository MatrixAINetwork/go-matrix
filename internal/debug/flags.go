// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

package debug

import (
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/MatrixAINetwork/go-matrix/log"
	"github.com/MatrixAINetwork/go-matrix/log/term"
	"github.com/MatrixAINetwork/go-matrix/metrics"
	"github.com/MatrixAINetwork/go-matrix/metrics/exp"
	"github.com/fjl/memsize/memsizeui"
	colorable "github.com/mattn/go-colorable"
	"gopkg.in/urfave/cli.v1"
)

var Memsize memsizeui.Handler

var (
	verbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
		Value: 3,
	}
	verOutPutFlag = cli.IntFlag{
		Name:  "outputinfo",
		Usage: "Logging output: 0=console, 1=file, 2=both",
		Value: 0,
	}
	verOutPutDirFlag = cli.StringFlag{
		Name:  "outputdir",
		Usage: "Logging output dir:eg:linux  /var/log/MatrixLog, window C:\\MatrixLog, otherwise null means current path /MatrixLog/",
		Value: "",
	}
	verNetlogFlag = cli.StringFlag{
		Name:  "netlog",
		Usage: "netlog addr,egg: 192.168.1.12:514 ",
		Value: "",
	}
	verNetlogModeFlag = cli.StringFlag{
		Name:  "netlogmode",
		Usage: "netlog mode,egg: tcp,udp ",
		Value: "udp",
	}
	vmoduleFlag = cli.StringFlag{
		Name:  "vmodule",
		Usage: "Per-module verbosity: comma-separated list of <pattern>=<level> (e.g. man/*=5,p2p=4)",
		Value: "",
	}
	backtraceAtFlag = cli.StringFlag{
		Name:  "backtrace",
		Usage: "Request a stack trace at a specific logging statement (e.g. \"block.go:271\")",
		Value: "",
	}
	debugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "Prepends log messages with call-site location (file and line number)",
	}
	pprofFlag = cli.BoolFlag{
		Name:  "pprof",
		Usage: "Enable the pprof HTTP server",
	}
	pprofPortFlag = cli.IntFlag{
		Name:  "pprofport",
		Usage: "pprof HTTP server listening port",
		Value: 6060,
	}
	pprofAddrFlag = cli.StringFlag{
		Name:  "pprofaddr",
		Usage: "pprof HTTP server listening interface",
		Value: "127.0.0.1",
	}
	memprofilerateFlag = cli.IntFlag{
		Name:  "memprofilerate",
		Usage: "Turn on memory profiling with the given rate",
		Value: runtime.MemProfileRate,
	}
	blockprofilerateFlag = cli.IntFlag{
		Name:  "blockprofilerate",
		Usage: "Turn on block profiling with the given rate",
	}
	cpuprofileFlag = cli.StringFlag{
		Name:  "cpuprofile",
		Usage: "Write CPU profile to the given file",
	}
	traceFlag = cli.StringFlag{
		Name:  "trace",
		Usage: "Write execution trace to the given file",
	}
)

// Flags holds all command-line flags required for debugging.
var Flags = []cli.Flag{
	verbosityFlag, verOutPutFlag, verOutPutDirFlag, verNetlogFlag, verNetlogModeFlag, vmoduleFlag, backtraceAtFlag, debugFlag,
	pprofFlag, pprofAddrFlag, pprofPortFlag,
	memprofilerateFlag, blockprofilerateFlag, cpuprofileFlag, traceFlag,
}

var (
	ostream log.Handler
	glogger *log.GlogHandler
)

func init() {
	usecolor := term.IsTty(os.Stdout.Fd()) && os.Getenv("TERM") != "dumb"

	output := io.Writer(os.Stdout)

	if usecolor {
		output = colorable.NewColorableStdout()
	}
	ostream = log.StreamHandler(output, log.TerminalFormat(usecolor))
	glogger = log.NewGlogHandler(ostream)
}

// Setup initializes profiling and logging based on the CLI flags.
// It should be called as early as possible in the program.
func Setup(ctx *cli.Context, logdir string) error {
	// logging
	log.PrintOrigins(ctx.GlobalBool(debugFlag.Name))
	logPath := ctx.GlobalString(verOutPutDirFlag.Name)
	if logPath == "" {
		logPath = "MatrixLog"
	}
	flgnet := ctx.GlobalString(verNetlogFlag.Name)
	var netHandler log.Handler = nil
	var err error
	if len(flgnet) > 0 {
		//netHandler,err = log.SetNetLogHandler("udp", "192.168.122.7:514","gman_netlog",log.TerminalFormat(false))
		mode := ctx.GlobalString(verNetlogModeFlag.Name)
		if len(mode) == 0 {
			mode = "udp"
		}
		netHandler, err = log.SetNetLogHandler(mode, flgnet, "gman_netlog", log.TerminalFormat(false))
		if err != nil {
			netHandler = nil
		}
	}
	//neth,err:= log.SetNetLogHandler("udp", "192.168.122.7:514","gman_netlog",log.TerminalFormat(false))

	flg := ctx.GlobalInt(verOutPutFlag.Name)
	if flg > 0 {
		if logPath != "" {
			rfh, err := log.RotatingFileHandler(
				logPath,      //logdir,
				1024*1024*50, //262144,
				/*log.JSONFormatOrderedEx(false, true),*/
				log.TerminalFormat(false),
			)
			if err != nil {
				return err
			}
			if netHandler == nil {
				if flg == 1 {
					glogger.SetHandler(log.MultiHandler(rfh))
				} else {
					glogger.SetHandler(log.MultiHandler(ostream, rfh))
				}
			} else {
				if flg == 1 {
					glogger.SetHandler(log.MultiHandler(rfh, netHandler))
				} else {
					glogger.SetHandler(log.MultiHandler(ostream, rfh, netHandler))
				}
			}
		}
	}
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(verbosityFlag.Name)))
	glogger.Vmodule(ctx.GlobalString(vmoduleFlag.Name))
	glogger.BacktraceAt(ctx.GlobalString(backtraceAtFlag.Name))
	log.Root().SetHandler(glogger)

	// profiling, tracing
	runtime.MemProfileRate = ctx.GlobalInt(memprofilerateFlag.Name)
	Handler.SetBlockProfileRate(ctx.GlobalInt(blockprofilerateFlag.Name))
	if traceFile := ctx.GlobalString(traceFlag.Name); traceFile != "" {
		if err := Handler.StartGoTrace(traceFile); err != nil {
			return err
		}
	}
	if cpuFile := ctx.GlobalString(cpuprofileFlag.Name); cpuFile != "" {
		if err := Handler.StartCPUProfile(cpuFile); err != nil {
			return err
		}
	}

	// pprof server
	if ctx.GlobalBool(pprofFlag.Name) {
		address := fmt.Sprintf("%s:%d", ctx.GlobalString(pprofAddrFlag.Name), ctx.GlobalInt(pprofPortFlag.Name))
		StartPProf(address)
	}
	return nil
}

func StartPProf(address string) {
	// Hook go-metrics into expvar on any /debug/metrics request, load all vars
	// from the registry into expvar, and execute regular expvar handler.
	exp.Exp(metrics.DefaultRegistry)
	http.Handle("/memsize/", http.StripPrefix("/memsize", &Memsize))
	log.Info("Starting pprof server", "addr", fmt.Sprintf("http://%s/debug/pprof", address))
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Error("Failure in running pprof server", "err", err)
		}
	}()
}

// Exit stops all running profiles, flushing their output to the
// respective file.
func Exit() {
	Handler.StopCPUProfile()
	Handler.StopGoTrace()
}
