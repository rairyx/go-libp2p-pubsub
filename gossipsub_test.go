package pubsub

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
//         "sync"
	"github.com/libp2p/go-libp2p-core/host"
)

func getGossipsub(ctx context.Context, h host.Host, opts ...Option) *PubSub {
	ps, err := NewGossipSub(ctx, h, opts...)
	if err != nil {
		panic(err)
	}
	return ps
}

func getGossipsubs(ctx context.Context, hs []host.Host, opts ...Option) []*PubSub {
	var psubs []*PubSub
	for _, h := range hs {
		psubs = append(psubs, getGossipsub(ctx, h, opts...))
	}
	return psubs
}

func TestSparseGossipsub(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	sparseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)
	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
		        got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			//fmt.Println(string(got.Data))
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}

	}
}

func TestDenseGossipsub(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	denseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			//fmt.Println(string(got.Data))
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}
}

func TestGossipsubFanout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs[1:] {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	denseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := 0

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
//			fmt.Println(string(got.Data))
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}

	// subscribe the owner
	subch, err := psubs[0].Subscribe("foobar")
	if err != nil {
		t.Fatal(err)
	}
	msgs = append(msgs, subch)

	// wait for a heartbeat
	time.Sleep(time.Second * 2)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := 0

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
	//		fmt.Println(string(got.Data))
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}

}

func TestGossipsubFanoutMaintenance(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts,WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs[1:] {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	denseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := 0

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}

	// unsubscribe all peers to exercise fanout maintenance
	for _, sub := range msgs {
		sub.Cancel()
	}
	msgs = nil

	// wait for heartbeats
	time.Sleep(time.Second * 2)

	// resubscribe and repeat
	for _, ps := range psubs[1:] {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	time.Sleep(time.Second * 2)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := 0

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}
}

func TestGossipsubFanoutExpiry(t *testing.T) {
	GossipSubFanoutTTL = 1 * time.Second
	defer func() { GossipSubFanoutTTL = 60 * time.Second }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 10)

	psubs := getGossipsubs(ctx, hosts,WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs[1:] {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	denseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	for i := 0; i < 5; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := 0

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}

	psubs[0].eval <- func() {
		if len(psubs[0].rt.(*GossipSubRouter).fanout) == 0 {
			t.Fatal("owner has no fanout")
		}
	}

	// wait for TTL to expire fanout peers in owner
	time.Sleep(time.Second * 2)

	psubs[0].eval <- func() {
		if len(psubs[0].rt.(*GossipSubRouter).fanout) > 0 {
			t.Fatal("fanout hasn't expired")
		}
	}

	// wait for it to run in the event loop
	time.Sleep(10 * time.Millisecond)
}

func TestGossipsubGossip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	denseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}

		// wait a bit to have some gossip interleaved
		time.Sleep(time.Millisecond * 100)
	}

	// and wait for some gossip flushing
	
	time.Sleep(time.Second * 2)
}

func TestGossipsubGossipPiggyback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	var xmsgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("bazcrux")
		if err != nil {
			t.Fatal(err)
		}

		xmsgs = append(xmsgs, subch)
	}

	denseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish("foobar", msg)
		psubs[owner].Publish("bazcrux", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}

		for _, sub := range xmsgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}

		// wait a bit to have some gossip interleaved
		time.Sleep(time.Millisecond * 100)
	}

	// and wait for some gossip flushing
	time.Sleep(time.Second * 2)
}

func TestGossipsubGossipPropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := getNetHosts(t, ctx, 20)
	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	hosts1 := hosts[:GossipSubD+1]
	hosts2 := append(hosts[GossipSubD+1:], hosts[0])

	denseConnect(t, hosts1)
	denseConnect(t, hosts2)

	var msgs1 []*Subscription
	for _, ps := range psubs[1 : GossipSubD+1] {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs1 = append(msgs1, subch)
	}

	time.Sleep(time.Second * 1)

	for i := 0; i < 10; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := 0

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs1 {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}

	time.Sleep(time.Millisecond * 100)

	var msgs2 []*Subscription
	for _, ps := range psubs[GossipSubD+1:] {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs2 = append(msgs2, subch)
	}

	var collect [][]byte
	for i := 0; i < 10; i++ {
		for _, sub := range msgs2 {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			collect = append(collect, got.Data)
		}
	}

	for i := 0; i < 10; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))
		gotit := false
		for j := 0; j < len(collect); j++ {
			if bytes.Equal(msg, collect[j]) {
				gotit = true
				break
			}
		}
		if !gotit {
			t.Fatalf("Didn't get message %s", string(msg))
		}
	}
}

