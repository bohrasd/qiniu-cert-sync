package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/BurntSushi/toml"
	"github.com/bohrasd/qiniu-cert-sync/qiniu"
)

var (
	config Configs
	qm     qiniu.QiniuManager
)

type Configs struct {
	Mode    string
	QnAuth  QiniuAuth `toml:"qn_auth"`
	Secrets []SyncSecret
}

type QiniuAuth struct {
	QiniuAK string `toml:"qiniu_access_key"`
	QiniuSK string `toml:"qiniu_secret_key"`
}

type SyncSecret struct {
	Secret     string
	Namespace  string
	CommonName []string `toml:"common_name"`
}

func main() {

	//获取配置
	configFile := flag.String("config", "/etc/qiniu-cert-sync/config.toml", "absolute path to the config file")
	flag.Parse()
	file, err := os.Open(*configFile)
	if err != nil {
		panic(err)
	}
	configData, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	_, err = toml.Decode(string(configData), &config)

	//初始化七牛
	qm = qiniu.NewQiniuManager(config.QnAuth.QiniuAK, config.QnAuth.QiniuSK)

	//初始化 k8s
	var clientConfig *rest.Config
	switch config.Mode {
	case "local":
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			fmt.Println(home)
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
		clientConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err)
		}

	case "cluster":
		fallthrough
	default:
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	}

	clientset, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	//开始同步证书
	qiniuCertList := getQiniuCertList()

	var wg sync.WaitGroup
secret:
	for _, secret := range config.Secrets {
		if secret.Secret == "" {
			continue
		}

		secretsClient := clientset.CoreV1().Secrets(secret.Namespace)
		k8sCertSecret, err := secretsClient.Get(context.TODO(), secret.Secret, metav1.GetOptions{})
		if err != nil {
			log.Printf("Can't access secret %s , error %s\n", secret.Secret, err.Error())
			continue
		}
		certName := strings.Join([]string{secret.Secret, secret.Namespace, k8sCertSecret.ResourceVersion}, "-")
		k8sCertPub, err := parseCertificate(k8sCertSecret)
		if err != nil {
			log.Printf("Certificate %s can't be parsed, error %s\n", certName, err.Error())
			continue
		}
		for _, qnCert := range qiniuCertList.Certs {

			if qnCert.Name == certName {
				if time.Unix(int64(qnCert.NotAfter), 0).Before(k8sCertPub.NotAfter) {
					fmt.Printf("old cert %s going to be deleted\n", qnCert.CertID)

					defer qm.DelCert(qnCert.CertID)
				}

				if time.Unix(int64(qnCert.NotAfter), 0).Equal(k8sCertPub.NotAfter) {

					fmt.Printf("same cert %s already exists\n", certName)
					wg.Add(1)
					go domainSet(secret.CommonName, qnCert.CertID, &wg)

					continue secret
				}

			}
		}

		certUpResp, err := qm.UploadCert(certName, k8sCertPub.Subject.CommonName, string(k8sCertSecret.Data["tls.crt"]), string(k8sCertSecret.Data["tls.key"]))
		if err != nil {
			log.Printf("Certificate %s can't upload to Qiniu, error %s\n", certName, err.Error())
			continue
		}

		fmt.Printf("Certificate %s uploaded, Cert ID: %s\n", certName, certUpResp.CertID)
		wg.Add(1)
		go domainSet(secret.CommonName, certUpResp.CertID, &wg)

	}

	wg.Wait()
}

//为域名修改证书
func domainSet(commonNames []string, certID string, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, cn := range commonNames {
		domain, err := qm.GetDomain(cn)
		if err != nil {
			log.Printf("Domain name %s not found on qiniu, error %s\n", cn, err.Error())
			continue
		}

		if domain.CertID != certID {
			err = qm.ChangeDomainCert(cn, certID)
			if err != nil {
				log.Printf("Certificate of domain %s change failed, error: %s\n", cn, err.Error())
				continue
			}
			log.Printf("Updated certificate of %s with: %s\n", domain.Name, certID)
		}
	}
}

//获取七牛证书列表
func getQiniuCertList() qiniu.CertListResp {

	qiniuCertList, err := qm.CertList()
	if err != nil {
		panic(err)
	}
	return qiniuCertList
}

//解析证书
func parseCertificate(k8sCertSecret *v1.Secret) (k8sCertPub *x509.Certificate, err error) {
	block, _ := pem.Decode(k8sCertSecret.Data["tls.crt"])
	if block == nil || block.Type != "CERTIFICATE" {
		log.Fatal("failed to decode PEM block containing public key")
	}

	k8sCertPub, err = x509.ParseCertificate(block.Bytes)
	return
}
