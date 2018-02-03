all:: $(GOPATH)/pkg/darwin_amd64/github.com/boynton/ion-go/ion/ion.a

test::
	go test 

$(GOPATH)/bin/ion2rdl: ion2rdl/main.go $(GOPATH)/pkg/darwin_amd64/github.com/boynton/ion-go/ion/ion.a
	go install github.com/boynton/ion-go/ion2rdl

$(GOPATH)/pkg/darwin_amd64/github.com/boynton/ion-go/ion/ion.a: ion/parser.go
	go install github.com/boynton/ion-go/ion

clean::
	rm -f $(GOPATH)/bin/ion2rdl
	rm -rf $(GOPATH)/pkg/darwin_amd64/github.com/boynton/ion-go/ion/ion.a