func TestGossipsubPrune(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts,WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	denseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	// disconnect some peers from the mesh to get some PRUNEs
	for _, sub := range msgs[:5] {
		sub.Cancel()
	}

	// wait a bit to take effect
	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 10; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs[5:] {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}
}

func TestGossipsubGraft(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	sparseConnect(t, hosts)

	time.Sleep(time.Second * 1)

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)

		// wait for announce to propagate
		time.Sleep(time.Millisecond * 100)
	}

	time.Sleep(time.Second * 1)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}
}

func TestGossipsubRemovePeer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 20)

	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	denseConnect(t, hosts)

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	// disconnect some peers to exercise RemovePeer paths
	for _, host := range hosts[:5] {
		host.Close()
	}

	// wait a heartbeat
	time.Sleep(time.Second * 1)

	for i := 0; i < 10; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := 5 + rand.Intn(len(psubs)-5)

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs[5:] {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}
}

func TestGossipsubGraftPruneRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := getNetHosts(t, ctx, 10)
	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))
	denseConnect(t, hosts)

	var topics []string
	var msgs [][]*Subscription
	for i := 0; i < 35; i++ {
		topic := fmt.Sprintf("topic%d", i)
		topics = append(topics, topic)

		var subs []*Subscription
		for _, ps := range psubs {
			subch, err := ps.Subscribe(topic)
			if err != nil {
				t.Fatal(err)
			}

			subs = append(subs, subch)
		}
		msgs = append(msgs, subs)
	}

	// wait for heartbeats to build meshes
	time.Sleep(time.Second * 5)

	for i, topic := range topics {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish(topic, msg)

		for _, sub := range msgs[i] {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}
}

func TestGossipsubControlPiggyback(t *testing.T) {
	t.Skip("travis regularly fails on this test")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := getNetHosts(t, ctx, 10)
	psubs := getGossipsubs(ctx, hosts,WithMessageSigning(false))
	denseConnect(t, hosts)

	for _, ps := range psubs {
		subch, err := ps.Subscribe("flood")
		if err != nil {
			t.Fatal(err)
		}
		go func(sub *Subscription) {
			for {
				_, err := sub.Next(ctx)
				if err != nil {
					break
				}
			}
		}(subch)
	}

	time.Sleep(time.Second * 1)

	// create a background flood of messages that overloads the queues
	done := make(chan struct{})
	go func() {
		owner := rand.Intn(len(psubs))
		for i := 0; i < 10000; i++ {
			msg := []byte("background flooooood")
			psubs[owner].Publish("flood", msg)
		}
		done <- struct{}{}
	}()

	time.Sleep(time.Millisecond * 20)

	// and subscribe to a bunch of topics in the meantime -- this should
	// result in some dropped control messages, with subsequent piggybacking
	// in the background flood
	var topics []string
	var msgs [][]*Subscription
	for i := 0; i < 5; i++ {
		topic := fmt.Sprintf("topic%d", i)
		topics = append(topics, topic)

		var subs []*Subscription
		for _, ps := range psubs {
			subch, err := ps.Subscribe(topic)
			if err != nil {
				t.Fatal(err)
			}

			subs = append(subs, subch)
		}
		msgs = append(msgs, subs)
	}

	// wait for the flood to stop
	<-done

	// and test that we have functional overlays
	for i, topic := range topics {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish(topic, msg)

		for _, sub := range msgs[i] {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}
}

func TestMixedGossipsub(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hosts := getNetHosts(t, ctx, 30)

	gsubs := getGossipsubs(ctx, hosts[:20],WithMessageSigning(false))
	fsubs := getPubsubs(ctx, hosts[20:],WithMessageSigning(false))
	psubs := append(gsubs, fsubs...)

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	sparseConnect(t, hosts)
	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	for i := 0; i < 100; i++ {
		msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))

		owner := rand.Intn(len(psubs))

		psubs[owner].Publish("foobar", msg)

		for _, sub := range msgs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
		}
	}
}

func TestGossipsubMultihops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := getNetHosts(t, ctx, 6)

	psubs := getGossipsubs(ctx, hosts,WithMessageSigning(false))

	connect(t, hosts[0], hosts[1])
	connect(t, hosts[1], hosts[2])
	connect(t, hosts[2], hosts[3])
	connect(t, hosts[3], hosts[4])
	connect(t, hosts[4], hosts[5])

	var subs []*Subscription
	for i := 1; i < 6; i++ {
		ch, err := psubs[i].Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}
		subs = append(subs, ch)
	}

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	msg := []byte("i like cats")
	err := psubs[0].Publish("foobar", msg)
	if err != nil {
		t.Fatal(err)
	}

	// last node in the chain should get the message
	select {
	case out := <-subs[4].ch:
		if !bytes.Equal(out.GetData(), msg) {
			t.Fatal("got wrong data")
		}
	case <-time.After(time.Second * 5):
		t.Fatal("timed out waiting for message")
	}
}

