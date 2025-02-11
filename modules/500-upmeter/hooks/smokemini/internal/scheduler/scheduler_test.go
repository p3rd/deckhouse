/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/deckhouse/deckhouse/go_lib/set"
	snapshot "github.com/deckhouse/deckhouse/modules/500-upmeter/hooks/smokemini/internal/snapshot"
)

func Test_scheduler_cleaning(t *testing.T) {
	storageClass := "default"
	image := "smoke-mini"
	zone := "A"

	nodesInOneZone := []snapshot.Node{
		fakeNode(1, zone),
		fakeNode(2, zone),
		fakeNode(3, zone),
		fakeNode(4, zone),
		fakeNode(5, zone),
	}
	pods := fakePods(5)

	type fields struct {
		indexSelector IndexSelector
		nodeFilter    NodeFilter
		image         string
		storageClass  string

		pods []snapshot.Pod
	}
	type args struct {
		state State
		nodes []snapshot.Node
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		asserter deletedResourceAssertion
	}{
		{
			name: "deletes nothing, if nothing changes and pod is ok",
			fields: fields{
				indexSelector: &fakeIndexSelector{"b"},
				nodeFilter:    &noopNodeFilter{},
				pods:          pods,
				image:         image,
				storageClass:  storageClass,
			},
			args: args{
				state: fakeStateInSingleZone(zone),
				nodes: nodesInOneZone,
			},
			asserter: deletedResourceAssertion{},
		},
		{
			name: "deletes pvc, sts, and pod if storage class changed",
			fields: fields{
				indexSelector: &fakeIndexSelector{"a"},
				nodeFilter:    &noopNodeFilter{},
				pods:          pods,
				image:         image,
				storageClass:  storageClass + "_new", // changed
			},
			args: args{
				state: fakeStateInSingleZone(zone),
				nodes: nodesInOneZone,
			},
			asserter: deletedResourceAssertion{x: "a", pod: true, sts: true, pvc: true},
		},
		{
			name: "deletes pvc and pod if zone changed",
			fields: fields{
				indexSelector: &fakeIndexSelector{"c"},
				nodeFilter: &mockNodeFilter{nodes: []snapshot.Node{
					fakeNode(3, "ZZZONE"),
				}},
				pods:         pods,
				image:        image,
				storageClass: storageClass,
			},
			args: args{
				state: fakeStateInSingleZone(zone),
				nodes: []snapshot.Node{
					fakeNode(1), fakeNode(2),
					fakeNode(3, "ZZZONE"),
					fakeNode(4), fakeNode(5),
				},
			},
			asserter: deletedResourceAssertion{x: "c", pod: true, pvc: true},
		},
		{
			name: "deletes pod if it is not running",
			fields: fields{
				indexSelector: &fakeIndexSelector{"e"},
				nodeFilter:    &noopNodeFilter{},
				pods: append(fakePods(4), snapshot.Pod{
					Index:   "e",
					Node:    named("node", 5),
					Phase:   v1.PodPending,
					Created: time.Now(),
				}),
				image:        image,
				storageClass: storageClass,
			},
			args: args{
				state: fakeStateInSingleZone(zone),
				nodes: nodesInOneZone,
			},
			asserter: deletedResourceAssertion{x: "e", pod: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kleaner := &kubeCleaner{
				pods:       tt.fields.pods,
				podDeleter: &fakeDeleter{},
				pvcDeleter: &fakeDeleter{},
				stsDeleter: &fakeDeleter{},
			}
			s := &Scheduler{
				indexSelector: tt.fields.indexSelector,
				nodeFilter:    tt.fields.nodeFilter,
				cleaner:       kleaner,
				image:         tt.fields.image,
				storageClass:  tt.fields.storageClass,
			}
			_, _, _ = s.Schedule(tt.args.state, tt.args.nodes)
			tt.asserter.Assert(t, kleaner)
		})
	}
}

type deletedResourceAssertion struct {
	x             string
	pod, sts, pvc bool
}

func (a deletedResourceAssertion) Assert(t *testing.T, cc *kubeCleaner) {
	pvcs := cc.pvcDeleter.(*fakeDeleter).names
	pods := cc.podDeleter.(*fakeDeleter).names
	sts := cc.stsDeleter.(*fakeDeleter).names

	x := snapshot.Index(a.x)

	if a.pod {
		assert.True(t, pods.Has(x.PodName()), "Pod should be deleted")
	} else {
		assert.False(t, pods.Has(x.PodName()), "Pod should not be deleted")
	}

	if a.pvc {
		assert.True(t, pvcs.Has(x.PersistenceVolumeClaimName()), "PVC should be deleted")
	} else {
		assert.False(t, pvcs.Has(x.PersistenceVolumeClaimName()), "PVC should not be deleted")
	}

	if a.sts {
		assert.True(t, sts.Has(x.StatefulSetName()), "StatefulSet should be deleted")
	} else {
		assert.False(t, sts.Has(x.StatefulSetName()), "StatefulSet should not be deleted")
	}
}

type fakeDeleter struct {
	names set.Set
}

func (d *fakeDeleter) Delete(name string) {
	if d.names == nil {
		d.names = set.New()
	}
	d.names.Add(name)
}

type fakeIndexSelector struct {
	index string
}

func (s *fakeIndexSelector) Select(_ State) (string, error) {
	return s.index, nil
}

type fakeIndexErrorSelector struct {
	err error
}

func (s *fakeIndexErrorSelector) Select(_ State) (string, error) {
	return "", s.err
}

type noopNodeFilter struct{}

func (f *noopNodeFilter) Filter(nodes []snapshot.Node, _ string) []snapshot.Node {
	return nodes
}

type mockNodeFilter struct {
	nodes []snapshot.Node
}

func (f *mockNodeFilter) Filter(_ []snapshot.Node, _ string) []snapshot.Node {
	return f.nodes
}
