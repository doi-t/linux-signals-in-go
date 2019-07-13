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

## Ex.1

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

<details><summary>Ex.1 result</summary>
<p>

### Sender
```shell
sender:$ # Ex.1
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	S (sleeping)
SigQ:	0/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent SIGSTOP (stopped (signal))
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
Sent SIGTERM (terminated)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
Sent SIGTERM (terminated)
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	T (stopped)
SigQ:	1/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
Sent SIGINT (interrupt)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
Sent SIGINT (interrupt)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGINT
Sent SIGINT (interrupt)
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	T (stopped)
SigQ:	2/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent SIGPIPE (broken pipe)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent SIGPIPE (broken pipe)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent SIGPIPE (broken pipe)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent SIGPIPE (broken pipe)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPIPE
Sent SIGPIPE (broken pipe)
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	T (stopped)
SigQ:	3/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent SIGCONT (continued)
sender:$ egrep 'State|SigQ' /proc/$(pgrep sigexp)/status
State:	S (sleeping)
SigQ:	0/7867
```

### Receiver
```shell
receiver:$ ./sigexp -mode=receiver
PID: 40

[1]+  Stopped                 ./sigexp -mode=receiver
receiver:$ 1: Received SIGINT (interrupt)
2: Received SIGPIPE (broken pipe)
3: Received SIGTERM (terminated)
4: Received SIGCONT (continued)
```

</p>
</details>

## Ex.2
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

<details><summary>Ex.2 result</summary>
<p>

### Sender

```shell
sender:$ # Ex. 2
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent SIGSTOP (stopped (signal))
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTP
Unknown signal name: SIGSTP
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTTIN
Sent SIGTTIN (stopped (tty input))
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	1/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTTOU
Sent SIGTTOU (stopped (tty output))
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	2/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent SIGCONT (continued)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
```

### Receiver

```shell
receiver:$ # Ex. 2
receiver:$ 5: Received SIGCONT (continued)
```

</p>
</details>

## Ex.3

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

<details><summary>Ex.2 result</summary>
<p>

### Sender

```shell
sender:$ # Ex. 3
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent SIGSTOP (stopped (signal))
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGPROF
Sent SIGPROF (profiling timer expired)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	1/7867
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent SIGCONT (continued)
sender:$ egrep 'SigQ' /proc/$(pgrep sigexp)/status
SigQ:	0/7867
```

### Receiver

```shell
receiver:$ # Ex.3
receiver:$ 6: Received SIGCONT (continued)
```

</p>
</details>

## Ex.4

Reset SIGTERM and send it to terminate the receiver.

```shell
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGUSR1 # I customized this signal to reset SIGTERM and SIGINT
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
```

<details><summary>Ex.2 result</summary>
<p>

### Sender

```shell
sender:$ # Ex.4
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGSTOP
Sent SIGSTOP (stopped (signal))
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGUSR1 # I customized this signal to reset SIGTERM and SIGINT
Sent SIGUSR1 (user defined signal 1)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
Sent SIGTERM (terminated)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGCONT
Sent SIGCONT (continued)
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp) -signal=SIGTERM
Sent SIGTERM (terminated)
```

### Receiver

```shell
receiver:$ #. Ex.4
receiver:$ 7: Received SIGUSR1 (user defined signal 1)
Reset SIGINT and SIGTERM. Now you can interrupt this program with SIGINT and SIGTERM
8: Received SIGTERM (terminated)
9: Received SIGCONT (continued)

[1]+  Terminated              ./sigexp -mode=receiver
```

</p>
</details>

## Ex.5

