package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Ref. https://github.com/opencontainers/runc/blob/master/signalmap.go#L12
// Ref. https://github.com/golang/go/blob/release-branch.go1.12/src/syscall/zerrors_linux_amd64.go#L1341-L1378
var signalMap = map[string]syscall.Signal{
	"SIGABRT":   syscall.SIGABRT,
	"SIGALRM":   syscall.SIGALRM,
	"SIGBUS":    syscall.SIGBUS,
	"SIGCHLD":   syscall.SIGCHLD,
	"SIGCLD":    syscall.SIGCLD,
	"SIGCONT":   syscall.SIGCONT,
	"SIGFPE":    syscall.SIGFPE,
	"SIGHUP":    syscall.SIGHUP,
	"SIGILL":    syscall.SIGILL,
	"SIGINT":    syscall.SIGINT,
	"SIGIO":     syscall.SIGIO,
	"SIGIOT":    syscall.SIGIOT,
	"SIGKILL":   syscall.SIGKILL,
	"SIGPIPE":   syscall.SIGPIPE,
	"SIGPOLL":   syscall.SIGPOLL,
	"SIGPROF":   syscall.SIGPROF,
	"SIGPWR":    syscall.SIGPWR,
	"SIGQUIT":   syscall.SIGQUIT,
	"SIGSEGV":   syscall.SIGSEGV,
	"SIGSTKFLT": syscall.SIGSTKFLT,
	"SIGSTOP":   syscall.SIGSTOP,
	"SIGSYS":    syscall.SIGSYS,
	"SIGTERM":   syscall.SIGTERM,
	"SIGTRAP":   syscall.SIGTRAP,
	"SIGTSTP":   syscall.SIGTSTP,
	"SIGTTIN":   syscall.SIGTTIN,
	"SIGTTOU":   syscall.SIGTTOU,
	"SIGUNUSED": syscall.SIGUNUSED,
	"SIGURG":    syscall.SIGURG,
	"SIGUSR1":   syscall.SIGUSR1,
	"SIGUSR2":   syscall.SIGUSR2,
	"SIGVTALRM": syscall.SIGVTALRM,
	"SIGWINCH":  syscall.SIGWINCH,
	"SIGXCPU":   syscall.SIGXCPU,
	"SIGXFSZ":   syscall.SIGXFSZ,
}

func main() {
	mode := flag.String("mode", "receiver", "'sender' (signal sender) or 'receiver' (signal receiver)")
	pid := flag.Int("pid", 0, "target process id")
	signame := flag.String("signal", "", "If it is not given, all signals except for SIGKILL will be sent.")
	flag.Parse()

	if *mode == "receiver" {
		signalReciever()
	} else if *mode == "sender" {
		if *pid == 0 {
			fmt.Printf("Usage: %v -mode=sender -pid=<target process id>\n", os.Args[0])
			os.Exit(1)
		}
		signalSender(*pid, *signame)
	} else {
		fmt.Printf("Unknown mode: %v (mode should be either 'sender' or 'receiver')\n", *mode)
		os.Exit(1)
	}
}

func signalReciever() {
	fmt.Printf("PID: %v\n", os.Getpid())
	signals := make(chan os.Signal, len(signalMap)) // Signal queue for pending/queued signals
	signal.Notify(signals)                          //  All incoming signals will be relayed to 'signals'
	i := 0
	for {
		select {
		case s := <-signals:
			i++
			for k, v := range signalMap {
				if v == s {
					fmt.Printf("%v: Received %v (%v)\n", i, k, s)
				}
			}
			if s == syscall.SIGUSR1 {
				signal.Reset(syscall.SIGINT, syscall.SIGTERM)
				fmt.Println("Reset SIGINT and SIGTERM. Now you can interrupt this program with SIGINT and SIGTERM")
			}
		}
	}
}

func signalSender(pid int, signame string) {
	if signame != "" {
		sendSignal(pid, signame)
	} else {
		sendAllSignals(pid)
	}
}
func sendSignal(pid int, signame string) {
	if s, ok := signalMap[signame]; ok {
		err := syscall.Kill(pid, s)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Sent %v (%v)\n", signame, s)
	} else {
		fmt.Printf("Unknown signal name: %v\n", signame)
	}
}

func sendAllSignals(pid int) {
	sendSignal(pid, "SIGSTOP")
	i := 0
	for s := range signalMap {
		i++
		fmt.Printf("%v: ", i)
		if s != "SIGKILL" && s != "SIGCONT" {
			sendSignal(pid, s) // All signals will be queued stop signals.
		} else {
			fmt.Printf("Skipped to send %v (%v)\n", s, signalMap[s])
		}
	}
	sendSignal(pid, "SIGCONT")
}
