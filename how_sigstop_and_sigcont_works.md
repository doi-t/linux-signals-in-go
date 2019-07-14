# How `SIGSTOP` and `SIGCONT` works

## `SIGTERM` does not terminate a process when it stops

All signals except for `SIGKILL` will be queued until a stopped process resumes after `SIGCONT`.
```shell
root/go# sleep 1000 &
[1] 208
root/go# egrep 'State|SigQ' /proc/208/status
State:	S (sleeping)
SigQ:	0/7867
root/go# kill -s SIGSTOP 208
root/go# egrep 'State|SigQ' /proc/208/status
State:	T (stopped)
SigQ:	0/7867

[1]+  Stopped                 sleep 1000
root/go# kill -s SIGTERM 208
root/go# egrep 'State|SigQ' /proc/208/status
State:	T (stopped)
SigQ:	1/7867
root/go# kill -s SIGTSTP 208
root/go# egrep 'State|SigQ' /proc/208/status
State:	T (stopped)
SigQ:	2/7867
root/go# kill -s SIGINT 208
root/go# egrep 'State|SigQ' /proc/208/status
State:	T (stopped)
SigQ:	3/7867
root/go# kill -s SIGCONT 208
root/go# egrep 'State|SigQ' /proc/208/status
grep: /proc/208/status: No such file or directory
[1]+  Interrupt               sleep 1000
```

## Preparation
Open a terminal and start receiver inside golang container.

```shell
./run_golang_container.sh
# Now you are in golang container
export PS1='receiver:$ '
go build sigexp.go # generated binary will be shared between sender and receiver
./sigexp -mode=receiver
```

<details><summary>env info at the moment of experiments</summary>
<p>

```
$ uname -a
Linux 9f3368d0a9e5 4.9.125-linuxkit #1 SMP Fri Sep 7 08:20:28 UTC 2018 x86_64 GNU/Linux
$ cat /etc/os-release
PRETTY_NAME="Debian GNU/Linux 9 (stretch)"
NAME="Debian GNU/Linux"
VERSION_ID="9"
VERSION="9 (stretch)"
ID=debian
HOME_URL="https://www.debian.org/"
SUPPORT_URL="https://www.debian.org/support"
BUG_REPORT_URL="https://bugs.debian.org/"
```

</p>
</details>

Open another terminal and prepare sender console inside the same golang container.

```shell
docker exec -it dev_golang /bin/bash # Login to running container
# Now you are in golang container
export PS1='sender:$ '
```

## Check the existence of signal queue

Make sure signal queue count. You will see changes of 'SigQ'.
Since Linux kernel clean up duplicated signals in the signal queue, the same signals won't increment the queue counter.

```shell
egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
```

<details><summary>Result</summary>
<p>

### Sender

```shell
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	S (sleeping)
SigQ:	0/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent: SIGSTOP (stopped (signal))
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
Sent: SIGTERM (terminated)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
Sent: SIGTERM (terminated)
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	T (stopped)
SigQ:	1/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
Sent: SIGINT (interrupt)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
Sent: SIGINT (interrupt)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
Sent: SIGINT (interrupt)
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	T (stopped)
SigQ:	2/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent: SIGPIPE (broken pipe)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent: SIGPIPE (broken pipe)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent: SIGPIPE (broken pipe)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent: SIGPIPE (broken pipe)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent: SIGPIPE (broken pipe)
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	T (stopped)
SigQ:	3/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent: SIGCONT (continued)
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	S (sleeping)
SigQ:	0/7867
```

### Receiver
```shell
receiver:$ ./sigexp -mode=receiver
PID: 52

[1]+  Stopped                 ./sigexp -mode=receiver
receiver:$
receiver:$ Received: SIGINT (interrupt)
Received: SIGPIPE (broken pipe)
Received: SIGTERM (terminated)
Received: SIGCONT (continued)
```

</p>
</details>

