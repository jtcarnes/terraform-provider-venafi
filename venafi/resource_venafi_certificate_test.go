package venafi

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	r "github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"strings"
	"testing"
	"time"
)

func TestDevSignedCert(t *testing.T) {
	t.Log("Testing Dev ECDSA certificate")
	data := testData{}
	data.cn = "dev-random.venafi.example.com"
	data.dns_ns = "dev-web01-random.example.com"
	config := fmt.Sprintf(`
            provider "venafi" {
              alias = "dev"
              dev_mode = true
            }
			resource "venafi_certificate" "dev_certificate" {
            provider = "venafi.dev"
            common_name = "%s"
            algorithm = "RSA"
            rsa_bits = "2048"
            san_dns = [
              "%s"
            ]
            san_ip = [
              "10.1.1.1",
              "192.168.0.1"
            ]
            san_email = [
              "dev@venafi.com",
              "dev2@venafi.com"
            ]
          }
          output "certificate" {
			  value = "${venafi_certificate.dev_certificate.certificate}"
          }
          output "private_key" {
            value = "${venafi_certificate.dev_certificate.private_key_pem}"
          }
                `, data.cn, data.dns_ns)
	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: config,
				Check: func(s *terraform.State) error {
					err := checkStandartCert(t, &data, s)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	})
}

func TestDevSignedCertECDSA(t *testing.T) {
	t.Log("Testing Dev ECDSA certificate")
	data := testData{}
	data.cn = "dev-random.venafi.example.com"
	data.private_key_password = "123xxx"
	config := fmt.Sprintf(`
	provider "venafi" {
		alias = "dev"
		dev_mode = true
	}
	resource "venafi_certificate" "dev_certificate" {
		provider = "venafi.dev"
		common_name = "%s"
		algorithm = "ECDSA"
		key_password = "%s"
	}
	output "certificate" {
		value = "${venafi_certificate.dev_certificate.certificate}"
	}
	output "private_key" {
		value = "${venafi_certificate.dev_certificate.private_key_pem}"
	}`, data.cn, data.private_key_password)
	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: config,
				Check: func(s *terraform.State) error {
					err := checkStandartCert(t, &data, s)
					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	})
}

func TestCloudSignedCert(t *testing.T) {
	t.Log("Testing Cloud certificate")
	data := testData{}
	rand := randSeq(9)
	domain := "venafi.example.com"
	data.cn = rand + "." + domain
	data.private_key_password = "123xxx"
	config := fmt.Sprintf(`
            variable "CLOUDURL" {}
            variable "CLOUDAPIKEY" {}
            variable "CLOUDZONE" {}
            provider "venafi" {
              alias = "cloud"
              url = "${var.CLOUDURL}"
              api_key = "${var.CLOUDAPIKEY}"
              zone = "${var.CLOUDZONE}"
            }
			resource "venafi_certificate" "cloud_certificate" {
            provider = "venafi.cloud"
            common_name = "%s"
            algorithm = "RSA"
            rsa_bits = "2048"
			key_password = "%s"
          }
          output "certificate" {
			  value = "${venafi_certificate.cloud_certificate.certificate}"
          }
          output "private_key" {
            value = "${venafi_certificate.cloud_certificate.private_key_pem}"
          }
                `, data.cn, data.private_key_password)
	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: config,
				Check: func(s *terraform.State) error {
					err := checkStandartCert(t, &data, s)
					if err != nil {
						return err
					}
					return nil

				},
			},
		},
	})
}

func TestTPPSignedCert(t *testing.T) {
	t.Log("Testing TPP certificate")
	data := testData{}
	rand := randSeq(9)
	domain := "venafi.example.com"
	data.cn = rand + "." + domain
	data.dns_ns = "alt-" + data.cn
	data.dns_ip = "192.168.1.1"
	data.dns_email = "venafi@example.com"
	data.private_key_password = "123xxx"
	config := fmt.Sprintf(`
            variable "TPPUSER" {}
            variable "TPPPASSWORD" {}
            variable "TPPURL" {}
            variable "TPPZONE" {}
			variable "TRUST_BUNDLE" {}
            provider "venafi" {
              alias = "tpp"
              url = "${var.TPPURL}"
              tpp_username = "${var.TPPUSER}"
              tpp_password = "${var.TPPPASSWORD}"
              zone = "${var.TPPZONE}"
              trust_bundle = "${file(var.TRUST_BUNDLE)}"
            }
			resource "venafi_certificate" "tpp_certificate" {
            provider = "venafi.tpp"
            common_name = "%s"
            san_dns = [
              "%s"
            ]
            san_ip = [
              "%s"
            ]
            san_email = [
              "%s"
            ]
            algorithm = "RSA"
            rsa_bits = "2048"
			key_password = "%s"
          }
          output "certificate" {
			  value = "${venafi_certificate.tpp_certificate.certificate}"
          }
          output "private_key" {
            value = "${venafi_certificate.tpp_certificate.private_key_pem}"
          }`, data.cn, data.dns_ns, data.dns_ip, data.dns_email, data.private_key_password)

	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: config,
				Check: func(s *terraform.State) error {
					return checkStandartCert(t, &data, s)
				},
			},
			r.TestStep{
				Config: config,
				Check: func(s *terraform.State) error {
					t.Log("Testing update")
					gotSerial := data.serial
					gotTime := data.timeCheck
					err := checkStandartCert(t, &data, s)
					if err != nil {
						return err
					} else {
						t.Logf("Compare updated certificate serial at time %s against original at time %s", data.timeCheck, gotTime)
						if gotSerial == data.serial {
							return fmt.Errorf("serial number from updated certificate %s is the same as in original number %s", data.serial, gotSerial)
						} else {
							return nil
						}
					}
				},
			},
		},
	})
}

