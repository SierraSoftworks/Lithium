# Lithium [![Build Status](https://travis-ci.org/SierraSoftworks/Lithium.svg?branch=master)](https://travis-ci.org/SierraSoftworks/Lithium)
**Secure Asymmetric Licensing Protocol**

Lithium is a licensing protocol which provides the ability to provide time locked, floating
and leased licensing both over the internet and through an intranet server.

The primary design considerations with Lithium are that it should support licensing of a
wide range of applications with varying requirements, it should enable clients to verify the
integrity of a license without needing to communicate with a license server and should cater
to both node-locked and floating license use cases.

Beyond that, it should be hardened against passive side-channel attacks, should be easily
implemented on different platforms and should offer a reasonable degree of safety for general
use cases. Active attacks such as binary manipulation are not covered under the scope of Lithium
and should make use of an alternative strategy to protect distributed binaries against
manipulation.

## Architecture

### License Data
Licenses are constructed out of a structured JSON object containing license metadata as well
as a customizable payload segment. The payload segment is intended to be customized by the
license consumer for toggling the availability of functionality and configuring options in
the software.

```json
{
    "meta": {
        "id": "8ddcb55cd6a341b48c4aa7b611e1721b",
        
        "activates": "2016-02-14T06:08:12.012Z",
        "expires": "2016-02-14T08:10:12.014Z"
    },
    
    "payload": {
        "feature1": true
    }
}
```

### License Validation
Licenses are signed using asymmetric cryptographic signatures. These signatures are generated
using the private key of the server generating the license, allowing the validity to be confirmed
by checking the license's signature against the public key of that server. This ensures that
the license originated from the specified server, while the server's published certificate
chain enables clients to verify that the server's license originated from a trusted license
source.

The server's certificate (public key) may be signed for a limited period of time, in which
case licenses it generates will only remain valid for, at most, its validity period. Servers
may also opt to only sign licenses for a portion of the server's validity period - enabling
use cases like floating licenses or time-limited offline usage.

![Signing Hierarchy](resources/signing_hierarchy.png)

### License Protection
License files are also encrypted to prevent them from being readable by anybody but the
target machine. This is achieved by encrypting the license pack using the machine's public
key, which is sent with the license request.

Locally, the machine's private key may be encrypted using machine specific data like processor
codes, MAC addresses etc. Decisions about how the private key is protected fall outside the
purview of this specification and will depend on the platform you are targetting.

## License Packing
Licenses are packed using a sequence of PEM blocks. These blocks include the `LITHIUM LICENSE KEY`,
`LITHIUM LICENSE`, `LITHIUM SIGNATURE` and `LITHIUM CERTIFICATE` blocks. With the exception of the
`LITHIUM CERTIFICATE` blocks, ordering is not important.

