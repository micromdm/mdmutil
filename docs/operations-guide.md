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

#### Full MDM CSR and APNs push certificate issuance

Since the MDM CSR needs to be renewed yearly, let's walk through the complete end-to-end process for both the MDM CSR and APNs push certificate issuance. See also the [MicroMDM "Understanding MDM Certificates" blog post](https://micromdm.io/blog/certificates/) for a conceptual overview of this process.

You'll need an [Apple Developer Account](https://developer.apple.com/programs/) with the "MDM CSR" certificate option enabled.

##### 1. Generate the MDM CSR ... CSR and private key

```bash
openssl req -out mdmcsr.csr -newkey rsa:2048 -keyout mdmcsr.key -nodes -subj "/C=US/CN=MDM CSR/emailAddress=mdm-admin@example.com"
```

Be sure to replace any values in the template you'd like (e.g. the email address).

##### 2. Get the MDM CSR certificate signed by Apple

Upload the `mdmcsr.csr` to Apple Developer ["Certificates, Identifiers & Profiles" portal](https://developer.apple.com/account/resources/certificates/list). For reference [this older MicroMDM talk video walks through this process](https://www.youtube.com/watch?v=WGKT-PyHz6I&t=2152s) around the 35:52 mark.

1. From the "Certificates, Identifiers & Profiles" portal add a new certificate (click the blue "+" button).
2. Select the "MDM CSR" option and click the "Continue" button. See then note below if you do not have this option.
3. Upload the `mdmcsr.csr` file you generated above to the portal page and "Continue"
4. You should be offered a screen to download the signed certificate. It'll usually be called just `mdm.cer`.

You now have an Apple-signed MDM CSR. Please keep the signed `mdm.cer` and `mdmcsr.key` around as you'll need them to sign APNs push certificate requests.

##### 3. Generate the APNs Push CSR

```bash
openssl req -out push.csr -newkey rsa:2048 -keyout push.key -nodes -subj "/C=US/CN=APNs Push/emailAddress=mdm-admin@example.com"
```

Be sure to replace any values in the template you'd like (e.g. the email address or the `CN` describing which APNs push certificate).

##### 4. Sign the APNs Push CSR request

Use `mdmutil mdmcsr-sign` command which uses the MDM CSR private key and Apple-signed certificate to sign the APNs push certificate request:

```bash
./mdmutil-darwin-amd64 mdmcsr-sign -apns-csr push.csr -mdmcsr-private-key mdmcsr.key -mdmcsr-certificate mdm.cer -out push.plist.b64.req
```

If everything went smoothly, there should be no output and a new file: `push.plist.b64.req`.

##### 5. Get the APNs Push certificate signed by Apple

Upload the `push.plist.b64.req` Push Certificate request to the [Apple Push Certificates Portal](https://identity.apple.com/pushcert/).

1. Login to the [Apple Push Certificates Portal](https://identity.apple.com/)
2. **If you're renewing a certificate you must use the blue "Renew" button** for a previously issued push certificate. Failure to follow this step will require devices to be re-enrolled in MDM as the certificate "Topic" (`UserID` attribute) will be different.
3. If you're creating a new APNs certificate use the green "Create a Certificate" button.
4. Select the `push.plist.b64.req` for upload and optionally include a note for this APNs push certificate.
5. Download the certificate. It is typically named `MDM_<account>_Certificate.pem` with the account replaced with the individual or business name of the Apple Developer Account.

There you have it, you've had Apple mint you an APNs Push certificate. The `MDM_<account>_Certificate.pem` together with `push.key` should be a valid APNs Push certificate and private key pair for sending MDM APNs push notifications to devices once they've enrolled in an MDM server.

#### Tips and FYIs

- While the "MDM CSR" option used to require an [Enterprise](https://developer.apple.com/programs/enterprise/) Developer Account, a standard developer account now works just fine.
- Only the Apple Developer [Account Holder](https://developer.apple.com/help/account/access/roles/) is able to get the "MDM CSR" option for signing.
- If the Apple Developer Account Account Holder does not have the "MDM CSR" option in your "Certificates, Identifiers & Profiles" portal in your Apple Developer Account you'll need to [request it from Developer Support](https://developer.apple.com/contact/). See [this MacAdmins.org Slack thread](https://macadmins.slack.com/archives/C19RTE0L9/p1760296585318419?thread_ts=1758700336.136079&cid=C19RTE0L9) for more information and guidance on submitting this request.
- Both the MDM CSR certificate and the APNs push certificate have **yearly expiries**.
- As noted you must *Renew* (in the APNs push portal) the certificate or else **you'll need to re-enroll all your devices**. This implies careful control of the Apple ID used for the Apple Push Certificates Portal.
- The Apple ID for the Apple Push Certificates Portal need not be linked with the Apple Developer Account nor Account Holder. The original intent was for an MDM vendor legal entity to control the Apple Developer Account while an MDM customer's legal entity controlled the Apple ID for the Push Certificate portal. Of course they may be very well be the same entity for self-hosted and open source MDM solutions, but this does not imply they need to be the same Apple ID or email address. If the account is shared beware of 2fac phone number management and related concerns.
- It is possible to work with Apple to migrate Apple IDs for Push Certificates. See [Rich Trouton's blog post on APNs cert migration](https://derflounder.wordpress.com/2023/04/11/migrating-an-apns-certificate-from-one-apple-id-to-another-apple-id/) and [Apple's documentation](https://support.apple.com/en-us/118629).