func TestTPPECDSASignedCert(t *testing.T) {
	t.Log("Testing TPP certificate")
	data := testData{}
	rand := randSeq(9)
	domain := "venafi.example.com"
	data.cn = rand + "." + domain
	data.dns_ns = "alt-" + data.cn
	data.dns_ip = "192.168.1.1"
	data.dns_email = "venafi@example.com"
	data.private_key_password = "123xxx"
	config := fmt.Sprintf(`
            variable "TPPUSER" {}
            variable "TPPPASSWORD" {}
            variable "TPPURL" {}
            variable "TPPZONE" {}
			variable "TRUST_BUNDLE" {}
            provider "venafi" {
              alias = "tpp"
              url = "${var.TPPURL}"
              tpp_username = "${var.TPPUSER}"
              tpp_password = "${var.TPPPASSWORD}"
              zone = "${var.TPPZONE}"
              trust_bundle = "${file(var.TRUST_BUNDLE)}"
            }
			resource "venafi_certificate" "tpp_certificate" {
            provider = "venafi.tpp"
            common_name = "%s"
            san_dns = [
              "%s"
            ]
            san_ip = [
              "%s"
            ]
            san_email = [
              "%s"
            ]
            algorithm = "ECDSA"
            ecdsa_curve = "P521"
			key_password = "%s"
          }
          output "certificate" {
			  value = "${venafi_certificate.tpp_certificate.certificate}"
          }
          output "private_key" {
            value = "${venafi_certificate.tpp_certificate.private_key_pem}"
          }`, data.cn, data.dns_ns, data.dns_ip, data.dns_email, data.private_key_password)

	r.Test(t, r.TestCase{
		Providers: testProviders,
		Steps: []r.TestStep{
			r.TestStep{
				Config: config,
				Check: func(s *terraform.State) error {
					return checkStandartCert(t, &data, s)
				},
			},
			r.TestStep{
				Config: config,
				Check: func(s *terraform.State) error {
					t.Log("Testing update")
					gotSerial := data.serial
					gotTime := data.timeCheck
					err := checkStandartCert(t, &data, s)
					if err != nil {
						return err
					} else {
						t.Logf("Compare updated certificate serial at time %s against original at time %s", data.timeCheck, gotTime)
						if gotSerial == data.serial {
							return fmt.Errorf("serial number from updated certificate %s is the same as in original number %s", data.serial, gotSerial)
						} else {
							return nil
						}
					}
				},
			},
		},
	})
}

func checkStandartCert(t *testing.T, data *testData, s *terraform.State) error {
	t.Log("Testing certificate with cn", data.cn)
	certUntyped := s.RootModule().Outputs["certificate"].Value
	certificate, ok := certUntyped.(string)
	if !ok {
		return fmt.Errorf("output for \"certificate\" is not a string")
	}

	t.Logf("Testing certificate PEM:\n %s", certificate)
	if !strings.HasPrefix(certificate, "-----BEGIN CERTIFICATE----") {
		return fmt.Errorf("key is missing cert PEM preamble")
	}
	block, _ := pem.Decode([]byte(certificate))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("error parsing cert: %s", err)
	}
	if expected, got := data.cn, cert.Subject.CommonName; got != expected {
		return fmt.Errorf("incorrect subject common name: expected %v, certificate %v", expected, got)
	}
	if len(data.dns_ns) > 0 {
		if expected, got := []string{data.cn, data.dns_ns}, cert.DNSNames; !sameStringSlice(got, expected) {
			return fmt.Errorf("incorrect DNSNames: expected %v, certificate %v", expected, got)
		}
	} else {
		if expected, got := []string{data.cn}, cert.DNSNames; !sameStringSlice(got, expected) {
			return fmt.Errorf("incorrect DNSNames: expected %v, certificate %v", expected, got)
		}
	}

	data.serial = cert.SerialNumber.String()
	data.timeCheck = time.Now().String()

	keyUntyped := s.RootModule().Outputs["private_key"].Value
	privateKey, ok := keyUntyped.(string)
	if !ok {
		return fmt.Errorf("output for \"private_key\" is not a string")
	}

	t.Logf("Testing private key PEM:\n %s", privateKey)
	privKeyPEM, err := getPrivateKey([]byte(privateKey), data.private_key_password)

	_, err = tls.X509KeyPair([]byte(certificate), privKeyPEM)
	if err != nil {
		return fmt.Errorf("error comparing certificate and key: %s", err)
	}

	return nil
}