```
-----BEGIN LITHIUM LICENSE KEY-----
6Ja6TUrx0euTsau/tjUoTfZm4QF9lC4uDLoWg6RoIZYmSF1zxjbmwysaxmYy13Dd
BrdNNz3f97TD5/t5MaEM56bGtpQj0dqgutP57Ue/kTqNADilcEgSvvR9LGhtNH4z
JGzbWDU3+TEWDyqUpIoxMHOitvg53u3O2aw99BpokeA=
-----END LITHIUM LICENSE KEY-----
-----BEGIN LITHIUM LICENSE-----
algorithm: aes256
iv: BAdz8wzMhcbiJyXpkaaXIQ==
9MAhZiUgECGOHgD9Lg2wJg5+grzdY9GsCctT6DYUMiUjicChIoov2aCAZevBXBXa
wiRpkmn0Bqgt525ZXgFakL1FlNFWgqnLR79Ij//J3aqU/TjP2oKoblhYLeHTY3vu
79uyIQvIyn6WRn3GoO3hVYFVPlWRXK2R9wd7pIj4yCGX1Ug=
-----END LITHIUM LICENSE-----
-----BEGIN LITHIUM SIGNATURE-----
algorithm: sha256
gFVD3kuzTLQYjLTvf6d3jyBZe9SFLz5Le4JLsCiVhd3CCwrnivWYOMwwtsdVJO+N
szVGbLOY6mttXNsP4So+Ucfy4xI0T3Gvz+afeiNCPnAZ2m0rOuqoxC+31vWMnuWt
JLKSMz587E06qHEk7I6/AuKWJBtFn0umQvhSNktr/ME=
-----END LITHIUM SIGNATURE-----
-----BEGIN LITHIUM CERTIFICATE-----
MIICsjCCAhugAwIBAgIBATANBgkqhkiG9w0BAQsFADB2MRkwFwYDVQQKExBTaWVy
cmEgU29mdHdvcmtzMRowGAYDVQQLExFMaXRoaXVtIExpY2Vuc2luZzEiMCAGA1UE
AxMZTGl0aGl1bSBUZXN0aW5nICh0ZXN0aW5nKTEZMBcGA1UEBRMQUm9vdCBDZXJ0
aWZpY2F0ZTAgFw0xNjA1MDMxNDE0MzRaGA8yMTE2MDQwOTE0MTQzNFowdjEZMBcG
A1UEChMQU2llcnJhIFNvZnR3b3JrczEaMBgGA1UECxMRTGl0aGl1bSBMaWNlbnNp
bmcxIjAgBgNVBAMTGUxpdGhpdW0gVGVzdGluZyAodGVzdGluZykxGTAXBgNVBAUT
EFJvb3QgQ2VydGlmaWNhdGUwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAJ+d
q1UwLUQaAm3rxC94PHiJkxwkgH1JSOE9NvFKYPOjGXVqAxpUUorvTkQLXnVS9tcB
Oe5kOKpsppluX7r8bkf78jAI2Pm447EtnMvHJKfRJqRqd3zmoL2goMIV21j9WuWj
Pjl6HPGsAvfdfSAvGZ18Huf4lEpj37HUL2gaPLKPAgMBAAGjTjBMMA4GA1UdDwEB
/wQEAwIB9jAPBgNVHSUECDAGBgRVHSUAMBMGA1UdEwEB/wQJMAcBAf8CAgCAMBQG
A1UdEQQNMAuCCWxvY2FsaG9zdDANBgkqhkiG9w0BAQsFAAOBgQAKjLSIOy1AFTNl
NessCe8ETzw3M2+gyH/IO7e2DFOYsflkkcaNERM9OA64afEQOgW+Klo9BGjYAB67
aZPcbxPanPD9ULUioF593phV4DufyW+qmfhWnAzQjHQpv8r2U8KXJdR02v7xEOg4
5lw1MYJgxleG6haeb6S3FUaEyi9hMA==
-----END LITHIUM CERTIFICATE-----
-----BEGIN LITHIUM CERTIFICATE-----
MIIClDCCAf2gAwIBAgIFAJW/VocwDQYJKoZIhvcNAQELBQAwdjEZMBcGA1UEChMQ
U2llcnJhIFNvZnR3b3JrczEaMBgGA1UECxMRTGl0aGl1bSBMaWNlbnNpbmcxIjAg
BgNVBAMTGUxpdGhpdW0gVGVzdGluZyAodGVzdGluZykxGTAXBgNVBAUTEFJvb3Qg
Q2VydGlmaWNhdGUwHhcNNzAwMTAxMDAwMDAwWhcNNzAwMTAxMDAwMDAwWjBqMRkw
FwYDVQQKExBTaWVycmEgU29mdHdvcmtzMRowGAYDVQQLExFMaXRoaXVtIExpY2Vu
c2luZzEiMCAGA1UEAxMZTGl0aGl1bSBUZXN0aW5nICh0ZXN0aW5nKTENMAsGA1UE
BRMEdGVzdDCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEA8mFWK7JkzR/ELQmQ
3m3OjxOBmf8YcvK0wdJ+w1DzN+upKS6/E5sswRQGDf5YENHzA/WiesTyzHtVWTzh
qTlP4+KXyAFUcogjsf00DQPs/TUBOvlHzXxCVF9qQrdlmz1p9TXIbOED+m7wisrK
peJpBxEXAzGONBlxMEyEAqtkuC8CAwEAAaM6MDgwDgYDVR0PAQH/BAQDAgSQMBAG
A1UdEwEB/wQGMAQCAgCAMBQGA1UdEQQNMAuCCWxvY2FsaG9zdDANBgkqhkiG9w0B
AQsFAAOBgQCIPQ9hrI2AiHAYBsNbnpA1iN6hlx1DWCN6llAs7QwjPp8bR1hW7Yjn
A79u+KzYSLuKSd6lP3leaeNKkNl3JOrPeHeaJLMYL5TY09MwlID33+40/y3XWK8/
9aGgpXdxJ7IQv1MBZJZH2tELb4zyJzAquzeg54SuTvYdDm9MGBzbtw==
-----END LITHIUM CERTIFICATE-----
```

