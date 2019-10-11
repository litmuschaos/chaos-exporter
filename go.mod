module github.com/litmuschaos/chaos-exporter

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.0

go 1.12

require (
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/litmuschaos/chaos-operator v0.0.0-20191004175208-654d76a51a46
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.0
	github.com/prometheus/client_golang v1.1.0
	github.com/sirupsen/logrus v1.4.2 // indirect
	golang.org/x/crypto v0.0.0-20191010185427-af544f31c8ac // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	sigs.k8s.io/controller-runtime v0.3.0 // indirect
)
