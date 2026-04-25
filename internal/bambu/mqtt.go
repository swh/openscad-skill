package bambu

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ReadAMSLive opens an MQTT connection to a Bambu printer over the LAN,
// requests a full status push, and returns the AMS state seen in the first
// message that contains it. Tracks an AMSState shape (single-printer map)
// for parity with ReadAMSCache.
//
// Bambu LAN MQTT particulars: TLS on 8883 with a self-signed cert (we
// disable verification), username "bblp", password = LAN access code.
//
// timeout caps total time spent on connect + subscribe + first message.
func ReadAMSLive(host, serial, accessCode string, timeout time.Duration) (AMSState, error) {
	if host == "" || serial == "" || accessCode == "" {
		return nil, fmt.Errorf("MQTT requires host, serial and access code")
	}
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("ssl://%s:8883", host))
	opts.SetClientID(fmt.Sprintf("openscad-pack-3mf-%d", time.Now().UnixNano()))
	opts.SetUsername("bblp")
	opts.SetPassword(accessCode)
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	opts.SetAutoReconnect(false)
	opts.SetConnectTimeout(timeout)
	opts.SetWriteTimeout(timeout)

	// Channel receives the parsed slots (or an error) from the message handler.
	type result struct {
		slots []string
		err   error
	}
	resCh := make(chan result, 1)
	var sentOnce sync.Once
	send := func(r result) { sentOnce.Do(func() { resCh <- r }) }

	opts.SetDefaultPublishHandler(func(_ mqtt.Client, m mqtt.Message) {
		slots, ok, err := parseAMSPayload(m.Payload())
		if err != nil {
			send(result{err: err})
			return
		}
		if ok {
			send(result{slots: slots})
		}
	})

	c := mqtt.NewClient(opts)
	tok := c.Connect()
	if !tok.WaitTimeout(timeout) {
		return nil, fmt.Errorf("MQTT connect to %s timed out", host)
	}
	if err := tok.Error(); err != nil {
		return nil, fmt.Errorf("MQTT connect to %s: %w", host, err)
	}
	defer c.Disconnect(100)

	subTopic := fmt.Sprintf("device/%s/report", serial)
	if tok := c.Subscribe(subTopic, 0, nil); !tok.WaitTimeout(timeout) {
		return nil, fmt.Errorf("MQTT subscribe %s timed out", subTopic)
	} else if err := tok.Error(); err != nil {
		return nil, fmt.Errorf("MQTT subscribe %s: %w", subTopic, err)
	}

	// Ask the printer to send a full snapshot (sometimes it volunteers one
	// on subscribe; sometimes it doesn't).
	pubTopic := fmt.Sprintf("device/%s/request", serial)
	pushPayload := []byte(`{"pushing":{"sequence_id":"0","command":"pushall"}}`)
	if tok := c.Publish(pubTopic, 0, false, pushPayload); !tok.WaitTimeout(timeout) {
		return nil, fmt.Errorf("MQTT publish %s timed out", pubTopic)
	} else if err := tok.Error(); err != nil {
		return nil, fmt.Errorf("MQTT publish %s: %w", pubTopic, err)
	}

	select {
	case r := <-resCh:
		if r.err != nil {
			return nil, r.err
		}
		return AMSState{serial: r.slots}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("MQTT timeout waiting for AMS state from %s", host)
	}
}

// parseAMSPayload digs through Bambu's report JSON for `print.ams.ams[]`,
// flattening per-AMS tray lists into a single ordered slot list of
// `tray_info_idx` strings. Returns ok=false (no error) when the payload
// doesn't carry AMS state, so the message handler can wait for one that
// does.
func parseAMSPayload(payload []byte) (slots []string, ok bool, err error) {
	var top map[string]any
	if err := json.Unmarshal(payload, &top); err != nil {
		// Non-JSON or partial payload — ignore.
		return nil, false, nil
	}
	pr, _ := top["print"].(map[string]any)
	if pr == nil {
		return nil, false, nil
	}
	amsWrap, _ := pr["ams"].(map[string]any)
	if amsWrap == nil {
		return nil, false, nil
	}
	amsList, _ := amsWrap["ams"].([]any)
	if amsList == nil {
		return nil, false, nil
	}
	var out []string
	found := false
	for _, a := range amsList {
		am, ok := a.(map[string]any)
		if !ok {
			continue
		}
		trays, _ := am["tray"].([]any)
		for _, tv := range trays {
			t, ok := tv.(map[string]any)
			if !ok {
				continue
			}
			id, _ := t["tray_info_idx"].(string)
			out = append(out, strings.TrimSpace(id))
			found = true
		}
	}
	return out, found, nil
}