### License Key
The license key block comprises an encrypted key for the decryption of the License block.
This key is encrypted using an asymmetric encryption scheme like RSA and should only be
decryptable by the node which will be making use of the license. This ensures that a license
may not be used across multiple nodes while also making the license contents opaque to
3rd parties.

### License
The license field contains the structured license object in encrypted form. It should be
decrypted using the key sourced from the `LITHIUM LICENSE KEY` section, yielding a 
structured JSON object as defined by the `license.data.schema.json` schema.

The `iv` and `algorithm` headers exist to enable interoperability across various
implementations. `iv` is encoded using standard Base64 encoding and is generated
for each license encryption operation, while the `algorithm` will likely be `aes256`
in most cases.

Once the license has been confirmed to originate from a trusted source,
the application is responsible for confirming that the current time rests between
the `activates` and `expires` times. If it does not, the license is considered to have
expired and the user should be informed.

### Signature
The signature field represents the asymmetric cryptographic signature of the raw data
within the `LITHIUM LICENSE` field (in encrypted form). By analyzing the signature it is
therefore possible to determine whether the data has been tampered with.

If the signature does not match the expected signature for the data, given the public key
available in the signature chain, then the certificate is considered tampered with and
invalid. The application should inform the user to this effect.

### Certificate
Each instance of the `LITHIUM CERTIFICATE` block represents a DER encoded x509 certificate.
The ordering of these blocks is important, as the first block is expected to match the
certificate embedded within the client application (a trusted root) while each subsequent
certificate is expected to be signed by the previous certificate. In this way, it is
possible to trace the authenticity of any certificate back to the root certificate.

Similarly, the final certificate is expected to be the certificate used to generate the
`signature` for the license.

If any of these conditions is not met, the license is determined to be invalid and the
application should inform the user to this effect.

## Implementation Details

### Clock Skew
Due to the nature of distributed systems, it is likely that the local machine's clock will
have an offset from the license server's clock. When making a license request, the server
should accept a reasonable clock skew parameter (a couple of minutes at most) and generate
a license which caters for the target machine's clock time.

Alternative, though illadvised, approaches include storing a clock offset on the local machine
as well as updating the local machine's time to match the license server's time. The former
opens an attack surface, namely that the file may be modified to inject any desired offset
and thereby extend a license indefinitely; the latter may not be possible across different
platforms or may pose problems for certain classes of user or software.

### Floating Licenses
Floating licenses, specifically those which work on a "seats" basis, are intended to be
implemented through the use of continually renewed, short-lived licenses. These would be
set to expire after between 60 seconds and 5 minutes, depending on the application. The
application would request a license renewal once its local time approached a certain threshold
from the current license's expiry time, for example - 30 seconds from expiry.

On the server side, licenses will only be generated when there are available license tokens
in the pool or when a renewal request is received for a valid license. In the case of a renewal
request, the expiry time for the relevant license (identified by its `id` is updated and sent
to the requesting client). It is important to note that renewed licenses should only be generated
for the client who previously checked out the license so as to prevent multiple nodes using the
same license. 

Licenses are returned to the pool when they expire, ensuring that offline users do not hold
licenses which they are unable to use.