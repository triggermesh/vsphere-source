// Copyright 2019 TriggerMesh, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package adapter

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

func TestInitClient(t *testing.T) {
	a := Adapter{}
	if err := a.initClient(); err != nil {
		t.Error("failed to create cloudevent client,", err)
	}
}

func TestStart(t *testing.T) {
	a := Adapter{
		SinkURI:         "qwer",
		VSphereURL:      "example.com",
		VSphereUser:     "foo",
		VSpherePassword: "bar",
	}

	ctx := context.Background()
	stopCh := make(chan struct{}, 1)

	stopCh <- struct{}{}
	if err := a.Start(ctx, stopCh); err != nil {
		t.Error("expected Start to return cleanly, but got", err)
	}

}

// func TestPollLoop(t *testing.T) {
// 	a := Adapter{
// 		SinkURI:         "qwer",
// 		VSphereURL:      "example.com",
// 		VSphereUser:     "foo",
// 		VSpherePassword: "bar",
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	client, err := vSphereClient(ctx, a.VSphereURL, a.VSphereUser, a.VSpherePassword)
// 	if err != nil {
// 		cancel()
// 		t.Log(err)
// 		return
// 	}
// 	cancel()

// 	if err := a.pollLoop(ctx, client); err != nil {
// 		t.Error("expected poll loop to return cleanly, but got", err)
// 	}
// }

func TestReceiveMessage_ServeHTTP(t *testing.T) {
	m := &types.Event{
		Key:                  int32(1),
		CreatedTime:          time.Now(),
		FullFormattedMessage: "foo",
	}
	testCases := map[string]struct {
		sink  func(http.ResponseWriter, *http.Request)
		acked bool
	}{
		"happy": {
			sink:  sinkAccepted,
			acked: true,
		},
		"rejected": {
			sink:  sinkRejected,
			acked: false,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			h := &fakeHandler{
				handler: tc.sink,
			}
			sinkServer := httptest.NewServer(h)
			defer sinkServer.Close()

			a := &Adapter{
				VSphereURL: "example.com",
				SinkURI:    sinkServer.URL,
			}

			if err := a.initClient(); err != nil {
				t.Errorf("failed to create cloudevent client, %v", err)
			}

			ack := new(bool)

			a.receiveMessage(context.TODO(), m, func() { *ack = true })

			if tc.acked != *ack {
				t.Error("expected message ack ", tc.acked, " but real is ", *ack)
			}

		})
	}
}

type fakeHandler struct {
	body []byte

	handler func(http.ResponseWriter, *http.Request)
}

func (h *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can not read body", http.StatusBadRequest)
		return
	}
	h.body = body

	defer r.Body.Close()
	h.handler(w, r)
}

func sinkAccepted(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func sinkRejected(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusRequestTimeout)
}
