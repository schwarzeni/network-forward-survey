sudo ufw disable

truncate -s 1G bigfile.iso


10.211.55.54 --> 10.211.55.57 --> 10.211.55.58 --> 10.211.55.59 --> 10.211.55.60 --> 10.211.55.61 --> 10.211.55.52

10.211.55.57

```bash
T_MOD="-A"
T_CURR_PORT="9001"
T_DST_PORT="9001"
T_CURR="10.211.55.57"
T_DST="10.211.55.58"

iptables -t filter -P FORWARD ACCEPT
iptables -t nat $T_MOD PREROUTING --dst $T_CURR -p tcp --dport $T_CURR_PORT  -j DNAT --to-destination $T_DST:$T_DST_PORT
iptables -t nat $T_MOD POSTROUTING --dst $T_DST -p tcp --dport $T_DST_PORT -j SNAT --to-source $T_CURR

# true version
# iptables -t nat $T_MOD PREROUTING --dst $T_CURR -p tcp --dport $T_CURR_PORT  -j DNAT --to-destination $T_DST:$T_DST_PORT
# iptables -t nat $T_MOD POSTROUTING --dst $T_DST -p tcp --dport $T_DST_PORT -j SNAT --to-source $T_CURR

# iptables -t nat $T_MOD PREROUTING --dst $T_CURR -p tcp --dport $T_CURR_PORT  -j DNAT --to-destination $T_DST:$T_DST_PORT
# iptables -t nat $T_MOD POSTROUTING --dst $T_DST -p tcp --dport $T_DST_PORT -j SNAT --to-source $T_CURR:$T_CURR_PORT
```

10.211.55.58

```bash
T_MOD="-A"
T_CURR_PORT="9001"
T_DST_PORT="9001"
T_CURR="10.211.55.58"
T_DST="10.211.55.59"

iptables -t filter -P FORWARD ACCEPT
iptables -t nat $T_MOD PREROUTING --dst $T_CURR -p tcp --dport $T_CURR_PORT  -j DNAT --to-destination $T_DST:$T_DST_PORT
iptables -t nat $T_MOD POSTROUTING --dst $T_DST -p tcp --dport $T_DST_PORT -j SNAT --to-source $T_CURR

```

10.211.55.59

```bash
T_MOD="-A"
T_CURR_PORT="9001"
T_DST_PORT="9001"
T_CURR="10.211.55.59"
T_DST="10.211.55.52"

iptables -t filter -P FORWARD ACCEPT
iptables -t nat $T_MOD PREROUTING --dst $T_CURR -p tcp --dport $T_CURR_PORT  -j DNAT --to-destination $T_DST:$T_DST_PORT
iptables -t nat $T_MOD POSTROUTING --dst $T_DST -p tcp --dport $T_DST_PORT -j SNAT --to-source $T_CURR

```


10.211.55.60

```bash
T_MOD="-A"
T_CURR_PORT="9001"
T_DST_PORT="9001"
T_CURR="10.211.55.60"
T_DST="10.211.55.61"

iptables -t filter -P FORWARD ACCEPT
iptables -t nat $T_MOD PREROUTING --dst $T_CURR -p tcp --dport $T_CURR_PORT  -j DNAT --to-destination $T_DST:$T_DST_PORT
iptables -t nat $T_MOD POSTROUTING --dst $T_DST -p tcp --dport $T_DST_PORT -j SNAT --to-source $T_CURR

```

10.211.55.61

```bash
T_MOD="-A"
T_CURR_PORT="9001"
T_DST_PORT="9001"
T_CURR="10.211.55.61"
T_DST="10.211.55.52"

iptables -t filter -P FORWARD ACCEPT
iptables -t nat $T_MOD PREROUTING --dst $T_CURR -p tcp --dport $T_CURR_PORT  -j DNAT --to-destination $T_DST:$T_DST_PORT
iptables -t nat $T_MOD POSTROUTING --dst $T_DST -p tcp --dport $T_DST_PORT -j SNAT --to-source $T_CURR

```