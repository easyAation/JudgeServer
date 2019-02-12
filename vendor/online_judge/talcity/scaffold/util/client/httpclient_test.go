package client

import (
	"encoding/json"
	"testing"

	"online_judge/talcity/scaffold/util/wait"
)

var (
	addr      = "http://pt-dev.platform.facethink.com/accounts/internal/v1/account/user?uids=ZzLZpR1fTyqo3TRZH6SSsg"
	newClient = map[string]*HttpClient{
		"NewClientManual":   NewClient(addr),
		"BaiduClientManual": NewClient(addr),
	}
)

func TestHttpClient(t *testing.T) {
	for _, name := range []string{"NewClientManual", "BaiduClientManual"} {
		t.Run(name, func(t *testing.T) {
			client := newClient[name]
			t.Run("ClientGet", func(t *testing.T) {
				testClientGet(t, client)
			})
			t.Run("100GetPerMinu", func(t *testing.T) {
				test100GetPerMinu(t, client)
			})
		})
	}
}

func testClientGet(t *testing.T, client *HttpClient) {
	body, err := client.Get("", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := struct {
		Code        int           `json:"code"`
		Msg         string        `json:"msg"`
		Description string        `json:"description"`
		Body        []interface{} `json:"body"`
	}{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", resp)
}

func test100GetPerMinu(t *testing.T, client *HttpClient) {
	t.Parallel()

	var group wait.Group

	group.Start(func() {
		reqCount := 1024
		count := 1
		for count < reqCount {
			testClientGet(t, client)
			count++
		}
	})

	group.Wait()
}

func BenchmarkHttpClient(b *testing.B) {
	for _, name := range []string{"NewClientManual"} {
		b.Run(name, func(b *testing.B) {
			client := newClient[name]
			b.Run("ClientGet", func(b *testing.B) {
				benchmarkClientGet(b, client)
			})
		})
	}
}

func benchmarkClientGet(b *testing.B, client *HttpClient) {
	for i := 0; i < b.N; i++ {
		body, err := client.Get("", nil, nil)
		if err != nil {
			b.Fatal(err)
		}

		resp := struct {
			Code        int           `json:"code"`
			Msg         string        `json:"msg"`
			Description string        `json:"description"`
			Body        []interface{} `json:"body"`
		}{}
		err = json.Unmarshal(body, &resp)
		if err != nil {
			b.Fatal(err)
		}

		b.Logf("%#v", resp)
	}
}
