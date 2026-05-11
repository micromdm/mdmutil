# MDMUtil Operations Guide

This is a breif overview of working with the `mdmutil` CLI tool. MDMUtil is a tool for working with various data and aspects of Apple MDM.

## mdmutil

The `mdmutil` CLI is broken up into "sub"-commands which provide the actual functionality but the "root" `mdmutil` does have some shared and utility flags itself:

```
$ ./mdmutil-darwin-amd64 -h
./mdmutil-darwin-amd64 [flags] command [flags] 

Flags:
  -version
    	print version and exit
  ...

Commands:
  passwd
  ...
```

### Flags

Command line flags can be specified using command line arguments or in some cases environment variables.

#### -h, -help

Built-in flag that prints all other available flags and available commands.

#### -version

* print version and exit

Print version and exit.

## mdmutil passwd

The `passwd` command generates a salted SHA-512 PBKDF2 hash of a plaintext password. The output is an Apple Property List (plist) containing the `SALTED-SHA512-PBKDF2` dictionary. This is the format required for the [`AccountConfiguration`](https://developer.apple.com/documentation/devicemanagement/account-configuration-command) and [`SetAutoAdminPassword`](https://developer.apple.com/documentation/devicemanagement/set-auto-admin-password-command) MDM "v1" commands.

### Flags

#### -h, -help

Built-in flag that prints all other available flags.

#### -b64

* Output base64-encoded Plist

By default the plist output is in XML text form. Use this flag to output the plist as a base64-encoded string instead — which is the format accepted by the e.g. `SetAutoAdminPassword` MDM command's `passwordHash` key.

#### -password string

* password to hash (also as MDMUTIL_PASSWORD environment var)

The plaintext password to hash. Can also be supplied via the `MDMUTIL_PASSWORD` environment variable. At least one of the flag or environment variable must be provided.

### Example usage

Generate a password hash and print the plist:

```bash
$ ./mdmutil-darwin-amd64 passwd -password mysecretpassword
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>SALTED-SHA512-PBKDF2</key>
  <dict>
    [..snip..]
  </dict>
</dict>
</plist>
```

Generate a password hash as a base64-encoded plist:

```bash
$ ./mdmutil-darwin-amd64 passwd -password mysecretpassword -b64
PD94bWwgdmVyc2lvbj0iMS4wIiBlbmN[..snip..]
```
