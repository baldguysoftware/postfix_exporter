# Postfix Exporter for Prometheus

This is a simple server that periodically scrapes postfix queue status and exports them via HTTP for Prometheus
consumption.

# Installation

To install it:

```bash
git clone https://github.com/baldguysoftware/postfix_exporter.git
cd postfix_exporter
make
```
# Running

To run it manually...

```bash
sudo ./postfix_exporter [flags]
```
# Configuration 

Help on flags:
```bash
./postfix_exporter --help
```

A default installation of Postfix storing its queue directories under
`/var/spool/postfix` will require no options. The exporter will automatically
descend all hashed directories when it finds them - no added config needed.


# TODO
* I want to add native (ie. no shell calls requiring additional tools) breakout of destination domain counts.
