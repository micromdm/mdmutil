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

The `mdmutil passwd` command generates a salted SHA-512 PBKDF2 hash of a plaintext password. The output is an Apple Property List (plist) containing the `SALTED-SHA512-PBKDF2` dictionary. This is the format required for the [`AccountConfiguration`](https://developer.apple.com/documentation/devicemanagement/account-configuration-command) and [`SetAutoAdminPassword`](https://developer.apple.com/documentation/devicemanagement/set-auto-admin-password-command) MDM "v1" commands.

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

## mdmutil mdmcsr-sign

The `mdmutil mdmcsr-sign` command signs an APNs push certificate using an Apple MDM CSR keypair and generates the Base64 encoded property list required by the [Apple Push Certificates Portal](https://identity.apple.com/). This is described by Apple as [Setting Up Push Notifications for Your MDM Customers](https://developer.apple.com/documentation/devicemanagement/setting-up-push-notifications-for-your-mdm-customers). See also the [MicroMDM "Understanding MDM Certificates" blog post](https://micromdm.io/blog/certificates/).

### Flags

#### -h, -help

Built-in flag that prints all other available flags and available commands.

#### -apns-csr string

* Path to APNs CSR that the MDMCSR private key will sign in PEM format

#### -intermediate string

* Path to read and save Apple Intermediate Certificate in DER format; downloaded once from https://www.apple.com/certificateauthority/AppleWWDRCAG3.cer (default "AppleWWDRCAG3.cer")

The Apple Intermediate Certificate is included in the output signed request and must be present. It is automatically downloaded if it doesn't exist at this path.

#### -mdmcsr-certificate string

* Path to MDM CSR certificate in DER or PEM format

The signed certificate provided by Apple from the Apple Developer Portal, the MDM CSR certificate is typically provided in "DER" (ASN.1) format.

#### -mdmcsr-private-key string

* Path to MDM CSR private key in PEM PKCS#1 or PKCS#8 format

The private key that signed the CSR that was uploaded to the Apple Developer Portal to get a signed MDM CSR certificate. Legacy private key PEM armor (encryption) is not supported.

#### -out string

* Output filename of the signed MDM CSR request; "-" for stdout (default "-")

The output filename for the Base-64 encoded Apple XML property list required by the [Apple Push Certificates Portal](https://identity.apple.com/).

#### -root-ca string

* Path to read and save Apple Root CA in DER format; downloaded once from https://www.apple.com/appleca/AppleIncRootCertificate.cer (default "AppleIncRootCertificate.cer")

The Apple Root CA is included in the output signed request and must be present. It is automatically downloaded if it doesn't exist at this path.

#### -v

* Print verbose messages to stderr

`mdmutil mdmcsr-sign` prints what it's doing, as it's doing it.

### Example usage

#### Simple invocation

This is a simple invocation of `mdmutil mdmcsr-sign` which assumes you have all the requirements already and turns on verbose mode:

```bash
./mdmutil-darwin-amd64 mdmcsr-sign \
    -apns-csr apns.csr \
    -mdmcsr-private-key /path/to/mdm.key \
    -mdmcsr-certificate /path/to/mdm.cer \
    -out apns_push.plist.b64.req \
    -v
reading certificate AppleIncRootCertificate.cer
reading certificate AppleWWDRCAG3.cer
reading certificate /path/to/mdm.cer
reading APNs CSR apns.csr
reading MDM CSR private key /path/to/mdm.key
signing
writing apns_push.plist.b64.req
```

Then `apns_push.plist.b64.req` can then be uploaded to the [Apple Push Certificates Portal](https://identity.apple.com/) to generated a signed MDM APNs push certificate.
