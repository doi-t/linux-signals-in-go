## How `SIGSTOP` and `SIGCONT` works

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

# Ex.1

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

## Sender
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

## Receiver
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

# Ex.2
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

## Sender

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

## Receiver

```shell
receiver:$ # Ex. 2
receiver:$ 5: Received SIGCONT (continued)
```

</p>
</details>

# Ex.3

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

## Sender

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

## Receiver

```shell
receiver:$ # Ex.3
receiver:$ 6: Received SIGCONT (continued)
```

</p>
</details>

# Ex.4

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

## Sender

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

## Receiver

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
