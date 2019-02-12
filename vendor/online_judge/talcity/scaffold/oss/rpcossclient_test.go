package oss

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/grpc/grpc-go/status"
	"github.com/satori/go.uuid"
)

var (
	pool *OOSClientPool

	address = "10.1.31.105:1805"

	endpoint        = "221.122.128.3"
	accessKey       = "OFHRQRBF1IC07YDYIMFE"
	secretKey       = "TNNExvPDvQrxesXcVREWNhHjtVs9NOgHO6fkcQFW"
	bucket          = "oss-sdk"
	ttl       int64 = 60 * 60 * 24 * 30 * 10 // 10年
)

func init() {
	var err error
	pool, err = NewOOSClientPool(address, 1, 1, time.Second)
	if err != nil {
		log.Fatalf(fmt.Sprintf("new oos client pool failed: %s", err.Error()))
	}
}

func TestPutObject(t *testing.T) {
	defer pool.Close()
	ctx := context.Background()
	client, err := pool.GetOOSClient(ctx)
	if err != nil {
		t.Fatalf("get oss client from pool failed: %s", err.Error())
	}
	defer pool.Put(client)
	bucket := &Bucket{
		endpoint, accessKey, secretKey, bucket,
		true, 5, 5,
	}
	fileName, fileContent := readFile()
	req := &PutObjectReq{
		bucket, fileName, fileContent,
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err = client.PutObject(ctx, req)
	if err != nil {
		covertStatus(err)
	}

	urlReq := &GenerateURLReq{
		bucket, fileName, ttl,
	}
	urlResp, err := client.GenerateURL(ctx, urlReq)
	if err != nil {
		covertStatus(err)
	}

	if urlResp == nil {
		t.Error("get access url is failed")
	}

	t.Logf("access url is %s", urlResp.GetURI())
}

func TestNewOOSClientPool(t *testing.T) {
	defer pool.Close()
	t.Logf("pool stats info: %s", pool.StatsJSON())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := pool.GetOOSClient(ctx)
	if err != nil {
		t.Fatalf("get oss client from pool failed: %s", err.Error())
	}

	pool.Put(client)
}

func TestNewOOSClientPoolTimeout(t *testing.T) {
	ctx := context.Background()
	defer pool.Close()

	c, err := pool.GetOOSClient(ctx)
	if err != nil {
		t.Fatal(err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	_, err = pool.GetOOSClient(timeoutCtx)
	cancel()

	want := "resource pool timed out"
	if err == nil || err.Error() != want {
		t.Errorf("got %v, want %s", err, want)
	}

	pool.Put(c)
}

func TestNewOOSClientPoolExpired(t *testing.T) {
	defer pool.Close()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	c, err := pool.GetOOSClient(ctx)
	if err == nil {
		pool.Put(c)
	}
	cancel()

	want := "resource pool timed out"
	if err == nil || err.Error() != want {
		t.Errorf("got %v, want %s", err, want)
	}
}

func TestNewOOSClientPoolFail(t *testing.T) {
	pool.Close()
	ctx := context.Background()
	p, err := NewOOSClientPool(address, 3, 4, 1*time.Millisecond)
	if err != nil {
		t.Fatalf(fmt.Sprintf("new oos client pool failed: %s", err.Error()))
	}

	defer p.Close()

	ch := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			c, _ := p.GetOOSClient(ctx)
			p.Put(c)
			println("get oss client ch is true")
			ch <- true
		}()
	}

	for i := 0; i < 3; i++ {
		println("read ch value", <-ch)
	}

	time.Sleep(3 * time.Second)

	println(p.StatsJSON())
	println(p.Available())
	if p.Available() != 3 {
		t.Errorf("expected 3, received %d", p.Available())
	}

	println("end")
}

// read defautl file content
func readFile() (string, []byte) {
	name := "client.proto"
	content, err := ioutil.ReadFile(name)
	if err != nil {
		content = strconv.AppendQuote(content, err.Error())
	}
	return fmt.Sprintf("%s_%s", uuid.NewV4().String(), name), content
}

func covertStatus(err error) {
	s := status.Convert(err)
	log.Printf("status detail info length is %d", len(s.Details()))
	for _, d := range s.Details() {
		switch info := d.(type) {
		default:
			log.Print(info)
		}
	}
	os.Exit(1)
}
