package oss

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"online_judge/talcity/scaffold/criteria/merr"
)

const (
	// 存储过期时间,存储不支持永久url, 需要设置一个较长的过期时间
	LongTermExpireToSec = 60 * 60 * 24 * 30 * 12 * 50 // 50年
)

// OSSConfig oss sdk config
type OSSConfig struct {
	Addr             string
	Port             int64
	Cap              int
	MaxCap           int
	PublicEndpoint   string
	InternalEndpoint string
	AccessKey        string
	SecretKey        string
	Bucket           string
}

var (
	pool *OOSClientPool
	conf *OSSConfig
)

func Init(c *OSSConfig) {
	p, err := NewOOSClientPool(strings.Join([]string{c.Addr, strconv.FormatInt(c.Port, 10)}, ":"), c.Cap, c.MaxCap, 0)
	if err != nil {
		panic(err)
	}
	pool = p
	conf = c
}

func Get(ctx context.Context) (*OOSClientResource, error) {
	client, err := pool.GetOOSClient(ctx)
	if err != nil {
		return nil, merr.WrapDefaultCode(err)
	}
	return client, nil
}

func Clean(client *OOSClientResource) {
	pool.Put(client)
}

// GetInternalBucket
// use internal endpoint, get oss Bucket
// when PubObject, use.
func GetInternalBucket() *Bucket {
	return getBucket(conf.InternalEndpoint)
}

// GetPublicBucket
// use public endpoint, get oss Bucket
// when GenerateAccessUrl, use.
func GetPublicBucket() *Bucket {
	return getBucket(conf.PublicEndpoint)
}

// getBucket
// get oss Bucket
func getBucket(endpoint string) *Bucket {
	return &Bucket{
		Endpoint:  endpoint,
		AccessKey: conf.AccessKey,
		SecretKey: conf.SecretKey,
		Bucket:    conf.Bucket,
	}
}

// GetGenerateURLReq
// get oss GenerateURLReq struct
// ocean 暂时不支持永久url, 需要设置一个较长的过期时间
func GetGenerateURLReq(objectName string, endpoint string) *GenerateURLReq {
	return &GenerateURLReq{
		Bucket:       getBucket(endpoint),
		ObjectName:   objectName,
		ExpiresToSec: LongTermExpireToSec,
	}
}

// GenerateURL
// generate access url, use public endpoint.
func GenerateAccessUrl(ctx context.Context, objectNames ...string) (accessUrl []string, err error) {
	client, err := Get(ctx)
	if err != nil {
		return nil, merr.WrapDefaultCode(err)
	}

	defer Clean(client)

	accessUrl = make([]string, len(objectNames))
	for i, objectName := range objectNames {
		if objectName == "" {
			accessUrl[i] = objectName
			continue
		}

		if _, err := url.ParseRequestURI(objectName); err == nil {
			accessUrl[i] = objectName
			continue
		}

		resp, err := client.GenerateURL(ctx, GetGenerateURLReq(objectName, conf.PublicEndpoint))
		if err != nil {
			return nil, merr.WrapDefaultCode(err, "GenerateAccessUrl by [%s] from [%s] Failure.", objectName, conf.PublicEndpoint)
		}

		accessUrl[i] = resp.GetURI()
	}

	return
}

// GetPutObjectReq
// get PutObjectReq struct.
func GetPutObjectReq(objectName string, endpoint string, content []byte) *PutObjectReq {
	return &PutObjectReq{
		Bucket:     getBucket(endpoint),
		ObjectName: objectName,
		Body:       content,
	}
}

// PutObject
// put object, use internal endpoint.
func PutObject(ctx context.Context, objectName string, content []byte) error {
	client, err := Get(ctx)
	if err != nil {
		return merr.WrapDefaultCode(err)
	}

	defer Clean(client)

	_, err = client.PutObject(ctx, GetPutObjectReq(objectName, conf.InternalEndpoint, content))
	if err != nil {
		return merr.WrapDefaultCode(err, "PutObject [%s] to %s Failure.", objectName, conf.InternalEndpoint)
	}

	return nil
}
