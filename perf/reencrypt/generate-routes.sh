#!/usr/bin/env bash

: "${SHARD:="perf"}"
: "${NAMESPACE:="scale"}"
: "${DOMAIN:="ocp411.int.frobware.com"}"

for i in $(seq ${1:-1} ${2:-10}); do
    cat <<-EOF
kind: Route
apiVersion: route.openshift.io/v1
metadata:
  labels:
    type: perf
  name: http-perf-reencrypt-${i}
spec:
  host: http-perf-reencrypt-${i}-${NAMESPACE}.${SHARD}.${DOMAIN}
  port:
    targetPort: https
  tls:
    termination: reencrypt
    certificate: |
      -----BEGIN CERTIFICATE-----
      MIIDezCCAmOgAwIBAgIUIDE86s4g/jNL6iHVCeMKzsLL/rwwDQYJKoZIhvcNAQEL
      BQAwTTELMAkGA1UEBhMCRVMxDzANBgNVBAgMBk1hZHJpZDEPMA0GA1UEBwwGTWFk
      cmlkMRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMB4XDTIxMDEyMDExNDAw
      NFoXDTIxMDIxOTExNDAwNFowTTELMAkGA1UEBhMCRVMxDzANBgNVBAgMBk1hZHJp
      ZDEPMA0GA1UEBwwGTWFkcmlkMRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRk
      MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvb+XB/1moWr0MJZoSRN/
      MIvXyPM0pR0N8Gxfpk9LF0qHLX3I35BkugWZpua85HXmurbZgW12UpplTuA/wack
      YjVsTjWNb/p4cQF/o1MIm5iWWCrZY6+Du52gYb36lHNU3Uv2flsNZRHulMq9Ptov
      I01mtYpr/PujU5r+QN01kiswrC4Z7xip3VTRtn4L8VYXIDapNLw4YinNJXIUd+tS
      YsEYG343E9dPXFbobpHi/6JNjrcs6oe4VKvM8pgeAPTM9dbLJtX/TFLAkPLoeIoT
      R/TmP/y7EbZXdAQL2schIL4Nx0iHaLoLPF9mGL/1D/9skAHr3RRC8eAmKKVoZVee
      qwIDAQABo1MwUTAdBgNVHQ4EFgQU3gxPTXzT++e2GtTEnkHNXWTw2KMwHwYDVR0j
      BBgwFoAU3gxPTXzT++e2GtTEnkHNXWTw2KMwDwYDVR0TAQH/BAUwAwEB/zANBgkq
      hkiG9w0BAQsFAAOCAQEAf+cvaTY02vA1BwW8K8kbWS//F+GFgJhFZ7fi2szVZjBk
      JErJcCr8+lwY4GVytPglhO8zWDfMY+c4LKFchPN1LnVYb7v18xbNb7Smz02SQQDl
      sXgTf/f+Fc+lwVRjtVNegfwqc5wZj2ZqBcXq0UnxIBBzXS9EL6czeOqW4gPy5bPa
      KejQwkkULk6KqKGT2tp71BsMyXCM8fMQkvM6FxHPKNUR/GxxQoH4mTQlOgnqrdAc
      4GUhmQZc3IQyUHTUoxMw4BodsLJBH3kCqQ5dy/O3h3GTNzivZNRpItVI6HHjrN2s
      cZm6iaKNoUXr3bv3ugZwe+1R4ISvw/FXS1pMxiD0YA==
      -----END CERTIFICATE-----
    key: |
      -----BEGIN RSA PRIVATE KEY-----
      MIIEpAIBAAKCAQEAvb+XB/1moWr0MJZoSRN/MIvXyPM0pR0N8Gxfpk9LF0qHLX3I
      35BkugWZpua85HXmurbZgW12UpplTuA/wackYjVsTjWNb/p4cQF/o1MIm5iWWCrZ
      Y6+Du52gYb36lHNU3Uv2flsNZRHulMq9PtovI01mtYpr/PujU5r+QN01kiswrC4Z
      7xip3VTRtn4L8VYXIDapNLw4YinNJXIUd+tSYsEYG343E9dPXFbobpHi/6JNjrcs
      6oe4VKvM8pgeAPTM9dbLJtX/TFLAkPLoeIoTR/TmP/y7EbZXdAQL2schIL4Nx0iH
      aLoLPF9mGL/1D/9skAHr3RRC8eAmKKVoZVeeqwIDAQABAoIBAGSKrlZ3ePgzGezc
      5alDAXQRxWcfJ1gOCyLH6e7PuTRAM1xxeAyuEBFZgk8jmBdeOcHZvWqNO9MNKH0g
      6eeMzwSS1i6ixaz+BO+sIZvDFZ6Mva0+Fy5xA9ZX8XGZHrumWONhqtzNFk3lsIt6
      2cgCCFQmYTP0gr/r/mEAkZSBIi+udRIEIsR7T4iLw+H+MWwnXEbM9yke9Qvvvzyt
      ULLWzJT9TrmW7teAvtKYx7D0dyJEXnM8zEeuDd+unzeV2uqrfptU34fjvuQLaUKk
      Br02YeOvs7rSSkGReZNnt8ZXWYFssYs3Kf8s/5EkXYFPukK+8LjII9L8zFgviBMf
      bIqnIYECgYEA6Ns1gkkWtnBmfzSGrFkM5ztQvuLwgaMRLnzsY7JQBnTjWFXTRAqu
      pBtTaxPGEg0P1A4fRnjiSjV1BGzS5yUdyZSYTbc9jiILQwdayB97chQAE+hnVokU
      1zPqtCfiMJETOnV8fRNMBaDRHoBoe1l6va7G37P9NzNgWkpAj+MNKssCgYEA0JuL
      b9IEk0BhOBaV0ROsMJnErsDNbESY4kVotVssZYL1U61jDfC7KbQMmW3frPtGWJi8
      HW99ulXpGvRmwrOhRtPaXzSBwpF8KTseAaRJmPmtJAFYhYRfMFTZWE2JKNGOhkws
      olO6FKkEfgR+m/HtgIqeOdQ/nI1q192SbAHff6ECgYEA4sqN5ST2gB4dVgt8l2Ps
      E1JMJH63rCt8UoDNY5SKKJ+zxZdhusWErsUGjCWoJnCeV/ShNWwLSieinvq2tvYJ
      ewnFBPxRcZtqyI/jNUKkYslkAf+6lifRKoCgOXMW9CJ4Tdmbs94Vju3AfyqlmG3g
      A9q0S7DsENVzJL1pADst2d0CgYEAxRALOsj1FX2d2XRMdsPUx9ya1lLAO+TZX/cd
      oSTN3d9GjZOfnU2qIQ07Ub1frXN50rwGCPCHnv0FRjdW09sJIXWENqfNZNY2qmR0
      RizCccZ67yZuT0LrASdGYopsZakAsJFJINdjU50O51Srnfl+2Q0Zx5tftC5Lnjxr
      06g5T8ECgYAIqSiZU3HluraETpV4OD3XtOTnuEkCA/cd1msPCWsszgvA8Ij0PYjY
      Wl4wlnBAz0mrZbdbJXZ5SqXRiwA+M/bckrODU5ZFiT8Fk5mJlvROqW5rLFIpnJKZ
      cKYMl5I6JC6SQHcbCQ743Yohay5TyGmLFCMnCktx/lHXjZTk8ITY7g==
      -----END RSA PRIVATE KEY-----
  to:
    kind: Service
    name: http-perf-${i}
EOF
done