## Some type of signals are the same
As you can see in [here](https://github.com/golang/go/blob/release-branch.go1.12/src/syscall/zerrors_linux_amd64.go#L1341-L1378), some type of signals are exactly the same. For example, `SIGABRT` and `SIGIOT` are `Signal(0x6)`. If we send `SIGABRT` and `SIGIOT` during signal stop, the second signal will be canceled because they are the same signal.

```shell
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGABRT
egrep 'SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGIOT
egrep 'SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
egrep 'SigQ' /proc/$(pgrep sigexp)/status
```

<details><summary>Result</summary>
<p>

### Sender

`SigQ` didn't become `2`.
```shell
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent: SIGSTOP (stopped (signal))
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGABRT
Sent: SIGABRT (aborted)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	1/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGIOT
Sent: SIGIOT (aborted)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	1/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent: SIGCONT (continued)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
sender:$
```

### Receiver
```shell
receiver:$ ./sigexp -mode=receiver
PID: 133

[1]+  Stopped                 ./sigexp -mode=receiver
receiver:$ Received: SIGABRT SIGIOT (aborted)
Received: SIGCONT (continued)
```

</p>
</details>

## Any pending stop signals are discorded
The following signals will be discarded when receiver receives `SIGCONT`.

- SIGSTOP (stopped (signal))
- SIGTSTP (stopped)
- SIGTTIN (stopped (tty input))
- SIGTTOU (stopped (tty output))

> While a process is stopped, no more signals can be delivered to it until it is continued, except SIGKILL signals and (obviously) SIGCONT signals. The signals are marked as pending, but not delivered until the process is continued. The SIGKILL signal always causes termination of the process and can’t be blocked, handled or ignored. You can ignore SIGCONT, but it always causes the process to be continued anyway if it is stopped. Sending a SIGCONT signal to a process causes any pending stop signals for that process to be discarded. Likewise, any pending SIGCONT signals for a process are discarded when it receives a stop signal.
> Ref. https://www.gnu.org/software/libc/manual/html_node/Job-Control-Signals.html

```shell
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
egrep 'SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTP
egrep 'SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTTIN
egrep 'SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTTOU
egrep 'SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
egrep 'SigQ' /proc/$(pgrep sigexp)/status
```

<details><summary>Result</summary>
<p>

### Sender

```shell
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent: SIGSTOP (stopped (signal))
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTP
Unknown signal name: SIGSTP
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTTIN
Sent: SIGTTIN (stopped (tty input))
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	1/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTTOU
Sent: SIGTTOU (stopped (tty output))
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	2/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent: SIGCONT (continued)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
```

### Receiver

```shell
receiver:$ ./sigexp -mode=receiver
PID: 171

[1]+  Stopped                 ./sigexp -mode=receiver
receiver:$ Received: SIGCONT (continued)
```

</p>
</details>

## Can't handle SIGPROF

It seems that `SIGPROF` is also discorded after `SIGCONT`. I'm not sure why `SIGPROF` won't be handled. :thinking_face:

- SIGPROF (profiling timer expired)

> The SIGPROF signal is handled directly by the Go runtime to implement runtime.CPUProfile.
> Ref. https://golang.org/pkg/os/signal/#hdr-Default_behavior_of_signals_in_Go_programs

Does Go runtime do something?

```shell
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
egrep 'SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPROF
egrep 'SigQ' /proc/$(pgrep sigexp)/status
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
egrep 'SigQ' /proc/$(pgrep sigexp)/status
```

<details><summary>Result</summary>
<p>

### Sender

```shell
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent: SIGSTOP (stopped (signal))
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPROF
Sent: SIGPROF (profiling timer expired)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	1/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent: SIGCONT (continued)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
```

### Receiver

```shell
receiver:$ ./sigexp -mode=receiver
PID: 219

[1]+  Stopped                 ./sigexp -mode=receiver
receiver:$ Received: SIGCONT (continued)
```

</p>
</details>

## Reset signals

Reset SIGTERM and send it to terminate the receiver.

```shell
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGUSR1 # I customized this signal to reset SIGTERM and SIGINT
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
```

<details><summary>Result</summary>
<p>

### Sender

```shell
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent: SIGSTOP (stopped (signal))
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGUSR1 # I customized this signal to reset SIGTERM and SIGINT
Sent: SIGUSR1 (user defined signal 1)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
Sent: SIGTERM (terminated)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent: SIGCONT (continued)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
Sent: SIGTERM (terminated)
```

### Receiver

```shell
receiver:$ ./sigexp -mode=receiver
PID: 252

[1]+  Stopped                 ./sigexp -mode=receiver
receiver:$ Received: SIGUSR1 (user defined signal 1)
(Reset SIGINT and SIGTERM. Now you can interrupt this program with SIGINT and SIGTERM)
Received: SIGTERM (terminated)
Received: SIGCONT (continued)

[1]+  Terminated              ./sigexp -mode=receiver
```

</p>
</details>

## Send all signals

Test [all signals](https://github.com/golang/go/blob/release-branch.go1.12/src/syscall/zerrors_linux_amd64.go#L1341-L1378).

```shell
./sigexp -mode=sender -pid=$(pgrep sigexp)
```

<details><summary>Result</summary>
<p>

### Sender

```shell
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp)
Sent: SIGSTOP (stopped (signal))
Sent: SIGUSR1 (user defined signal 1)
(Skipped to send SIGCONT (continued))
Sent: SIGFPE (floating point exception)
Sent: SIGPROF (profiling timer expired)
Sent: SIGQUIT (quit)
Sent: SIGSEGV (segmentation fault)
Sent: SIGCHLD (child exited)
Sent: SIGSTOP (stopped (signal))
Sent: SIGTTIN (stopped (tty input))
Sent: SIGXCPU (CPU time limit exceeded)
Sent: SIGTTOU (stopped (tty output))
Sent: SIGBUS (bus error)
Sent: SIGHUP (hangup)
(Skipped to send SIGKILL (killed))
Sent: SIGPIPE (broken pipe)
Sent: SIGSTKFLT (stack fault)
Sent: SIGCLD (child exited)
Sent: SIGILL (illegal instruction)
Sent: SIGIO (I/O possible)
Sent: SIGPWR (power failure)
Sent: SIGUNUSED (bad system call)
Sent: SIGALRM (alarm clock)
Sent: SIGIOT (aborted)
Sent: SIGTSTP (stopped)
Sent: SIGURG (urgent I/O condition)
Sent: SIGTERM (terminated)
Sent: SIGTRAP (trace/breakpoint trap)
Sent: SIGUSR2 (user defined signal 2)
Sent: SIGPOLL (I/O possible)
Sent: SIGSYS (bad system call)
Sent: SIGVTALRM (virtual timer expired)
Sent: SIGWINCH (window changed)
Sent: SIGXFSZ (file size limit exceeded)
Sent: SIGABRT (aborted)
Sent: SIGINT (interrupt)
Sent: SIGCONT (continued)
```

### Receiver

As we confirmed above, receiver is missing the following signals even sender sent them during `Stopped` status.
You can also see some type of signals are the same in Go as we discussed it.
`SIGTERM` does not terminate receiver process after resetting it.
If we send `SIGTERM` again, the process will be terminated.

- SIGPROF (profiling timer expired)
- SIGSTOP (stopped (signal))
- SIGTSTP (stopped)
- SIGTTIN (stopped (tty input))
- SIGTTOU (stopped (tty output))

```shell
receiver:$ ./sigexp -mode=receiver
PID: 285

[1]+  Stopped                 ./sigexp -mode=receiver
receiver:$ Received: SIGHUP (hangup)
Received: SIGINT (interrupt)
Received: SIGQUIT (quit)
Received: SIGILL (illegal instruction)
Received: SIGTRAP (trace/breakpoint trap)
Received: SIGABRT SIGIOT (aborted)
Received: SIGBUS (bus error)
Received: SIGFPE (floating point exception)
Received: SIGUSR1 (user defined signal 1)
(Reset SIGINT and SIGTERM. Now you can interrupt this program with SIGINT and SIGTERM)
Received: SIGSEGV (segmentation fault)
Received: SIGUSR2 (user defined signal 2)
Received: SIGPIPE (broken pipe)
Received: SIGALRM (alarm clock)
Received: SIGTERM (terminated)
Received: SIGSTKFLT (stack fault)
Received: SIGCLD SIGCHLD (child exited)
Received: SIGCONT (continued)
Received: SIGURG (urgent I/O condition)
Received: SIGXCPU (CPU time limit exceeded)
Received: SIGXFSZ (file size limit exceeded)
Received: SIGVTALRM (virtual timer expired)
Received: SIGWINCH (window changed)
Received: SIGIO SIGPOLL (I/O possible)
Received: SIGPWR (power failure)
Received: SIGSYS SIGUNUSED (bad system call)
```

</p>
</details>