func TestGossipsubTreeTopology(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := getNetHosts(t, ctx, 10)
	psubs := getGossipsubs(ctx, hosts, WithMessageSigning(false))

	connect(t, hosts[0], hosts[1])
	connect(t, hosts[1], hosts[2])
	connect(t, hosts[1], hosts[4])
	connect(t, hosts[2], hosts[3])
	connect(t, hosts[0], hosts[5])
	connect(t, hosts[5], hosts[6])
	connect(t, hosts[5], hosts[8])
	connect(t, hosts[6], hosts[7])
	connect(t, hosts[8], hosts[9])

	/*
		[0] -> [1] -> [2] -> [3]
		 |      L->[4]
		 v
		[5] -> [6] -> [7]
		 |
		 v
		[8] -> [9]
	*/

	var chs []*Subscription
	for _, ps := range psubs {
		ch, err := ps.Subscribe("fizzbuzz")
		if err != nil {
			t.Fatal(err)
		}

		chs = append(chs, ch)
	}

	// wait for heartbeats to build mesh
	time.Sleep(time.Second * 2)

	assertPeerLists(t, hosts, psubs[0], 1, 5)
	assertPeerLists(t, hosts, psubs[1], 0, 2, 4)
	assertPeerLists(t, hosts, psubs[2], 1, 3)

	checkMessageRouting(t, "fizzbuzz", []*PubSub{psubs[9], psubs[3]}, chs)
}

func TestDandelionStemPhase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := getNetHosts(t, ctx, 20)
	psubs := getGossipsubs(ctx, hosts,WithMessageSigning(false))
	denseConnect(t, hosts)

	var msgs []*Subscription
	for _, ps := range psubs {
		subch, err := ps.Subscribe("foobar")
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, subch)
	}

	time.Sleep(time.Second * 10)


//	msg := []byte(fmt.Sprintf("%d it's not a floooooood %d", i, i))
	owner := rand.Intn(len(psubs))
	
	for psubs[owner].phase == fluff {
		owner = rand.Intn(len(psubs))
	}

	stemPeer:= psubs[owner].stemPeer
	inext:= owner
	if stemPeer == ""{ 
		t.Fatal("stem peer is empty!")
	}
	//	var nextHost host.Host
	var stemPeerSubs[]*Subscription
	var stemSubs[]*PubSub
	stemPeers:=make(map[string]struct{})
//	fmt.Printf("phase %d\n",psubs[inext].phase)
	for; psubs[inext].phase == stem; {
		stemPeerSubs =append(stemPeerSubs, msgs[inext])
		stemSubs = append(stemSubs, psubs[inext])
		stemPeers[string(psubs[inext].host.ID())] = struct{}{} 
		for i, h := range hosts {
			if h.ID() == psubs[inext].stemPeer{
				inext = i
				fmt.Printf("inext= %d\n",i)
				break
			}
		}
//		fmt.Printf("stem peer %s\n", stemPeer)
		stemPeer = psubs[inext].stemPeer
		_, ok := stemPeers[string(stemPeer)]
		if ok{
			break
		}
	}
//close down last stem peer
	stemSubs[len(stemSubs)-1].host.Close()
	
//	time.Sleep(time.Second * 1)
//	fmt.Println(len(stemPeerSubs))

	msg := []byte("i like dogs")
	psubs[owner].Publish("foobar", msg)
//	stemSubs[len(stemSubs)-1].host.Close()
//	hosts[inext].Close()
//	time.Sleep(time.Second * 5)
//check for stem message cache
	for _, sub := range stemSubs {
		fmt.Printf("stem sub peer %s\n",sub.host.ID())
		for _, smsg := range sub.stemMsgs{
			fmt.Printf("stem messge=%s\n", string(smsg.msg.Data))
		}
	
	}
	for i, sub := range stemPeerSubs {
			got, err := sub.Next(ctx)
			if err != nil {
				t.Fatal(sub.err)
			}
			fmt.Printf("peer %s\n",stemSubs[i].host.ID()) 
			fmt.Println(string(got.Data))
			fmt.Printf("state= %s\n", *got.State)
			//if !bytes.Equal(msg, got.Data) || *got.State!= "STEM" {
			if !bytes.Equal(msg, got.Data) {
				t.Fatal("got wrong message!")
			}
	}
}
