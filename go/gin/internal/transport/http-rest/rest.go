package rest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"pjt/internal/config"
	"pjt/internal/logger"
	"pjt/internal/transport/http-rest/controller"
	"time"
)

type RESTServer struct {
	server *http.Server
}

func NewRESTServer(controller controller.Controller, config *config.Configuration) *RESTServer {
	server := &http.Server{
		Addr: config.RestIp + ":" + config.RestPort,
		// if using sse, comment below
		// WriteTimeout: time.Second * 5,
		ReadTimeout: time.Second * 5,
		Handler:     controller.Router,
		//TLSConfig:   initHttpTLSconfig(),
	}

	return &RESTServer{
		server: server,
	}
}

func (r *RESTServer) Start() error {
	// return r.server.ListenAndServeTLS("filpath.crt", "filepath.key")
	return r.server.ListenAndServe()
}

func (r *RESTServer) Shutdown(ctx context.Context) (err error) {
	err = r.server.Shutdown(ctx)
	logger.Infoln("[REST] REST goroutine terminated")
	return err
}

func initHttpTLSconfig() *tls.Config {
	insecure := flag.Bool("[REST] insecure-ssl", false, "Accept/Ignore all server SSL certificates")
	flag.Parse()

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	//certs, err := ioutil.ReadFile(localCertFile)
	certs, err := ioutil.ReadFile("rootca.pem")
	if err != nil {
		log.Fatalf("[REST] Failed to read RootCAs: %v", err)
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		log.Println("[REST] Failed to append RootCAs")
	}

	// Trust the augmented cert pool in our client
	config := &tls.Config{
		InsecureSkipVerify: *insecure,
		RootCAs:            rootCAs,
	}

	return config
}

// make self signed crt example
/**
 $ openssl genrsa -des3 -out shevpca.key 2048
password : 123456

$ openssl req -x509 -new -nodes -key rtca.key -sha256 -days 18250 -out rtca.pem
Enter pass phrase for shevpca.key: 123456
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:KR
State or Province Name (full name) [Some-State]:rtca
Locality Name (eg, city) []:rtca
Organization Name (eg, company) [Internet Widgits Pty Ltd]:
Organizational Unit Name (eg, section) []:
Common Name (e.g. server FQDN or YOUR name) []:rtca.go.kr
Email Address []:rtca.go.kr


$ openssl x509 -outform der -in rtca.pem -out rtca.crt

여기까지 rootca생성 완료, 사용자 pc에 인증 rtca.crt를 카피하고 설치한다.

$ openssl genrsa -out domain.go.kr.key 2048

$ openssl req -new -key domain.go.kr.key -out domain.go.kr.csr
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [AU]:KR
State or Province Name (full name) [Some-State]:domain
Locality Name (eg, city) []:domain
Organization Name (eg, company) [Internet Widgits Pty Ltd]:domain
Organizational Unit Name (eg, section) []:
Common Name (e.g. server FQDN or YOUR name) []:domain.go.kr
Email Address []:domain.go.kr

Please enter the following 'extra' attributes
to be sent with your certificate request
A challenge password []:
An optional company name []:


$ vi domain.go.kr.ext
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = domain.go.kr

openssl x509 -req -in domain.go.kr.csr -CA rtca.pem -CAkey rtca.key -CAcreateserial -out domain.go.kr.crt -days 18250 -sha256 -extfile domain.go.kr.ext
Enter pass phrase for uwsigca.key: 123456

우분투 인증서 등록
$ sudo cp *.crt /usr/local/share/ca-certificates/
$ sudo update-ca-certificates
*/