Test [all signals](https://github.com/golang/go/blob/release-branch.go1.12/src/syscall/zerrors_linux_amd64.go#L1341-L1378).

```shell
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp)
```

<details><summary>Ex.5 result</summary>
<p>

### Sender

```shell
sender:$ ./sigexp -mode=sender -pid=$(pgrep sigexp)
Sent SIGSTOP (stopped (signal))
1: Sent SIGALRM (alarm clock)
2: Sent SIGCLD (child exited)
3: Skipped to send SIGCONT (continued)
4: Sent SIGFPE (floating point exception)
5: Sent SIGIO (I/O possible)
6: Sent SIGPIPE (broken pipe)
7: Sent SIGTERM (terminated)
8: Sent SIGTTIN (stopped (tty input))
9: Sent SIGUNUSED (bad system call)
10: Sent SIGCHLD (child exited)
11: Sent SIGHUP (hangup)
12: Sent SIGILL (illegal instruction)
13: Sent SIGIOT (aborted)
14: Sent SIGPROF (profiling timer expired)
15: Sent SIGUSR1 (user defined signal 1)
16: Sent SIGTRAP (trace/breakpoint trap)
17: Sent SIGURG (urgent I/O condition)
18: Sent SIGBUS (bus error)
19: Sent SIGINT (interrupt)
20: Sent SIGSTOP (stopped (signal))
21: Sent SIGSYS (bad system call)
22: Sent SIGWINCH (window changed)
23: Sent SIGXFSZ (file size limit exceeded)
24: Sent SIGABRT (aborted)
25: Sent SIGPOLL (I/O possible)
26: Sent SIGQUIT (quit)
27: Sent SIGSEGV (segmentation fault)
28: Sent SIGTTOU (stopped (tty output))
29: Sent SIGUSR2 (user defined signal 2)
30: Skipped to send SIGKILL (killed)
31: Sent SIGPWR (power failure)
32: Sent SIGSTKFLT (stack fault)
33: Sent SIGTSTP (stopped)
34: Sent SIGVTALRM (virtual timer expired)
35: Sent SIGXCPU (CPU time limit exceeded)
Sent SIGCONT (continued)
```

### Receiver

As we confirmed above, receiver is missing the following signals even sender sent them during `Stopped` status.
`SIGTERM` does not terminate receiver process after resetting it.
If we send `SIGTERM` again, the process will be terminated.

- SIGPROF (profiling timer expired)
- SIGSTOP (stopped (signal))
- SIGTSTP (stopped)
- SIGTTIN (stopped (tty input))
- SIGTTOU (stopped (tty output))

```shell
receiver:$ ./sigexp -mode=receiver
PID: 237

[1]+  Stopped                 ./sigexp -mode=receiver
receiver:$ 1: Received SIGHUP (hangup)
2: Received SIGINT (interrupt)
3: Received SIGQUIT (quit)
4: Received SIGILL (illegal instruction)
5: Received SIGTRAP (trace/breakpoint trap)
6: Received SIGABRT (aborted)
6: Received SIGIOT (aborted)
7: Received SIGBUS (bus error)
8: Received SIGFPE (floating point exception)
9: Received SIGUSR1 (user defined signal 1)
Reset SIGINT and SIGTERM. Now you can interrupt this program with SIGINT and SIGTERM
10: Received SIGSEGV (segmentation fault)
11: Received SIGUSR2 (user defined signal 2)
12: Received SIGPIPE (broken pipe)
13: Received SIGALRM (alarm clock)
14: Received SIGTERM (terminated)
15: Received SIGSTKFLT (stack fault)
16: Received SIGCHLD (child exited)
16: Received SIGCLD (child exited)
17: Received SIGCONT (continued)
18: Received SIGURG (urgent I/O condition)
19: Received SIGXCPU (CPU time limit exceeded)
20: Received SIGXFSZ (file size limit exceeded)
21: Received SIGVTALRM (virtual timer expired)
22: Received SIGWINCH (window changed)
23: Received SIGPOLL (I/O possible)
23: Received SIGIO (I/O possible)
24: Received SIGPWR (power failure)
25: Received SIGUNUSED (bad system call)
25: Received SIGSYS (bad system call)
```

</p>
</details>
